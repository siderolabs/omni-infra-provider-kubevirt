// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package provider implements KubeVirt infra provider core.
package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"time"

	"github.com/google/uuid"
	pointer "github.com/siderolabs/go-pointer"
	"github.com/siderolabs/omni/client/pkg/constants"
	"github.com/siderolabs/omni/client/pkg/infra/provision"
	"github.com/siderolabs/omni/client/pkg/omni/resources/infra"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kvv1 "kubevirt.io/api/core/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/siderolabs/omni-infra-provider-kubevirt/internal/pkg/provider/data"
	"github.com/siderolabs/omni-infra-provider-kubevirt/internal/pkg/provider/resources"
)

// Provisioner implements Talos emulator infra provider.
type Provisioner struct {
	k8sClient  client.Client
	namespace  string
	volumeMode v1.PersistentVolumeMode
}

// NewProvisioner creates a new provisioner.
func NewProvisioner(k8sClient client.Client, namespace, volumeMode string) *Provisioner {
	return &Provisioner{
		k8sClient:  k8sClient,
		namespace:  namespace,
		volumeMode: v1.PersistentVolumeMode(volumeMode),
	}
}

// ProvisionSteps implements infra.Provisioner.
//
//nolint:gocognit,gocyclo,cyclop,maintidx
func (p *Provisioner) ProvisionSteps() []provision.Step[*resources.Machine] {
	return []provision.Step[*resources.Machine]{
		provision.NewStep("validateRequest", func(_ context.Context, _ *zap.Logger, pctx provision.Context[*resources.Machine]) error {
			if len(pctx.GetRequestID()) > 62 {
				return fmt.Errorf("the machine request name can not be longer than 63 characters")
			}

			return nil
		}),
		provision.NewStep("createSchematic", func(ctx context.Context, logger *zap.Logger, pctx provision.Context[*resources.Machine]) error {
			schematic, err := pctx.GenerateSchematicID(ctx, logger,
				provision.WithExtraKernelArgs("console=ttyS0,38400n8"),
				provision.WithoutConnectionParams(),
			)
			if err != nil {
				return err
			}

			pctx.State.TypedSpec().Value.Schematic = schematic

			return nil
		}),
		provision.NewStep("ensureVolume", func(ctx context.Context, _ *zap.Logger, pctx provision.Context[*resources.Machine]) error {
			pctx.State.TypedSpec().Value.TalosVersion = pctx.GetTalosVersion()

			url, err := url.Parse(constants.ImageFactoryBaseURL)
			if err != nil {
				return err
			}

			var data data.Data

			err = pctx.UnmarshalProviderData(&data)
			if err != nil {
				return err
			}

			url = url.JoinPath("image",
				pctx.State.TypedSpec().Value.Schematic,
				pctx.GetTalosVersion(),
				fmt.Sprintf("nocloud-%s.qcow2", data.Architecture),
			)

			hash := sha256.New()

			if _, err = hash.Write([]byte(url.String())); err != nil {
				return err
			}

			volumeID := hex.EncodeToString(hash.Sum(nil))

			pctx.State.TypedSpec().Value.VolumeId = volumeID

			volume := cdiv1.DataVolume{
				Spec: cdiv1.DataVolumeSpec{
					Source: &cdiv1.DataVolumeSource{
						HTTP: &cdiv1.DataVolumeSourceHTTP{
							URL: url.String(),
						},
					},
					PVC: &v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						Resources: v1.VolumeResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("5Gi"),
							},
						},
					},
				},
			}

			if data.StorageClassName != "" {
				volume.Spec.PVC.StorageClassName = &data.StorageClassName
			}

			if p.volumeMode != "" {
				volume.Spec.PVC.VolumeMode = &p.volumeMode
			}

			if volume.Annotations == nil {
				volume.Annotations = map[string]string{}
			}

			volume.Annotations["cdi.kubevirt.io/storage.bind.immediate.requested"] = "true"

			vol := &cdiv1.DataVolume{}

			err = p.k8sClient.Get(ctx, client.ObjectKey{
				Namespace: p.namespace,
				Name:      volumeID,
			}, vol)
			if err != nil && !errors.IsNotFound(err) {
				return err
			}

			if vol.Status.Phase == cdiv1.Succeeded {
				return nil
			}

			if vol.Name == "" {
				volume.Name = volumeID
				volume.Namespace = p.namespace

				if err = p.k8sClient.Create(ctx, &volume); err != nil {
					return err
				}
			}

			return provision.NewRetryInterval(time.Second * 10)
		}),
		provision.NewStep("syncMachine", func(ctx context.Context, logger *zap.Logger, pctx provision.Context[*resources.Machine]) error {
			if pctx.State.TypedSpec().Value.Uuid == "" {
				pctx.State.TypedSpec().Value.Uuid = uuid.NewString()
			}

			logger = logger.With(zap.String("id", pctx.State.TypedSpec().Value.Uuid))

			vm := &kvv1.VirtualMachine{}

			err := p.k8sClient.Get(ctx, client.ObjectKey{
				Namespace: p.namespace,
				Name:      pctx.GetRequestID(),
			}, vm)
			if err != nil && !errors.IsNotFound(err) {
				return err
			}

			if vm.Name != "" && vm.Status.Ready {
				logger.Info("machine is ready")

				return nil
			}

			var data data.Data

			err = pctx.UnmarshalProviderData(&data)
			if err != nil {
				return err
			}

			vm.Spec.Running = pointer.To(true)

			if vm.Spec.Template == nil {
				vm.Spec.Template = &kvv1.VirtualMachineInstanceTemplateSpec{
					Spec: kvv1.VirtualMachineInstanceSpec{
						Domain: kvv1.DomainSpec{
							Resources: kvv1.ResourceRequirements{
								Requests: v1.ResourceList{},
							},
						},
					},
				}
			}

			vm.Spec.Template.Spec.Domain.Firmware = &kvv1.Firmware{
				UUID: types.UID(pctx.State.TypedSpec().Value.Uuid),
			}

			vm.Spec.Template.Spec.Architecture = data.Architecture
			vm.Spec.Template.Spec.Domain.CPU = &kvv1.CPU{
				Cores: uint32(data.Cores),
			}

			vm.Spec.Template.Spec.Domain.Resources.Requests[v1.ResourceMemory] = *resource.NewQuantity(int64(data.Memory)*1024*1024, resource.DecimalSI)

			vm.Spec.Template.Spec.Networks = []kvv1.Network{
				*kvv1.DefaultPodNetwork(),
			}

			networkInterface := *kvv1.DefaultBridgeNetworkInterface()
			if data.NetworkBinding == "passt" {
				networkInterface = kvv1.Interface{
					Name: networkInterface.Name,
					Binding: &kvv1.PluginBinding{
						Name: "passt",
					},
				}
			}

			vm.Spec.Template.Spec.Domain.Devices = kvv1.Devices{
				Disks: []kvv1.Disk{
					{
						Name:      "kv",
						BootOrder: pointer.To(uint(1)),
						DiskDevice: kvv1.DiskDevice{
							Disk: &kvv1.DiskTarget{
								Bus: kvv1.DiskBusVirtio,
							},
						},
					},
				},
				Interfaces: []kvv1.Interface{
					networkInterface,
				},
			}

			vm.Spec.Template.Spec.Volumes = []kvv1.Volume{
				{
					Name: "kv",
					VolumeSource: kvv1.VolumeSource{
						DataVolume: &kvv1.DataVolumeSource{
							Name: pctx.GetRequestID(),
						},
					},
				},
				{
					Name: "cloudinitdisk",
					VolumeSource: kvv1.VolumeSource{
						CloudInitNoCloud: &kvv1.CloudInitNoCloudSource{
							UserData:    pctx.ConnectionParams.JoinConfig,
							NetworkData: `version: 1`,
						},
					},
				},
			}

			if data.Tolerations != "" {
				var tolerations []v1.Toleration

				err = json.Unmarshal([]byte(data.Tolerations), &tolerations)
				if err != nil {
					return err
				}

				vm.Spec.Template.Spec.Tolerations = tolerations
			}

			volumeTemplate := kvv1.DataVolumeTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: pctx.GetRequestID(),
				},
				Spec: cdiv1.DataVolumeSpec{
					PVC: &v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						Resources: v1.VolumeResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse(fmt.Sprintf("%dGi", data.DiskSize)),
							},
						},
					},
					Source: &cdiv1.DataVolumeSource{
						PVC: &cdiv1.DataVolumeSourcePVC{
							Name:      pctx.State.TypedSpec().Value.VolumeId,
							Namespace: p.namespace,
						},
					},
				},
			}

			if data.StorageClassName != "" {
				volumeTemplate.Spec.PVC.StorageClassName = &data.StorageClassName
			}

			if p.volumeMode != "" {
				volumeTemplate.Spec.PVC.VolumeMode = &p.volumeMode
			}

			vm.Spec.DataVolumeTemplates = []kvv1.DataVolumeTemplateSpec{
				volumeTemplate,
			}

			// Apply user-provided labels to the launcher pod
			if len(data.VMLabels) > 0 {
				if vm.Spec.Template.ObjectMeta.Labels == nil {
					vm.Spec.Template.ObjectMeta.Labels = map[string]string{}
				}

				maps.Copy(vm.Spec.Template.ObjectMeta.Labels, data.VMLabels)
			}

			if vm.Name == "" {
				vm.Name = pctx.GetRequestID()
				vm.Namespace = p.namespace

				if err = p.k8sClient.Create(ctx, vm); err != nil {
					return err
				}
			} else {
				if err = p.k8sClient.Update(ctx, vm); err != nil && !errors.IsConflict(err) {
					return err
				}
			}

			return provision.NewRetryInterval(time.Second * 10)
		}),
	}
}

// Deprovision implements infra.Provisioner.
func (p *Provisioner) Deprovision(ctx context.Context, logger *zap.Logger, _ *resources.Machine, machineRequest *infra.MachineRequest) error {
	var vm kvv1.VirtualMachine

	err := p.k8sClient.Get(ctx, client.ObjectKey{
		Namespace: p.namespace,
		Name:      machineRequest.Metadata().ID(),
	}, &vm)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if vm.Name == "" {
		logger.Info("machine deprovisioned")

		return nil
	}

	err = p.k8sClient.Delete(ctx, &kvv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      machineRequest.Metadata().ID(),
			Namespace: p.namespace,
		},
	})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return provision.NewRetryInterval(time.Second * 5)
}

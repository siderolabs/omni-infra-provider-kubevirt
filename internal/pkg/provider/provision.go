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
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/siderolabs/omni/client/pkg/imagefactory"
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

// ensureFactorySecret creates or updates the secret CDI reads the image factory headers from, and
// returns its name.
//
// It is keyed by the factory host rather than by machine so that rotated credentials reach every
// DataVolume that already references it, including the ones whose import has not started yet.
func (p *Provisioner) ensureFactorySecret(ctx context.Context, factoryHost string, headers http.Header) (string, error) {
	digest := sha256.Sum256([]byte(factoryHost))
	name := "omni-image-factory-" + hex.EncodeToString(digest[:])[:16]

	// CDI reads each value of a secretExtraHeaders secret as one complete header line.
	data := make(map[string][]byte, len(headers))

	for header, values := range headers {
		for i, value := range values {
			data[fmt.Sprintf("%s-%d", strings.ToLower(header), i)] = []byte(header + ": " + value)
		}
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: p.namespace,
		},
		Data: data,
	}

	err := p.k8sClient.Create(ctx, secret)
	if err == nil {
		return name, nil
	}

	if !errors.IsAlreadyExists(err) {
		return "", fmt.Errorf("failed to create the image factory secret: %w", err)
	}

	if err = p.k8sClient.Update(ctx, secret); err != nil {
		return "", fmt.Errorf("failed to update the image factory secret: %w", err)
	}

	return name, nil
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
		provision.NewStep("ensureVolume", func(ctx context.Context, logger *zap.Logger, pctx provision.Context[*resources.Machine]) error {
			pctx.State.TypedSpec().Value.TalosVersion = pctx.GetTalosVersion()

			var data data.Data

			err := pctx.UnmarshalProviderData(&data)
			if err != nil {
				return err
			}

			media, err := pctx.EnsureInstallationMedia(
				ctx, logger, provision.MediaSpec{
					MediaSpec: imagefactory.MediaSpec{
						Kind:         imagefactory.InstallationMediaKindDisk,
						Platform:     "nocloud",
						Architecture: data.Architecture,
						Format:       "qcow2",
					},
					DownloadTokenTTL: time.Hour,
				},
				provision.WithExtraKernelArgs("console=ttyS0,38400n8"),
				provision.WithoutConnectionParams(),
			)
			if err != nil {
				return provision.NewRetryErrorf(time.Second*10, "error resolving the boot asset: %w", err)
			}

			pctx.State.TypedSpec().Value.Schematic = media.SchematicID

			// Omni derives the storage key from what decides the image's content, so the volume keeps its
			// name across a credential rotation. The URL cannot serve as the name, since it may carry
			// credentials and would rename the volume whenever they change.
			volumeID := media.StorageKey

			pctx.State.TypedSpec().Value.VolumeId = volumeID

			source := &cdiv1.DataVolumeSourceHTTP{URL: media.URL}

			if len(media.Headers) > 0 {
				secretName, secretErr := p.ensureFactorySecret(ctx, media.ImageFactoryHost, media.Headers)
				if secretErr != nil {
					return secretErr
				}

				source.SecretExtraHeaders = []string{secretName}
			}

			volume := cdiv1.DataVolume{
				Spec: cdiv1.DataVolumeSpec{
					Source: &cdiv1.DataVolumeSource{
						HTTP: source,
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

			vm.Spec.Running = new(true)

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
						BootOrder: new(uint(1)),
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

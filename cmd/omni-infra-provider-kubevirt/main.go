// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package main is the root cmd of the provider script.
package main

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"github.com/siderolabs/omni/client/pkg/client"
	"github.com/siderolabs/omni/client/pkg/infra"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	kvv1 "kubevirt.io/api/core/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/siderolabs/omni-infra-provider-kubevirt/internal/pkg/provider"
	"github.com/siderolabs/omni-infra-provider-kubevirt/internal/pkg/provider/meta"
)

//go:embed data/schema.json
var schema string

//go:embed data/icon.svg
var icon []byte

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:          "provider",
	Short:        "KubeVirt Omni infrastructure provider",
	Long:         `Connects to Omni as an infra provider and manages VMs in KubeVirt`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		loggerConfig := zap.NewProductionConfig()

		logger, err := loggerConfig.Build(
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
		if err != nil {
			return fmt.Errorf("failed to create logger: %w", err)
		}

		scheme := runtime.NewScheme()

		err = kvv1.AddToScheme(scheme)
		if err != nil {
			return err
		}

		err = cdiv1.AddToScheme(scheme)
		if err != nil {
			return err
		}

		config, err := clientcmd.BuildConfigFromFlags("", cfg.kubeconfigFile)
		if err != nil {
			return fmt.Errorf("failed to read Kubernetes config: %w", err)
		}

		k8sClient, err := k8sclient.New(config, k8sclient.Options{
			Scheme: scheme,
		})
		if err != nil {
			return err
		}

		if cfg.omniAPIEndpoint == "" {
			return fmt.Errorf("omni-api-endpoint flag is not set")
		}

		volumeOpts := []v1.PersistentVolumeMode{
			v1.PersistentVolumeBlock,
			v1.PersistentVolumeFilesystem,
		}

		if cfg.dataVolumeMode != "" && slices.Index(volumeOpts, v1.PersistentVolumeMode(cfg.dataVolumeMode)) == -1 {
			return fmt.Errorf("data-volume-mode flags should be one of %s", volumeOpts)
		}

		patches := make([]provider.ConfigPatch, 0, len(cfg.configPatches))

		for _, patch := range cfg.configPatches {
			data := []byte(patch)

			if strings.HasPrefix(patch, "@") {
				data, err = os.ReadFile(strings.TrimPrefix(patch, "@"))
				if err != nil {
					return err
				}
			}

			var p provider.ConfigPatch

			if err = yaml.Unmarshal(data, &p); err != nil {
				return err
			}

			patches = append(patches, p)
		}

		provisioner := provider.NewProvisioner(k8sClient, cfg.namespace, cfg.dataVolumeMode, patches)

		ip, err := infra.NewProvider(meta.ProviderID, provisioner, infra.ProviderConfig{
			Name:        cfg.providerName,
			Description: cfg.providerDescription,
			Icon:        base64.RawStdEncoding.EncodeToString(icon),
			Schema:      schema,
		})
		if err != nil {
			return fmt.Errorf("failed to create infra provider: %w", err)
		}

		logger.Info("starting infra provider")

		clientOptions := []client.Option{
			client.WithInsecureSkipTLSVerify(cfg.insecureSkipVerify),
		}

		if cfg.serviceAccountKey != "" {
			clientOptions = append(clientOptions, client.WithServiceAccount(cfg.serviceAccountKey))
		}

		return ip.Run(cmd.Context(), logger, infra.WithOmniEndpoint(cfg.omniAPIEndpoint), infra.WithClientOptions(
			clientOptions...,
		))
	},
}

var cfg struct {
	omniAPIEndpoint     string
	serviceAccountKey   string
	providerName        string
	providerDescription string
	kubeconfigFile      string
	namespace           string
	dataVolumeMode      string
	configPatches       []string
	insecureSkipVerify  bool
}

func main() {
	if err := app(); err != nil {
		os.Exit(1)
	}
}

func app() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)
	defer cancel()

	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.Flags().StringVar(&cfg.omniAPIEndpoint, "omni-api-endpoint", os.Getenv("OMNI_ENDPOINT"),
		"the endpoint of the Omni API, if not set, defaults to OMNI_ENDPOINT env var.")
	rootCmd.Flags().StringVar(&meta.ProviderID, "id", meta.ProviderID, "the id of the infra provider, it is used to match the resources with the infra provider label.")
	rootCmd.Flags().StringVar(&cfg.serviceAccountKey, "omni-service-account-key", os.Getenv("OMNI_SERVICE_ACCOUNT_KEY"), "Omni service account key, if not set, defaults to OMNI_SERVICE_ACCOUNT_KEY.")
	rootCmd.Flags().StringVar(&cfg.providerName, "provider-name", "KubeVirt", "provider name as it appears in Omni")
	rootCmd.Flags().StringVar(&cfg.providerDescription, "provider-description", "KubeVirt infrastructure provider", "Provider description as it appears in Omni")
	rootCmd.Flags().StringVar(&cfg.kubeconfigFile, "kubeconfig-file", "~/.kube/config", "Kubeconfig file to use to connect to the cluster where KubeVirt is running")
	rootCmd.Flags().StringVar(&cfg.namespace, "namespace", "default", "Kubernetes namespace to use for the resources created by the provider")
	rootCmd.Flags().StringVar(&cfg.dataVolumeMode, "data-volume-mode", "", "DataVolume PVC type to use (Block|Filesystem)")
	rootCmd.Flags().StringArrayVar(&cfg.configPatches, "config-patch", nil, "Applies config patches for all machines created by the infra provider")
	rootCmd.Flags().BoolVar(&cfg.insecureSkipVerify, "insecure-skip-verify", false, "ignores untrusted certs on Omni side")
}

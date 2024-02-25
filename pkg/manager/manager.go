/*
Copyright AppsCode Inc. and Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manager

import (
	"context"
	"embed"
	"go.bytebuilder.dev/license-proxyserver-addon/pkg/manager/controller"
	"go.bytebuilder.dev/license-proxyserver-addon/pkg/manager/rbac"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
	"open-cluster-management.io/addon-framework/pkg/agent"
	cmdfactory "open-cluster-management.io/addon-framework/pkg/cmd/factory"
	"open-cluster-management.io/api/addon/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//go:embed all:agent-manifests
var FS embed.FS

const (
	AddonName                  = "license-proxyserver"
	AgentName                  = "license-proxyserver"
	AgentManifestsDir          = "agent-manifests/license-proxyserver"
	AddonInstallationNamespace = "kubeops"
)

func NewRegistrationOption(kubeConfig *rest.Config, addonName, agentName string) *agent.RegistrationOption {
	return &agent.RegistrationOption{
		CSRConfigurations: agent.KubeClientSignerConfigurations(addonName, agentName),
		CSRApproveCheck:   agent.ApprovalAllCSRs,
		PermissionConfig:  rbac.SetupPermission(kubeConfig, agentName),
		AgentInstallNamespace: func(addon *v1alpha1.ManagedClusterAddOn) string {
			return AddonInstallationNamespace
		},
	}
}

func NewManagerCommand() *cobra.Command {
	cmd := cmdfactory.
		NewControllerCommandConfig(AddonName, version.Get(), runManagerController).
		NewCommand()
	cmd.Use = "manager"
	cmd.Short = "Starts the addon manager controller"

	return cmd
}

func runManagerController(ctx context.Context, cfg *rest.Config) error {
	log.SetLogger(klogr.New()) // nolint:staticcheck

	m, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		return err
	}

	if err := (controller.NewLicenseReconciler(
		m.GetClient(),
	)).SetupWithManager(m); err != nil {
		klog.Error(err, "unable to register LicenseReconciler")
		os.Exit(1)
	}

	registrationOption := NewRegistrationOption(cfg, AddonName, AgentName)

	mgr, err := addonmanager.New(cfg)
	if err != nil {
		return err
	}
	agent, err := addonfactory.NewAgentAddonFactory(AddonName, FS, AgentManifestsDir).
		WithScheme(scheme).
		WithConfigGVRs(
			schema.GroupVersionResource{Version: "v1", Resource: "secrets"},
		).
		WithGetValuesFuncs(GetConfigValues(m.GetClient())).
		WithAgentRegistrationOption(registrationOption).
		WithAgentHealthProber(agentHealthProber()).
		WithAgentInstallNamespace(func(addon *v1alpha1.ManagedClusterAddOn) string { return AddonInstallationNamespace }).
		WithCreateAgentInstallNamespace().
		BuildHelmAgentAddon()
	if err != nil {
		klog.Errorf("Failed to build agent: `%v`", err)
		return err
	}

	if err = mgr.AddAgent(agent); err != nil {
		return err
	}

	go func() {
		_ = mgr.Start(ctx)
	}()
	return m.Start(ctx)
}

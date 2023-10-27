package manager

import (
	"context"
	"embed"
	"github.com/RokibulHasan7/license-proxyserver-addon/pkg/controller"
	"github.com/RokibulHasan7/license-proxyserver-addon/pkg/rbac"
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
	_ "open-cluster-management.io/api/addon/v1alpha1"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//go:embed agent-manifests
//go:embed agent-manifests/license-proxyserver
//go:embed agent-manifests/license-proxyserver/templates/_helpers.tpl
var FS embed.FS

const (
	AddonName                  = "license-proxyserver-addon-manager"
	AgentManifestsDir          = "agent-manifests/license-proxyserver"
	AddonInstallationNamespace = "kubeops"
)

func NewRegistrationOption(kubeConfig *rest.Config, addonName, agentName string) *agent.RegistrationOption {
	return &agent.RegistrationOption{
		CSRConfigurations: agent.KubeClientSignerConfigurations(addonName, agentName),
		CSRApproveCheck:   agent.ApprovalAllCSRs,
		PermissionConfig:  rbac.SetupPermission(kubeConfig),
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

func runManagerController(ctx context.Context, kubeConfig *rest.Config) error {
	log.SetLogger(klogr.New())
	kubeClient, err := getKubeClient(kubeConfig)
	if err != nil {
		klog.Errorf("Creating kube client failed: `%v`", err)
		return err
	}

	registrationOption := NewRegistrationOption(
		kubeConfig,
		AddonName,
		"license-proxyserver-addon-manager")

	mgr, err := addonmanager.New(kubeConfig)
	if err != nil {
		return err
	}
	agent, err := addonfactory.NewAgentAddonFactory(AddonName, FS, AgentManifestsDir).
		WithScheme(scheme).
		WithConfigGVRs(
			schema.GroupVersionResource{Group: LicenseProxyServerConfigGroup, Version: LicenseProxyServerConfigVersion, Resource: LicenseProxyServerConfigResource},
		).
		WithGetValuesFuncs(GetConfigValues(kubeClient)).
		WithAgentRegistrationOption(registrationOption).
		WithAgentHealthProber(agentHealthProber()).
		WithAgentInstallNamespace(func(addon *v1alpha1.ManagedClusterAddOn) string { return AddonInstallationNamespace }).
		BuildHelmAgentAddon()
	if err != nil {
		klog.Error("Failed to build agent: `%v`", err)
		return err
	}

	if err = mgr.AddAgent(agent); err != nil {
		return err
	}

	m, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err := (controller.NewLicenseReconciler(
		kubeClient,
	)).SetupWithManager(m); err != nil {
		klog.Error(err, "unable to register LicenseReconciler")
		os.Exit(1)
	}

	go mgr.Start(ctx)
	go m.Start(ctx)
	<-ctx.Done()

	return nil
}

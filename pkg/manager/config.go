package manager

import (
	"context"
	"fmt"
	licensev1alpha1 "github.com/RokibulHasan7/license-proxyserver-addon/api/api/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	agentapi "open-cluster-management.io/addon-framework/pkg/agent"
	"open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	ocmv1 "open-cluster-management.io/api/cluster/v1"
	ocm "open-cluster-management.io/api/cluster/v1alpha1"
	workapiv1 "open-cluster-management.io/api/work/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// LicenseProxyServerConfigVersion defines the API version used for LicenseProxyServerConfig.
	LicenseProxyServerConfigVersion = "v1alpha1"

	// LicenseProxyServerConfigResource is the resource name for LicenseProxyServerConfig objects.
	LicenseProxyServerConfigResource = "licenseproxyserverconfigs"

	// LicenseProxyServerConfigGroup is the group name for LicenseProxyServerConfig objects.
	LicenseProxyServerConfigGroup = "licenses.appscode.com.open-cluster-management.io"
)

var (
	scheme = runtime.NewScheme()
)

func getKubeClient(kubeConfig *rest.Config) (client.Client, error) {
	utilruntime.Must(licensev1alpha1.AddToScheme(scheme))
	_ = ocm.Install(scheme)
	_ = ocmv1.Install(scheme)
	_ = workapiv1.Install(scheme)
	_ = apiregistrationv1.AddToScheme(scheme)
	_ = monitoringv1.AddToScheme(scheme)
	return client.New(kubeConfig, client.Options{Scheme: scheme})
}

func GetConfigValues(kc client.Client) addonfactory.GetValuesFunc {
	return func(cluster *clusterv1.ManagedCluster, addon *v1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {
		overrideValues := addonfactory.Values{}

		//for _, refConfig := range addon.Status.ConfigReferences {
		//	if refConfig.ConfigGroupResource.Group != LicenseProxyServerConfigGroup ||
		//		refConfig.ConfigGroupResource.Resource != LicenseProxyServerConfigResource {
		//		continue
		//	}

		lcConfig := licensev1alpha1.LicenseProxyServerConfig{}
		keyType := types.NamespacedName{Name: "licenseproxyserverconfig", Namespace: "open-cluster-management"}

		if err := kc.Get(context.TODO(), keyType, &lcConfig); err != nil {
			return nil, err
		}

		lcConfigSpec := lcConfig.Spec
		values, err := addonfactory.JsonStructToValues(lcConfigSpec)
		if err != nil {
			return nil, err
		}
		overrideValues = addonfactory.MergeValues(overrideValues, values)
		//}

		return overrideValues, nil
	}
}

func agentHealthProber() *agentapi.HealthProber {
	return &agentapi.HealthProber{
		Type: agentapi.HealthProberTypeWork,
		WorkProber: &agentapi.WorkHealthProber{
			ProbeFields: []agentapi.ProbeField{
				{
					ResourceIdentifier: workapiv1.ResourceIdentifier{
						Group:     "apps",
						Resource:  "deployments",
						Name:      "license-proxyserver-addon-manager",
						Namespace: AddonInstallationNamespace,
					},
					ProbeRules: []workapiv1.FeedbackRule{
						{
							Type: workapiv1.WellKnownStatusType,
						},
					},
				},
			},
			HealthCheck: func(identifier workapiv1.ResourceIdentifier, result workapiv1.StatusFeedbackResult) error {
				if len(result.Values) == 0 {
					return fmt.Errorf("no values are probed for deployment %s/%s", identifier.Namespace, identifier.Name)
				}
				for _, value := range result.Values {
					if value.Name != "ReadyReplicas" {
						continue
					}

					if *value.Value.Integer >= 1 {
						return nil
					}

					return fmt.Errorf("readyReplica is %d for deployement %s/%s", *value.Value.Integer, identifier.Namespace, identifier.Name)
				}
				return fmt.Errorf("readyReplica is not probed")
			},
		},
	}
}

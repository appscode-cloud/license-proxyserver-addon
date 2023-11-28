package controller

import (
	"context"
	"encoding/json"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	verifier "go.bytebuilders.dev/license-verifier"
	"go.bytebuilders.dev/license-verifier/apis/licenses/v1alpha1"
	licenseclient "go.bytebuilders.dev/license-verifier/client"
	"go.bytebuilders.dev/license-verifier/info"
	v1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/clock"
	ocmv1 "open-cluster-management.io/api/cluster/v1"
	ocm "open-cluster-management.io/api/cluster/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
)

func NewLicenseReconciler(hubClient client.Client) *LicenseReconciler {
	return &LicenseReconciler{
		HubClient: hubClient,
		clock:     clock.RealClock{},
	}
}

var _ reconcile.Reconciler = &LicenseReconciler{}

type LicenseReconciler struct {
	HubClient client.Client
	clock     clock.RealClock
}

// SetupWithManager sets up the controller with the Manager.
func (r *LicenseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ocmv1.ManagedCluster{}).Watches(&ocmv1.ManagedCluster{}, &handler.EnqueueRequestForObject{}).Complete(r)
}

func (r *LicenseReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Start reconciling")
	managed := &ocmv1.ManagedCluster{}
	err := r.HubClient.Get(ctx, request.NamespacedName, managed)
	if err != nil {
		return reconcile.Result{}, err
	}

	if len(managed.Status.ClusterClaims) > 1 {
		cc := ocm.ClusterClaim{}
		err := r.HubClient.Get(context.TODO(), client.ObjectKey{Name: "id.k8s.io"}, &cc)
		if err != nil {
			return reconcile.Result{}, err
		}

		cid := cc.Spec.Value

		err = r.HubClient.Get(context.TODO(), client.ObjectKey{Name: "licenses.appscode.com"}, &cc)
		if err != nil {
			return reconcile.Result{}, err
		}

		features := strings.Split(cc.Spec.Value, ",")
		err = licenseHelper(ctx, r.HubClient, cid, features, managed.Name)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, err
}

func licenseHelper(ctx context.Context, HubClient client.Client, cid string, features []string, clusterName string) error {
	lps := v2beta1.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "license-proxyserver",
			Namespace: "kubeops",
		},
	}

	err := HubClient.Get(context.TODO(), client.ObjectKey{Name: lps.Name, Namespace: lps.Namespace}, &lps)
	if err == nil {
		baseURL, token, err := getProxyServerURLAndToken(lps)
		if err != nil {
			return err
		}

		l, err := getLicense(baseURL, token, cid, features)
		if err != nil {
			return err
		}

		// get secret
		sec := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "license-proxyserver-licenses",
				Namespace: clusterName,
			},
		}
		err = HubClient.Get(context.TODO(), client.ObjectKey{Name: sec.Name}, &sec)
		if err != nil && kerr.IsNotFound(err) {
			// create secret
			var data map[string][]byte
			data[l.PlanName] = l.Data
			sec.Data = data
			err = HubClient.Create(ctx, &sec)
			if err != nil {
				return err
			} else {
				return nil
			}
		} else if err != nil {
			return err
		}

		// update secret
		sec.Data[l.PlanName] = l.Data
		err = HubClient.Update(context.TODO(), &sec)
		if err != nil {
			return err
		}
		return nil
	}

	return err
}

func getProxyServerURLAndToken(lps v2beta1.HelmRelease) (string, string, error) {
	val := make(map[string]interface{})
	jsonByte, err := json.Marshal(lps.Spec.Values)
	if err != nil {
		return "", "", err
	}
	if err = json.Unmarshal(jsonByte, &val); err != nil {
		return "", "", err
	}
	baseURL, _, err := unstructured.NestedString(val, "platform", "baseURL")
	if err != nil {
		return "", "", err
	}
	token, _, err := unstructured.NestedString(val, "platform", "token")
	if err != nil {
		return "", "", err
	}

	return baseURL, token, nil
}

func getLicense(baseURL, token, cid string, features []string) (*v1alpha1.License, error) {
	lc, err := licenseclient.NewClient(baseURL, token, cid)
	if err != nil {
		return nil, err
	}

	nl, _, err := lc.AcquireLicense(features)
	if err != nil {
		return nil, err
	}

	caData, err := info.LoadLicenseCA()
	if err != nil {
		return nil, err
	}
	caCert, err := info.ParseCertificate(caData)
	if err != nil {
		return nil, err
	}

	l, err := verifier.ParseLicense(verifier.ParserOptions{
		ClusterUID: cid,
		CACert:     caCert,
		License:    nl,
	})
	if err != nil {
		return nil, err
	}

	return &l, nil
}

package rbac

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/pointer"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"open-cluster-management.io/addon-framework/pkg/agent"
	addonapiv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

func SetupPermission(kubeConfig *rest.Config) agent.PermissionConfigFunc {
	return func(cluster *clusterv1.ManagedCluster, addon *addonapiv1alpha1.ManagedClusterAddOn) error {
		nativeClient, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			return err
		}
		namespace := cluster.Name
		agentUser := "system:open-cluster-management:cluster:" + cluster.Name + ":addon:license-proxyserver-addon-manager:agent:license-proxyserver-addon-manager"

		role := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      addon.Name,
				Namespace: namespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         "addon.open-cluster-management.io/v1alpha1",
						Kind:               "ManagedClusterAddOn",
						UID:                addon.UID,
						Name:               addon.Name,
						BlockOwnerDeletion: pointer.Bool(true),
					},
				},
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Verbs:     []string{"get", "list", "watch", "create", "update"},
					Resources: []string{"secrets"},
				},
			},
		}
		roleBinding := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      addon.Name,
				Namespace: namespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         "addon.open-cluster-management.io/v1alpha1",
						Kind:               "ManagedClusterAddOn",
						UID:                addon.UID,
						Name:               addon.Name,
						BlockOwnerDeletion: pointer.Bool(true),
					},
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "Role",
				Name: addon.Name,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind: rbacv1.UserKind,
					Name: agentUser,
				},
			},
		}

		_, err = nativeClient.RbacV1().Roles(cluster.Name).Get(context.TODO(), role.Name, metav1.GetOptions{})
		switch {
		case errors.IsNotFound(err):
			_, createErr := nativeClient.RbacV1().Roles(cluster.Name).Create(context.TODO(), role, metav1.CreateOptions{})
			if createErr != nil {
				return createErr
			}
		case err != nil:
			return err
		}

		_, err = nativeClient.RbacV1().RoleBindings(cluster.Name).Get(context.TODO(), roleBinding.Name, metav1.GetOptions{})
		switch {
		case errors.IsNotFound(err):
			_, createErr := nativeClient.RbacV1().RoleBindings(cluster.Name).Create(context.TODO(), roleBinding, metav1.CreateOptions{})
			if createErr != nil {
				return createErr
			}
		case err != nil:
			return err
		}

		return nil
	}
}

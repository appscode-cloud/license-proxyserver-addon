/*
Copyright 2023.

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

package v1alpha1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LicenseProxyServerConfigSpec defines the desired state of LicenseProxyServerConfig
type LicenseProxyServerConfigSpec struct {
	// +optional
	NameOverride string `json:"nameOverride"`
	// +optional
	FullnameOverride string `json:"fullnameOverride"`
	// +optional
	ReplicaCount int `json:"replicaCount"`
	// +optional
	RegistryFQDN string `json:"registryFQDN"`
	// +optional
	Image Image `json:"image"`
	// +optional
	ImagePullSecrets []core.LocalObjectReference `json:"imagePullSecrets"`
	// +optional
	//+kubebuilder:validation:Enum=Always;Never;IfNotPresent;""
	ImagePullPolicy core.PullPolicy `json:"imagePullPolicy"`
	// +optional
	CriticalAddon bool `json:"criticalAddon"`
	// +optional
	LogLevel int `json:"logLevel"`
	// +optional
	Annotations map[string]string `json:"annotations"`
	// +optional
	PodAnnotations map[string]string `json:"podAnnotations"`
	// +optional
	NodeSelector NodeSelector `json:"nodeSelector"`
	// +optional
	Tolerations []core.Toleration `json:"tolerations"`
	// +optional
	Affinity core.Affinity `json:"affinity"`
	// +optional
	PodSecurityContext PodSecurityContext `json:"podSecurityContext"`
	// +optional
	ServiceAccount ServiceAccount `json:"serviceAccount"`
	// +optional
	Apiserver Apiserver `json:"apiserver"`
	// +optional
	Monitoring Monitoring `json:"monitoring"`
	// +optional
	Platform Platform `json:"platform"`
	// +optional
	Licenses map[string]string `json:"licenses"`
	// +optional
	EncodedLicenses map[string]string `json:"encodedLicenses"`
}

type Capabilities struct {
	// +optional
	Drop []string `json:"drop"`
}
type SeccompProfile struct {
	// +optional
	Type string `json:"type"`
}
type SecurityContext struct {
	// +optional
	AllowPrivilegeEscalation bool `json:"allowPrivilegeEscalation"`
	// +optional
	Capabilities Capabilities `json:"capabilities"`
	// +optional
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem"`
	// +optional
	RunAsNonRoot bool `json:"runAsNonRoot"`
	// +optional
	RunAsUser int `json:"runAsUser"`
	// +optional
	SeccompProfile SeccompProfile `json:"seccompProfile"`
}
type Image struct {
	// +optional
	Registry string `json:"registry"`
	// +optional
	Repository string `json:"repository"`
	// +optional
	Tag string `json:"tag"`
	// +optional
	Resources map[string]string `json:"resources"`
	// +optional
	SecurityContext SecurityContext `json:"securityContext"`
}

type NodeSelector struct {
	// +optional
	KubernetesIoOs string `json:"kubernetes.io/os"`
}

type PodSecurityContext struct {
	// +optional
	FsGroup int `json:"fsGroup"`
}
type ServiceAccount struct {
	// +optional
	Create bool `json:"create"`
	// +optional
	Annotations map[string]string `json:"annotations"`
	// +optional
	Name string `json:"name"`
}
type Healthcheck struct {
	// +optional
	Enabled bool `json:"enabled"`
}
type ServingCerts struct {
	// +optional
	Generate bool `json:"generate"`
	// +optional
	CaCrt string `json:"caCrt"`
	// +optional
	ServerCrt string `json:"serverCrt"`
	// +optional
	ServerKey string `json:"serverKey"`
}
type Apiserver struct {
	// +optional
	GroupPriorityMinimum int `json:"groupPriorityMinimum"`
	// +optional
	VersionPriority int `json:"versionPriority"`
	// +optional
	UseKubeapiserverFqdnForAks bool `json:"useKubeapiserverFqdnForAks"`
	// +optional
	Healthcheck Healthcheck `json:"healthcheck"`
	// +optional
	ServingCerts ServingCerts `json:"servingCerts"`
}

type ServiceMonitor struct {
	// +optional
	Labels map[string]string `json:"labels"`
}

type Monitoring struct {
	// +optional
	Agent string `json:"agent"`
	// +optional
	ServiceMonitor ServiceMonitor `json:"serviceMonitor"`
}

type Platform struct {
	// +optional
	BaseURL string `json:"baseURL"`
	// +optional
	Token string `json:"token"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LicenseProxyServerConfig is the Schema for the licenseproxyserverconfigs API
type LicenseProxyServerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec LicenseProxyServerConfigSpec `json:"spec"`
}

//+kubebuilder:object:root=true

// LicenseProxyServerConfigList contains a list of LicenseProxyServerConfig
type LicenseProxyServerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LicenseProxyServerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LicenseProxyServerConfig{}, &LicenseProxyServerConfigList{})
}

/*
Copyright 2022.

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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentRuleSpec defines the desired state of DeploymentRule
type DeploymentRuleSpec struct {

	// Environment defines the Sleuth Environment to register the deployment against
	Environment string `json:"environment,omitempty"`
	// WebhookUrl defines the Sleuth Webhook to register the deployment at
	WebhookUrl string `json:"webhookUrl,omitempty"`
	// ApiToken defines the location within a Kubernetes Secret for the Sleuth Webhook api token
	ApiToken ApiToken `json:"apiToken,omitempty"`
	// Selector defines which labels the rule is matching against in order to trigger a Sleuth deployment
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}

type ApiToken struct {
	SecretKeyRef *v1.SecretKeySelector `json:"secretKeyRef,omitempty"`
	Value        *string               `json:"value,omitempty"`
}

// DeploymentRuleStatus defines the observed state of DeploymentRule
type DeploymentRuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DeploymentRule is the Schema for the deploymentrules API
type DeploymentRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentRuleSpec   `json:"spec,omitempty"`
	Status DeploymentRuleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DeploymentRuleList contains a list of DeploymentRule
type DeploymentRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeploymentRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeploymentRule{}, &DeploymentRuleList{})
}

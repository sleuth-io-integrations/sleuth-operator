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

package controllers

import (
	"context"
	corev1alpha1 "github.com/sleuth-io-integrations/sleuth-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("DeploymentRule controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		DeploymentRuleName      = "test-deployment-rule"
		DeploymentRuleNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	It("Should create a new DeploymentRule", func() {
		ctx := context.WithValue(context.Background(), "TODO", nil)
		secret := "secret"
		deploymentRule := &corev1alpha1.DeploymentRule{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentRule",
				APIVersion: "apiextensions.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      DeploymentRuleName,
				Namespace: DeploymentRuleNamespace,
			},
			Spec: corev1alpha1.DeploymentRuleSpec{
				Environment: "default",
				WebhookUrl:  "https://webhook.sleuth.io/test",
				ApiToken: corev1alpha1.ApiToken{
					SecretKeyRef: nil,
					Value:        &secret,
				},
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "test-app",
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, deploymentRule)).Should(Succeed())

		deploymentRuleLookupKey := types.NamespacedName{Name: DeploymentRuleName, Namespace: DeploymentRuleNamespace}
		createdDeploymentRule := &corev1alpha1.DeploymentRule{}

		// We'll need to retry getting this newly created DeploymentRule, given that creation may not immediately happen.
		Eventually(func() bool {
			err := k8sClient.Get(ctx, deploymentRuleLookupKey, createdDeploymentRule)
			if err != nil {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())

		// Let's make sure our Schedule string value was properly converted/handled.
		Expect(createdDeploymentRule.Spec.WebhookUrl).Should(Equal("https://webhook.sleuth.io/test"))
	})
})

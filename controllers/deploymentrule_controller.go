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
	"bytes"
	"context"
	"errors"
	"fmt"
	corev1alpha1 "github.com/sleuth-io-integrations/sleuth-operator/api/v1alpha1"
	kapps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
)

const (
	deploymentField = ".spec.selector.matchLabels"
	sleuthLabel     = "sleuth-label"
	shaLabel        = "sha"
)

// DeploymentRuleReconciler reconciles a DeploymentRule object
type DeploymentRuleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.sleuth.io,resources=deploymentrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.sleuth.io,resources=deploymentrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.sleuth.io,resources=deploymentrules/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch

// Reconcile does blah blah
func (r *DeploymentRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var deploymentRule corev1alpha1.DeploymentRule
	if err := r.Get(ctx, req.NamespacedName, &deploymentRule); err != nil {
		logger.Error(err, "unable to fetch DeploymentRule")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Another question
	// - how do we know a deployment was successfully rolled out
	// - are watching Deployments the right call..
	// i.e. we need a guarantee that the deploy status has finished its rolling deployment

	// Query deployments that match this DeploymentRule
	// - how do we detect recently changed Deployments (maintaining hash of Deployment or keep DeploymentVersion
	// maintained on DeploymentRule)
	// - use hash/version to determine whether to trigger webhook
	var deployments, err = r.findDeploymentsForRule(deploymentRule.Spec.Selector)
	if err != nil {
		logger.Error(err, "failed to fetch Deployments")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if len(deployments.Items) == 0 {
		logger.Info("no matching Deployments")
		return ctrl.Result{}, nil
	}

	// Heuristics for parsing SHA from resource.. annotations.... image tags... labels....
	logger.Info(fmt.Sprintf("found %n deployments for DeploymentRule %s", len(deployments.Items), deploymentRule.Name))

	// Call webhook defined within DeploymentRule
	for _, deployment := range deployments.Items {
		if err := r.PostDeploymentWebhook(deploymentRule, deployment); err != nil {
			logger.Error(err, "failed calling deployment webhook")
		}
	}

	return ctrl.Result{}, nil
}

func (r *DeploymentRuleReconciler) PostDeploymentWebhook(deploymentRule corev1alpha1.DeploymentRule, deployment kapps.Deployment) error {
	jsonBody, _ := json.Marshal(corev1alpha1.RegisterDeployBody{
		IgnoreIfDuplicate: true,
		Tags:              "",
		Environment:       deploymentRule.Spec.Environment,
		Email:             "",
		Date:              "",
		ApiKey:            *deploymentRule.Spec.ApiToken.Value,
		Sha:               deployment.GetLabels()[shaLabel],
		Links:             nil,
	})
	postBody := bytes.NewBuffer(jsonBody)

	response, err := http.Post(deploymentRule.Spec.WebhookUrl, "application/json", postBody)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1alpha1.DeploymentRule{}, deploymentField, func(rawObj client.Object) []string {
		deploymentRule := rawObj.(*corev1alpha1.DeploymentRule)
		value := getSleuthSelector(deploymentRule.Spec.Selector.MatchLabels)
		return []string{value}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.DeploymentRule{}).
		WithEventFilter(DeploymentRuleEventPredicate()).
		Watches(
			&source.Kind{Type: &kapps.Deployment{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForDeployment),
			builder.WithPredicates(
				predicate.ResourceVersionChangedPredicate{},
				DeploymentEventPredicate(),
			),
		).
		Complete(r)
}

func DeploymentRuleEventPredicate() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: nil,
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return false
		},
		UpdateFunc:  nil,
		GenericFunc: nil,
	}
}

func DeploymentEventPredicate() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			//return len(createEvent.Object.GetLabels()[sleuthLabel]) > 0
			return false
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return len(updateEvent.ObjectNew.GetLabels()[sleuthLabel]) > 0 &&
				updateEvent.ObjectNew.GetLabels()[shaLabel] != "" &&
				updateEvent.ObjectNew.GetLabels()[shaLabel] != updateEvent.ObjectOld.GetLabels()[shaLabel]

		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			// Suppress Delete events
			return false
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			// Suppress Generic events
			return false
		},
	}
}

func getValue(labelSelector *metav1.LabelSelector) string {
	var sb strings.Builder
	for label, value := range labelSelector.MatchLabels {
		if label == "" || value == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("%s:%s,", label, value))
	}
	return sb.String()
}

func getSleuthSelector(labels map[string]string) string {
	return fmt.Sprintf("%s:%s,", sleuthLabel, labels[sleuthLabel])
}

func (r *DeploymentRuleReconciler) findObjectsForDeployment(rawObj client.Object) []reconcile.Request {
	attachedDeploymentRules := &corev1alpha1.DeploymentRuleList{}
	deployment := rawObj.(*kapps.Deployment)

	deploymentValue := getSleuthSelector(deployment.ObjectMeta.GetLabels())

	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(deploymentField, deploymentValue),
	}
	if err := r.List(context.TODO(), attachedDeploymentRules, listOps); err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(attachedDeploymentRules.Items))
	for i, item := range attachedDeploymentRules.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

func (r *DeploymentRuleReconciler) findDeploymentsForRule(labelSelector *metav1.LabelSelector) (*kapps.DeploymentList, error) {
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return nil, err
	}

	listOps := &client.ListOptions{
		LabelSelector: selector,
	}

	var deploymentList kapps.DeploymentList
	if err := r.List(context.TODO(), &deploymentList, listOps); err != nil {
		return nil, err
	}

	return &deploymentList, nil
}

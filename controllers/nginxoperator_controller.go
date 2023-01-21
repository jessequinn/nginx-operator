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

package controllers

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"time"

	apiv2 "github.com/operator-framework/api/pkg/operators/v2"
	"github.com/operator-framework/operator-lib/conditions"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha2 "github.com/jessequinn/nginx-operator/api/v1alpha2"
	"github.com/jessequinn/nginx-operator/assets"
	"github.com/jessequinn/nginx-operator/controllers/metrics"
)

// NginxOperatorReconciler reconciles a NginxOperator object
type NginxOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operator.example.com,resources=nginxoperators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.example.com,resources=nginxoperators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.example.com,resources=nginxoperators/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NginxOperator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *NginxOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	metrics.ReconcilesTotal.Inc()
	logger := log.FromContext(ctx)

	operatorCR := &operatorv1alpha2.NginxOperator{}
	err := r.Get(ctx, req.NamespacedName, operatorCR)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Operator resource object not found.")
		return ctrl.Result{}, nil
	} else if err != nil {
		logger.Error(err, "Error getting operator resource object")
		meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
			Type:               "OperatorDegraded",
			Status:             metav1.ConditionTrue,
			Reason:             operatorv1alpha2.ReasonCRNotAvailable,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Message:            fmt.Sprintf("unable to get operator custom resource: %s", err.Error()),
		})
		return ctrl.Result{}, utilerrors.NewAggregate([]error{err, r.Status().Update(ctx, operatorCR)})
	}

	// Create/update deployment
	deployment := &appsv1.Deployment{}
	createDeployment := false
	err = r.Get(ctx, req.NamespacedName, deployment)
	if err != nil && errors.IsNotFound(err) {
		createDeployment = true
		deployment, err = assets.GetDeploymentFromFile("manifests/nginx_deployment.yaml")
		if err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		logger.Error(err, "Error getting existing Nginx deployment.")
		meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
			Type:               "OperatorDegraded",
			Status:             metav1.ConditionTrue,
			Reason:             operatorv1alpha2.ReasonDeploymentNotAvailable,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Message:            fmt.Sprintf("unable to get operand deployment: %s", err.Error()),
		})
		return ctrl.Result{}, utilerrors.NewAggregate([]error{err, r.Status().Update(ctx, operatorCR)})
	}

	deployment.Namespace = req.Namespace
	deployment.Name = req.Name //fmt.Sprintf("%s-deployment", req.Name)

	if operatorCR.Spec.Replicas != nil {
		deployment.Spec.Replicas = operatorCR.Spec.Replicas
	}

	if len(operatorCR.Spec.Ports) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Ports = operatorCR.Spec.Ports
	}

	// Add Volumes and VolumeMounts to be able to customize Nginx
	if len(operatorCR.Spec.VolumeMounts) > 0 && len(operatorCR.Spec.Volumes) > 0 {
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = operatorCR.Spec.VolumeMounts
		deployment.Spec.Template.Spec.Volumes = operatorCR.Spec.Volumes
	}

	if len(operatorCR.Spec.Selectors) > 0 {
		// Set the spec selector
		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: operatorCR.Spec.Selectors,
		}
		// Set the template metadata labels
		deployment.Spec.Template.ObjectMeta.Labels = operatorCR.Spec.Selectors
		// Set the metadata labels of the deployment
		deployment.ObjectMeta.Labels = operatorCR.Spec.Selectors
	}

	if err := ctrl.SetControllerReference(operatorCR, deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	if createDeployment {
		if err := r.Create(ctx, deployment); err != nil {
			meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
				Type:               "OperatorDegraded",
				Status:             metav1.ConditionTrue,
				Reason:             operatorv1alpha2.ReasonOperandDeploymentFailed,
				LastTransitionTime: metav1.NewTime(time.Now()),
				Message:            fmt.Sprintf("unable to create operand deployment: %s", err.Error()),
			})
			return ctrl.Result{}, utilerrors.NewAggregate([]error{err, r.Status().Update(ctx, operatorCR)})
		}
	} else {
		if err := r.Update(ctx, deployment); err != nil {
			meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
				Type:               "OperatorDegraded",
				Status:             metav1.ConditionTrue,
				Reason:             operatorv1alpha2.ReasonOperandDeploymentFailed,
				LastTransitionTime: metav1.NewTime(time.Now()),
				Message:            fmt.Sprintf("unable to update operand deployment: %s", err.Error()),
			})
			return ctrl.Result{}, utilerrors.NewAggregate([]error{err, r.Status().Update(ctx, operatorCR)})
		}
	}

	// Create/update service
	service := &v1.Service{}
	createService := false
	err = r.Get(ctx, req.NamespacedName, service)
	if err != nil && errors.IsNotFound(err) {
		createService = true
		service, err = assets.GetServiceFromFile("manifests/nginx_service.yaml")
		if err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		logger.Error(err, "Error getting existing Nginx service.")
		meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
			Type:               "OperatorDegraded",
			Status:             metav1.ConditionTrue,
			Reason:             operatorv1alpha2.ReasonServiceNotAvailable,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Message:            fmt.Sprintf("unable to get operand service: %s", err.Error()),
		})
		return ctrl.Result{}, utilerrors.NewAggregate([]error{err, r.Status().Update(ctx, operatorCR)})
	}

	service.Namespace = req.Namespace
	service.Name = req.Name //fmt.Sprintf("%s-service", req.Name)

	if len(operatorCR.Spec.ServicePorts) > 0 {
		service.Spec.Ports = operatorCR.Spec.ServicePorts
	}

	if operatorCR.Spec.ServiceType == v1.ServiceTypeClusterIP || operatorCR.Spec.ServiceType == v1.ServiceTypeLoadBalancer {
		service.Spec.Type = operatorCR.Spec.ServiceType
	}

	if len(operatorCR.Spec.Selectors) > 0 {
		// Set the spec selector
		service.Spec.Selector = operatorCR.Spec.Selectors
		// Set the metadata labels of the service
		service.ObjectMeta.Labels = operatorCR.Spec.Selectors
	}

	if err := ctrl.SetControllerReference(operatorCR, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	if createService {
		if err := r.Create(ctx, service); err != nil {
			meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
				Type:               "OperatorDegraded",
				Status:             metav1.ConditionTrue,
				Reason:             operatorv1alpha2.ReasonOperandServiceFailed,
				LastTransitionTime: metav1.NewTime(time.Now()),
				Message:            fmt.Sprintf("unable to create operand service: %s", err.Error()),
			})
			return ctrl.Result{}, utilerrors.NewAggregate([]error{err, r.Status().Update(ctx, operatorCR)})
		}
	} else {
		if err := r.Update(ctx, service); err != nil {
			meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
				Type:               "OperatorDegraded",
				Status:             metav1.ConditionTrue,
				Reason:             operatorv1alpha2.ReasonOperandServiceFailed,
				LastTransitionTime: metav1.NewTime(time.Now()),
				Message:            fmt.Sprintf("unable to update operand service: %s", err.Error()),
			})
			return ctrl.Result{}, utilerrors.NewAggregate([]error{err, r.Status().Update(ctx, operatorCR)})
		}
	}

	meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
		Type:               "OperatorDegraded",
		Status:             metav1.ConditionFalse,
		Reason:             operatorv1alpha2.ReasonSucceeded,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            "operator successfully reconciling",
	})

	if err := r.Status().Update(ctx, operatorCR); err != nil {
		return ctrl.Result{}, err
	}

	condition, err := conditions.InClusterFactory{Client: r.Client}.
		NewCondition(apiv2.ConditionType(apiv2.Upgradeable))
	if err != nil {
		return ctrl.Result{}, err
	}

	err = condition.Set(ctx, metav1.ConditionTrue,
		conditions.WithReason("OperatorUpgradeable"),
		conditions.WithMessage("The operator is currently upgradeable"))
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, utilerrors.NewAggregate([]error{err, r.Status().Update(ctx, operatorCR)})
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha2.NginxOperator{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

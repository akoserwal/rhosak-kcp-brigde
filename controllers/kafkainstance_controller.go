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
	"fmt"
	"github.com/akoserwal/rhosak-kcp-brigde/controllers/utils"
	"github.com/kcp-dev/logicalcluster/v2"
	se "github.com/pkg/errors"
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	cache3 "k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	rhosakkcpv1 "github.com/akoserwal/rhosak-kcp-brigde/api/v1"
)

// KafkaInstanceReconciler reconciles a KafkaInstance object
type KafkaInstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cache  cache3.ExpirationCache
}

//+kubebuilder:rbac:groups=rhosak.kcp.github.com,resources=kafkainstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rhosak.kcp.github.com,resources=kafkainstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rhosak.kcp.github.com,resources=kafkainstances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KafkaInstance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *KafkaInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := log.FromContext(ctx)
	// Include the clusterName from req.ObjectKey in the logger, similar to the namespace and name keys that are already
	// there.
	logger = logger.WithValues("clusterName", req.ClusterName)

	var allKafkaInstance rhosakkcpv1.KafkaInstanceList
	logger.Info("Listed all widgets across all workspaces", "count", len(allKafkaInstance.Items))
	if err := r.List(ctx, &allKafkaInstance); err != nil {
		return ctrl.Result{}, err
	}

	// Add the logical cluster to the context
	ctx = logicalcluster.WithCluster(ctx, logicalcluster.New(req.ClusterName))
	logger.Info("Getting KafkaRequests")

	var kafkaInstance rhosakkcpv1.KafkaInstance
	if err := r.Get(ctx, req.NamespacedName, &kafkaInstance); err != nil {
		if errors.IsNotFound(err) {
			// Normal - was deleted
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}
	logger.Info("Kafka instance name", "name", kafkaInstance.Spec.Name)
	logger.Info("kafka instance cloud", "cloud", kafkaInstance.Spec.CloudProvider)

	var offlinetoken = kafkaInstance.Spec.OfflineToken
	if offlinetoken != "" {

		if (kafkaInstance.Status.Phase != rhosakkcpv1.KafkaUnknown && kafkaInstance.Status.Phase != "") && kafkaInstance.Status.Id != "" {

			finalizerName := "kafka-instance.kafka.rhosak/finalizer"

			if kafkaInstance.ObjectMeta.DeletionTimestamp.IsZero() {
				if !utils.ContainsString(kafkaInstance.GetFinalizers(), finalizerName) {
					logger.Info("adding finalizer")
					controllerutil.AddFinalizer(&kafkaInstance, finalizerName)

					if err := r.Update(ctx, &kafkaInstance); err != nil {
						return ctrl.Result{}, se.WithStack(err)
					}
				}
			} else {
				// The object is being deleted
				if utils.ContainsString(kafkaInstance.GetFinalizers(), finalizerName) {
					// our finalizer is present, so lets handle any external dependency

					// Check state from the API
					c := utils.BuildControlAPIClient(offlinetoken, utils.DefaultClientID, utils.DefaultAuthURL, utils.DefaultAPIURL)
					if kafkaRequest, res, err := c.DefaultApi.GetKafkaById(ctx, kafkaInstance.Status.Id).Execute(); err != nil {
						if res.StatusCode == 404 {
							// The remote resource is now deleted so we can remove the resource
							// remove our finalizer from the list and update it.
							logger.Info("deletion complete, removing resource")
							controllerutil.RemoveFinalizer(&kafkaInstance, finalizerName)
							if err := r.Update(ctx, &kafkaInstance); err != nil {
								return ctrl.Result{}, se.WithStack(err)
							}
							// Stop reconcilation
							return ctrl.Result{}, nil
						}
						return ctrl.Result{}, err
					} else {
						if kafkaRequest.Status != nil && *kafkaRequest.Status != "deleting" && *kafkaRequest.Status != "deprovision" {
							if err := r.deleteKafkaInstance(ctx, &kafkaInstance, offlinetoken); err != nil {
								// if fail to delete the external dependency here, return with error
								// so that it can be retried
								return ctrl.Result{}, se.WithStack(err)
							}
						}
					}
				}
				logger.Info("deletion pending")
				return ctrl.Result{
					RequeueAfter: utils.WatchInterval,
				}, nil
			}
		}

		answer, err := r.createKafkaInstance(ctx, &kafkaInstance, offlinetoken)
		if err != nil {
			return ctrl.Result{}, se.WithStack(err)
		}
		return answer, nil
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rhosakkcpv1.KafkaInstance{}).
		Complete(r)
}

func (r *KafkaInstanceReconciler) createKafkaInstance(ctx context.Context, kafkaInstance *rhosakkcpv1.KafkaInstance, offlineToken string) (ctrl.Result, error) {
	c := utils.BuildControlAPIClient(offlineToken, utils.DefaultClientID, utils.DefaultAuthURL, utils.DefaultAPIURL)
	log := log.FromContext(ctx)
	annotations := kafkaInstance.ObjectMeta.GetAnnotations()
	if kafkaInstance.Status.Id != "" || kafkaInstance.Status.Phase != "" {
		return ctrl.Result{}, nil
	}
	if annotations != nil {
		if annotations["rhosak.kcp/created-externally"] == "true" {
			return ctrl.Result{}, nil
		}
	}
	fmt.Println("payload")
	payload := kafkamgmtclient.KafkaRequestPayload{
		CloudProvider:           stringToPointer(kafkaInstance.Spec.CloudProvider),
		Name:                    kafkaInstance.Spec.Name,
		Region:                  stringToPointer(kafkaInstance.Spec.Region),
		ReauthenticationEnabled: *kafkamgmtclient.NewNullableBool(kafkaInstance.Spec.ReauthenticationEnabled),
	}
	fmt.Println(payload)
	kafkaRequest, _, err := c.DefaultApi.CreateKafka(ctx).KafkaRequestPayload(payload).Async(true).Execute()
	if err != nil {
		apiErr := utils.GetAPIError(err)
		kafkaInstance.Status.Message = fmt.Sprintf("%s %s", apiErr.GetCode(), apiErr.GetReason())
		kafkaInstance.Status.Phase = rhosakkcpv1.KafkaUnknown
		log.Info("Error creating Kafka using kas-fleet-manager", "Message", kafkaInstance.Status.Message, "KafkaInstance", kafkaInstance)
		if err := r.Status().Update(ctx, kafkaInstance); err != nil {
			return ctrl.Result{}, se.WithStack(err)
		}
		return ctrl.Result{}, err
	} else {
		log.Info("Kafka Request", "admin api", kafkaRequest.AdminApiServerUrl)
	}
	return ctrl.Result{}, utils.UpdateKafkaInstanceStatus(r.Client, ctx, kafkaRequest, kafkaInstance)
}
func (r *KafkaInstanceReconciler) deleteKafkaInstance(ctx context.Context, kafkaInstance *rhosakkcpv1.KafkaInstance, offlineToken string) error {
	log := log.FromContext(ctx)
	// try to find the Kafka
	if kafkaInstance.Status.Id != "" {
		c := utils.BuildControlAPIClient(offlineToken, utils.DefaultClientID, utils.DefaultAuthURL, utils.DefaultAPIURL)
		log.Info("Deleting Kafka using kas-fleet-manager", "KafkaInstance", kafkaInstance)
		_, _, err := c.DefaultApi.DeleteKafkaById(ctx, kafkaInstance.Status.Id).Async(true).Execute()
		if err != nil {
			apiErr := utils.GetAPIError(err)
			kafkaInstance.Status.Message = fmt.Sprintf("%s %s", apiErr.GetCode(), apiErr.GetReason())
			kafkaInstance.Status.Phase = rhosakkcpv1.KafkaUnknown
			log.Info("Error deleting Kafka using kas-fleet-manager", "Message", kafkaInstance.Status.Message, "KafkaInstance", kafkaInstance)
			if err := r.Status().Update(ctx, kafkaInstance); err != nil {
				return err
			}
			return nil
		}
		return nil
	}
	return se.New(fmt.Sprintf("kafka instance has no Id. Kafka instance status: %v", kafkaInstance.Status))
}
func stringToPointer(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}

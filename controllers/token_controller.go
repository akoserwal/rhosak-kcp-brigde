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
	kcpclient "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	"github.com/kcp-dev/logicalcluster/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	rhosakkcpv1 "github.com/akoserwal/rhosak-kcp-brigde/api/v1"
)

// TokenReconciler reconciles a Token object
type TokenReconciler struct {
	client.Client
	//	coreClient kubernetes.ClusterInterface
	Scheme     *runtime.Scheme
	KcpClient  *kcpclient.Cluster
	KubeClient *kubernetes.Cluster
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=configmaps/finalizers,verbs=update

//+kubebuilder:rbac:groups=rhosak.kcp.github.com,resources=tokens,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rhosak.kcp.github.com,resources=tokens/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rhosak.kcp.github.com,resources=tokens/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Token object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *TokenReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	// Include the clusterName from req.ObjectKey in the logger, similar to the namespace and name keys that are already
	// there.
	logger = logger.WithValues("clusterName", req.ClusterName)

	var allTokens rhosakkcpv1.TokenList
	if err := r.List(ctx, &allTokens); err != nil {
		return ctrl.Result{}, err
	}

	var llistsecrets v1.SecretList
	if err := r.List(ctx, &llistsecrets); err != nil {
		logger.Error(err, "unable to list secrets")
		return ctrl.Result{}, nil
	}
	logger.Info("List secrets: got", "itemCount", len(llistsecrets.Items))

	logger.Info("Listed all token across all workspaces", "count", len(allTokens.Items))

	// Add the logical cluster to the context
	ctx = logicalcluster.WithCluster(ctx, logicalcluster.New(req.ClusterName))

	logger.Info("Getting Token")
	var w rhosakkcpv1.Token
	if err := r.Get(ctx, req.NamespacedName, &w); err != nil {
		if errors.IsNotFound(err) {
			// Normal - was deleted
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	logger.Info("listing name token", "name", w.Name)

	var token = string(w.Spec.OfflineToken)

	fmt.Println("kube cluster client")
	fmt.Println(req.ClusterName)

	if token != "" {
		fmt.Println("token is not null")

		fmt.Println(w.Name)
		fmt.Println(w.Namespace)
		fmt.Println(w.OwnerReferences)
		var secret1 = v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-adsd",
				Namespace: req.Namespace,
			},
			StringData: map[string]string{
				"offlineToken": token,
			},
		}

		err := r.Client.Create(ctx, &secret1)
		if err != nil {
			fmt.Println(err)
		}
		secretName := types.NamespacedName{
			Namespace: "default",
			Name:      "test-adsd",
		}
		var ssoRedHatComSecret v1.Secret
		err = r.Client.Get(ctx, secretName, &ssoRedHatComSecret)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(ssoRedHatComSecret.Data["offlineToken"])

		//clusterName := logicalcluster.New("root:default")
		//scopedContext := roundtripper.WithCluster(ctx, lcluster.New("root:default"))
		//klog.Info(clusterName)
		//
		//clusterClient, err := client.NewForConfig(config, lcluster.New("root:default"))
		//if err != nil {
		//	klog.Fatal(err)
		//}
		//
		//sec, err := clusterClient.Cluster(clusterName).CoreV1().Secrets("default").Get(scopedContext, "mysecret", metav1.GetOptions{})
		//if err != nil {
		//	klog.Fatal(err)
		//}
		//
		//fmt.Println(sec.Name)
		//fmt.Println(sec.ClusterName)

		//llop, err := r.KubeClient.Cluster(logicalcluster.New("root:rhosak")).CoreV1().Secrets("default").Create(ctx, &secret1, metav1.CreateOptions{})
		//if err != nil {
		//	fmt.Println(err)
		//}
		//fmt.Println(llop)
		//
		//fmt.Println(secret1)
		//var secret1 = v1.Secret{
		//	ObjectMeta: metav1.ObjectMeta{
		//		Name:      w.Name,
		//		Namespace: req.Namespace,
		//	},
		//	StringData: map[string]string{
		//		"offlineToken": token,
		//	},
		//}
		//err := r.Client.Create(ctx, &secret1)
		//if err != nil {
		//	fmt.Println(err)
		//
		//}
	}

	//nsKey := types.NamespacedName{Name: w.Namespace}
	//_, er := r.coreClient.Cluster(logicalcluster.New(req.ClusterName)).CoreV1().Secrets("default").List(ctx, metav1.ListOptions{})
	//if err != nil {
	//	logger.Error(err, "unable to list secrets")
	//}
	//if token != "" {
	//	c := utils.BuildControlAPIClient(token, utils.DefaultClientID, utils.DefaultAuthURL, utils.DefaultAPIURL)
	//	if kafkaRequest, res, err := c.DefaultApi.GetKafkas(ctx).Execute(); err != nil {
	//		logger.Info("response", "kas", res.Body)
	//		logger.Info("responsea", "kasa", kafkaRequest)
	//
	//		return ctrl.Result{}, err
	//	} else {
	//		logger.Info("response", "kas", res.Body)
	//		logger.Info("responsea", "kasa", kafkaRequest)
	//	}
	//
	//}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TokenReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rhosakkcpv1.Token{}).
		Owns(&v1.Secret{}).
		Complete(r)
}

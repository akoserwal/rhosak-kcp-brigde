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

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kcp-dev/logicalcluster/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/kcp"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	kcpclient "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	rhosakkcpv1 "github.com/akoserwal/rhosak-kcp-brigde/api/v1"
	"github.com/akoserwal/rhosak-kcp-brigde/controllers"

	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	//+kubebuilder:scaffold:imports
)

var (
	scheme            = runtime.NewScheme()
	setupLog          = ctrl.Log.WithName("setup")
	kubeconfigContext string
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(rhosakkcpv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
	flag.StringVar(&kubeconfigContext, "context", "", "kubeconfig context")
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var apiExportName string
	flag.StringVar(&apiExportName, "api-export-name", "", "The name of the APIExport.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}

	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	var err error

	ctx := ctrl.SetupSignalHandler()
	restConfig := ctrl.GetConfigOrDie()
	setupLog = setupLog.WithValues("api-export-name", apiExportName)

	kcpClientConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{}).ClientConfig()
	kcpClient, err := kcpclient.NewClusterForConfig(kcpClientConfig)

	var mgr ctrl.Manager

	options := ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "e4922586.akoserwal",
		LeaderElectionConfig:   restConfig,
	}
	if kcpAPIsGroupPresent(restConfig) {
		setupLog.Info("Looking up virtual workspace URL")
		cfg, err := restConfigForAPIExport(ctx, restConfig, apiExportName)
		if err != nil {
			setupLog.Error(err, "error looking up virtual workspace URL")
		}

		setupLog.Info("Using virtual workspace URL", "url", cfg.Host)

		options.LeaderElectionConfig = restConfig
		mgr, err = kcp.NewClusterAwareManager(cfg, options)
		if err != nil {
			setupLog.Error(err, "unable to start cluster aware manager")
			os.Exit(1)
		}
	} else {
		setupLog.Info("The apis.kcp.dev group is not present - creating standard manager")
		mgr, err = ctrl.NewManager(restConfig, options)
		if err != nil {
			setupLog.Error(err, "unable to start manager")
			os.Exit(1)
		}
	}

	getAPIExport, err := kcpClient.Cluster(logicalcluster.New("root:rhosak")).ApisV1alpha1().APIExports().Get(ctx, apiExportName, metav1.GetOptions{})
	setupLog.Error(err, "Failed to get GLBC APIExport "+apiExportName)
	//+kubebuilder:scaffold:builder
	getAPIExportVirtualWorkspaceURLAndIdentityHash(getAPIExport)

	glbcVWClientConfig := rest.CopyConfig(kcpClientConfig)
	kcpKubeClient, err := kubernetes.NewClusterForConfig(glbcVWClientConfig)

	if err = (&controllers.KafkaInstanceReconciler{
		Client: mgr.GetClient(),

		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KafkaInstance")
		os.Exit(1)
	}
	if err = (&controllers.TokenReconciler{
		Client:     mgr.GetClient(),
		KcpClient:  kcpClient,
		KubeClient: kcpKubeClient,
		Scheme:     mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Token")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// restConfigForAPIExport returns a *rest.Config properly configured to communicate with the endpoint for the
// APIExport's virtual workspace.
func restConfigForAPIExport(ctx context.Context, cfg *rest.Config, apiExportName string) (*rest.Config, error) {
	scheme := runtime.NewScheme()
	if err := apisv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("error adding apis.kcp.dev/v1alpha1 to scheme: %w", err)
	}

	apiExportClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("error creating APIExport client: %w", err)
	}

	var apiExport apisv1alpha1.APIExport

	if apiExportName != "" {
		if err := apiExportClient.Get(ctx, types.NamespacedName{Name: apiExportName}, &apiExport); err != nil {
			return nil, fmt.Errorf("error getting APIExport %q: %w", apiExportName, err)
		}
	} else {
		setupLog.Info("api-export-name is empty - listing")
		exports := &apisv1alpha1.APIExportList{}
		if err := apiExportClient.List(ctx, exports); err != nil {
			return nil, fmt.Errorf("error listing APIExports: %w", err)
		}
		if len(exports.Items) == 0 {
			return nil, fmt.Errorf("no APIExport found")
		}
		if len(exports.Items) > 1 {
			return nil, fmt.Errorf("more than one APIExport found")
		}
		apiExport = exports.Items[0]
	}

	if len(apiExport.Status.VirtualWorkspaces) < 1 {
		return nil, fmt.Errorf("APIExport %q status.virtualWorkspaces is empty", apiExportName)
	}

	cfg = rest.CopyConfig(cfg)
	// TODO(ncdc): sharding support
	cfg.Host = apiExport.Status.VirtualWorkspaces[0].URL

	return cfg, nil
}

func kcpAPIsGroupPresent(restConfig *rest.Config) bool {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		setupLog.Error(err, "failed to create discovery client")
		os.Exit(1)
	}
	apiGroupList, err := discoveryClient.ServerGroups()
	if err != nil {
		setupLog.Error(err, "failed to get server groups")
		os.Exit(1)
	}

	for _, group := range apiGroupList.Groups {
		fmt.Println(group.Name)
		if group.Name == apisv1alpha1.SchemeGroupVersion.Group {
			for _, version := range group.Versions {
				if version.Version == apisv1alpha1.SchemeGroupVersion.Version {
					return true
				}
			}
		}
	}
	return false
}
func exitOnError(err error, msg string) {
	if err != nil {
		os.Exit(1)
	}
}
func getAPIExportVirtualWorkspaceURLAndIdentityHash(export *apisv1alpha1.APIExport) {
	fmt.Println("virtualwork", export.Status.VirtualWorkspaces)

}

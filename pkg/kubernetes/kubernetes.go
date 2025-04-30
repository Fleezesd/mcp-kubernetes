package kubernetes

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type CloseWatchKubeConfig func() error

type Kubernetes struct {
	Kubeconfig                  string
	cfg                         *rest.Config
	clientCmdConfig             clientcmd.ClientConfig
	CloseWatchKubeConfig        CloseWatchKubeConfig
	clientSet                   kubernetes.Interface
	discoveryClient             *discovery.DiscoveryClient
	deferredDiscoveryRESTMapper *restmapper.DeferredDiscoveryRESTMapper
	dynamicClient               *dynamic.DynamicClient
	scheme                      *runtime.Scheme
	parameterCodec              runtime.ParameterCodec
}

func NewKubernetes(kubeconfig string) (*Kubernetes, error) {
	k := &Kubernetes{
		Kubeconfig: kubeconfig,
	}

	if err := k.resolveKubernentesConfigurations(); err != nil {
		return nil, err
	}

	if err := k.initializeClients(); err != nil {
		return nil, err
	}

	if err := k.initializeScheme(); err != nil {
		return nil, err
	}

	return k, nil
}

func (k *Kubernetes) resolveKubernentesConfigurations() error {
	// Initialize config from kubeconfig path if provided
	config, err := k.loadKubeConfig()
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Try in-cluster config if available
	if k.IsInCluster() {
		inClusterCfg, err := InClusterConfig()
		if err == nil && inClusterCfg != nil {
			k.cfg = inClusterCfg
			return nil
		}
	}

	// Fall back to kubeconfig file configuration
	k.cfg = config

	// Set default user agent if not specified
	if k.cfg != nil && k.cfg.UserAgent == "" {
		k.cfg.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// loadKubeConfig handles loading the kubernetes configuration from the kubeconfig file
func (k *Kubernetes) loadKubeConfig() (*rest.Config, error) {
	pathOptions := clientcmd.NewDefaultPathOptions()
	if k.Kubeconfig != "" {
		pathOptions.LoadingRules.ExplicitPath = k.Kubeconfig
	}

	k.clientCmdConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: pathOptions.GetDefaultFilename()},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}},
	)

	return k.clientCmdConfig.ClientConfig()
}

func (k *Kubernetes) initializeClients() error {
	var err error

	k.clientSet, err = kubernetes.NewForConfig(k.cfg)
	if err != nil {
		return fmt.Errorf("failed to create client set: %w", err)
	}

	k.discoveryClient, err = discovery.NewDiscoveryClientForConfig(k.cfg)
	if err != nil {
		return fmt.Errorf("failed to create discovery client: %w", err)
	}

	k.deferredDiscoveryRESTMapper = restmapper.NewDeferredDiscoveryRESTMapper(
		memory.NewMemCacheClient(k.discoveryClient),
	)

	k.dynamicClient, err = dynamic.NewForConfig(k.cfg)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return nil
}

func (k *Kubernetes) initializeScheme() error {
	k.scheme = runtime.NewScheme()
	if err := v1.AddToScheme(k.scheme); err != nil {
		return fmt.Errorf("failed to add v1 to scheme: %w", err)
	}
	k.parameterCodec = runtime.NewParameterCodec(k.scheme)
	return nil
}

func (k *Kubernetes) WatchKubeConfig(onKubeConfigChange func() error) {
	if k.clientCmdConfig == nil {
		return
	}
	kubeConfigFiles := k.clientCmdConfig.ConfigAccess().GetLoadingPrecedence()
	if len(kubeConfigFiles) == 0 {
		return
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	for _, file := range kubeConfigFiles {
		_ = watcher.Add(file)
	}
	go func() {
		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok {
					return
				}
				_ = onKubeConfigChange()
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()
	if k.CloseWatchKubeConfig != nil {
		_ = k.CloseWatchKubeConfig()
	}
	k.CloseWatchKubeConfig = watcher.Close
}

func (k *Kubernetes) Close() {
	if k.CloseWatchKubeConfig != nil {
		_ = k.CloseWatchKubeConfig()
	}
}

func (k *Kubernetes) configuredNamespace() (namespace string) {
	namespace, _, err := k.clientCmdConfig.Namespace()
	if err != nil {
		return ""
	}
	return namespace
}

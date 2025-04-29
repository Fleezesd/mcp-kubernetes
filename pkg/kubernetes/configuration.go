package kubernetes

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/yaml"
)

func (k *Kubernetes) IsInCluster() bool {
	if k.Kubeconfig != "" {
		return false
	}
	cfg, err := InClusterConfig()
	return err == nil && cfg != nil
}

// InClusterConfig is a variable that holds the function to get the in-cluster config
// Exposed for testing
var InClusterConfig = func() (*rest.Config, error) {
	// TODO use kubernetes.default.svc instead of resolved server
	// Currently running into: `http: server gave HTTP response to HTTPS client`
	inClusterConfig, err := rest.InClusterConfig()
	if inClusterConfig != nil {
		inClusterConfig.Host = "https://kubernetes.default.svc"
	}
	return inClusterConfig, err
}

func (k *Kubernetes) ConfigurationView(minify bool) (string, error) {
	var cfg clientcmdapi.Config
	var err error

	if k.IsInCluster() {
		cfg = k.buildInClusterConfig()
	} else if cfg, err = k.clientCmdConfig.RawConfig(); err != nil {
		return "", err
	}

	if minify {
		if err = clientcmdapi.MinifyConfig(&cfg); err != nil {
			return "", err
		}
	}

	convertedObj, err := latest.Scheme.ConvertToVersion(&cfg, latest.ExternalVersion)
	if err != nil {
		return "", err
	}

	return marshal(convertedObj)
}

func (k *Kubernetes) buildInClusterConfig() clientcmdapi.Config {
	cfg := *clientcmdapi.NewConfig()

	cfg.Clusters["cluster"] = &clientcmdapi.Cluster{
		Server:                k.cfg.Host,
		InsecureSkipTLSVerify: k.cfg.Insecure,
	}

	cfg.AuthInfos["user"] = &clientcmdapi.AuthInfo{
		Token: k.cfg.BearerToken,
	}

	cfg.Contexts["context"] = &clientcmdapi.Context{
		Cluster:  "cluster",
		AuthInfo: "user",
	}

	cfg.CurrentContext = "cluster"
	return cfg
}

func marshal(v any) (string, error) {
	switch t := v.(type) {
	case []unstructured.Unstructured:
		for i := range t {
			t[i].SetManagedFields(nil)
		}
	case []*unstructured.Unstructured:
		for i := range t {
			t[i].SetManagedFields(nil)
		}
	case unstructured.Unstructured:
		t.SetManagedFields(nil)
	case *unstructured.Unstructured:
		t.SetManagedFields(nil)
	}
	ret, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

package helmchart

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type HelmRESTClientGetter struct {
	namespace  string
	kubeConfig *string
	restConfig **rest.Config
}

func NewRESTClientGetter(namespace string, kubeConfig *string, restConfig **rest.Config) *HelmRESTClientGetter {
	return &HelmRESTClientGetter{
		namespace:  namespace,
		kubeConfig: kubeConfig,
		restConfig: restConfig,
	}
}

func (c *HelmRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	if c.kubeConfig != nil {
		config, err := clientcmd.RESTConfigFromKubeConfig([]byte(*c.kubeConfig))
		if err != nil {
			return nil, err
		}
		return config, nil
	}
	if c.restConfig != nil {
		return *c.restConfig, nil
	}
	return nil, errors.New("please specify the kubeConfig or restConfig for HelmRESTClientGetter")
}

func (c *HelmRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := c.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	// The more groups you have, the more discovery requests you need to make.
	// given 25 groups (our groups + a few custom conf) with one-ish version each, discovery needs to make 50 requests
	// double it just so we don't end up here again for a while.  This config is only used for discovery.
	config.Burst = 100

	discoveryClient, _ := discovery.NewDiscoveryClientForConfig(config)
	return memory.NewMemCacheClient(discoveryClient), nil
}

func (c *HelmRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}

func (c *HelmRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// use the standard defaults for this client command
	// DEPRECATED: remove and replace with something more accurate
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}
	overrides.Context.Namespace = c.namespace

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
}

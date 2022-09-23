package services

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/viney-shih/go-lock"
	"k8s.io/client-go/informers"
	informerAppsV1 "k8s.io/client-go/informers/apps/v1"
	informerCoreV1 "k8s.io/client-go/informers/core/v1"
	informerNetworkingV1 "k8s.io/client-go/informers/networking/v1"
	listerAppsV1 "k8s.io/client-go/listers/apps/v1"
	listerCoreV1 "k8s.io/client-go/listers/core/v1"
	listerNetworkingV1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/bentoml/yatai/api-server/models"
)

type CacheKey string

var (
	informerSyncTimeout = 30 * time.Second

	informerFactoryCache   = make(map[CacheKey]informers.SharedInformerFactory)
	informerFactoryCacheRW = lock.NewCASMutex()
)

type getSharedInformerFactoryOption struct {
	cluster   *models.Cluster
	namespace *string
}

func getSharedInformerFactory(ctx context.Context, option *getSharedInformerFactoryOption) (informers.SharedInformerFactory, error) {
	org, err := OrganizationService.GetAssociatedOrganization(ctx, option.cluster)
	if err != nil {
		err = errors.Wrap(err, "failed to get associated organization")
		return nil, err
	}
	var cacheKey CacheKey
	if option.namespace != nil {
		cacheKey = CacheKey(fmt.Sprintf("%s:%s:%s", org.Name, option.cluster.Name, *option.namespace))
	} else {
		cacheKey = CacheKey(fmt.Sprintf("%s:%s", org.Name, option.cluster.Name))
	}

	if locked := informerFactoryCacheRW.TryLockWithContext(ctx); !locked {
		return nil, errors.New("failed to get informer factory cache lock")
	}
	defer informerFactoryCacheRW.Unlock()

	var factory informers.SharedInformerFactory
	var ok bool
	if factory, ok = informerFactoryCache[cacheKey]; !ok {
		clientset, _, err := ClusterService.GetKubeCliSet(ctx, option.cluster)
		if err != nil {
			err = errors.Wrap(err, "failed to get kubernetes client set")
			return nil, err
		}
		informerOptions := make([]informers.SharedInformerOption, 0)
		if option.namespace != nil {
			informerOptions = append(informerOptions, informers.WithNamespace(*option.namespace))
		}
		factory = informers.NewSharedInformerFactoryWithOptions(clientset, 0, informerOptions...)
	}

	return factory, nil
}

func startAndSyncInformer(ctx context.Context, informer cache.SharedIndexInformer) (err error) {
	go informer.Run(ctx.Done())

	ctx_, cancel := context.WithTimeout(ctx, informerSyncTimeout)
	defer cancel()

	if !cache.WaitForCacheSync(ctx_.Done(), informer.HasSynced) {
		err = errors.New("Timed out waiting for caches to sync informer")
		return err
	}

	return nil
}

func GetPodInformer(ctx context.Context, cluster *models.Cluster, namespace string) (informerCoreV1.PodInformer, listerCoreV1.PodNamespaceLister, error) {
	factory, err := getSharedInformerFactory(ctx, &getSharedInformerFactoryOption{
		cluster:   cluster,
		namespace: &namespace,
	})
	if err != nil {
		return nil, nil, err
	}
	podInformer := factory.Core().V1().Pods()
	err = startAndSyncInformer(ctx, podInformer.Informer())
	if err != nil {
		return nil, nil, err
	}
	return podInformer, podInformer.Lister().Pods(namespace), nil
}

func GetDeploymentInformer(ctx context.Context, kubeCluster *models.Cluster, namespace string) (informerAppsV1.DeploymentInformer, listerAppsV1.DeploymentNamespaceLister, error) {
	factory, err := getSharedInformerFactory(ctx, &getSharedInformerFactoryOption{
		cluster:   kubeCluster,
		namespace: &namespace,
	})
	if err != nil {
		return nil, nil, err
	}
	deploymentInformer := factory.Apps().V1().Deployments()
	err = startAndSyncInformer(ctx, deploymentInformer.Informer())
	if err != nil {
		return nil, nil, err
	}
	return deploymentInformer, deploymentInformer.Lister().Deployments(namespace), nil
}

func GetStatefulSetInformer(ctx context.Context, kubeCluster *models.Cluster, namespace string) (informerAppsV1.StatefulSetInformer, listerAppsV1.StatefulSetNamespaceLister, error) {
	factory, err := getSharedInformerFactory(ctx, &getSharedInformerFactoryOption{
		cluster:   kubeCluster,
		namespace: &namespace,
	})
	if err != nil {
		return nil, nil, err
	}
	statefulSetInformer := factory.Apps().V1().StatefulSets()
	err = startAndSyncInformer(ctx, statefulSetInformer.Informer())
	if err != nil {
		return nil, nil, err
	}
	return statefulSetInformer, statefulSetInformer.Lister().StatefulSets(namespace), nil
}

func GetIngressInformer(ctx context.Context, kubeCluster *models.Cluster, namespace string) (informerNetworkingV1.IngressInformer, listerNetworkingV1.IngressNamespaceLister, error) {
	factory, err := getSharedInformerFactory(ctx, &getSharedInformerFactoryOption{
		cluster:   kubeCluster,
		namespace: &namespace,
	})
	if err != nil {
		return nil, nil, err
	}
	ingressInformer := factory.Networking().V1().Ingresses()
	err = startAndSyncInformer(ctx, ingressInformer.Informer())
	if err != nil {
		return nil, nil, err
	}
	return ingressInformer, ingressInformer.Lister().Ingresses(namespace), nil
}

func GetDaemonSetInformer(ctx context.Context, kubeCluster *models.Cluster, namespace string) (informerAppsV1.DaemonSetInformer, listerAppsV1.DaemonSetNamespaceLister, error) {
	factory, err := getSharedInformerFactory(ctx, &getSharedInformerFactoryOption{
		cluster:   kubeCluster,
		namespace: &namespace,
	})
	if err != nil {
		return nil, nil, err
	}
	daemonSetInformer := factory.Apps().V1().DaemonSets()
	err = startAndSyncInformer(ctx, daemonSetInformer.Informer())
	if err != nil {
		return nil, nil, err
	}
	return daemonSetInformer, daemonSetInformer.Lister().DaemonSets(namespace), nil
}

func GetEventInformer(ctx context.Context, cluster *models.Cluster, namespace string) (informerCoreV1.EventInformer, listerCoreV1.EventNamespaceLister, error) {
	factory, err := getSharedInformerFactory(ctx, &getSharedInformerFactoryOption{
		cluster:   cluster,
		namespace: &namespace,
	})
	if err != nil {
		return nil, nil, err
	}
	eventInformer := factory.Core().V1().Events()
	err = startAndSyncInformer(ctx, eventInformer.Informer())
	if err != nil {
		return nil, nil, err
	}
	return eventInformer, eventInformer.Lister().Events(namespace), nil
}

func GetNodeEventInformer(ctx context.Context, kubeCluster *models.Cluster) (informerCoreV1.EventInformer, listerCoreV1.EventLister, error) {
	factory, err := getSharedInformerFactory(ctx, &getSharedInformerFactoryOption{
		cluster:   kubeCluster,
		namespace: nil,
	})
	if err != nil {
		return nil, nil, err
	}
	eventInformer := factory.Core().V1().Events()
	err = startAndSyncInformer(ctx, eventInformer.Informer())
	if err != nil {
		return nil, nil, err
	}
	return eventInformer, eventInformer.Lister(), nil
}

func GetSecretInformer(ctx context.Context, kubeCluster *models.Cluster, namespace string) (informerCoreV1.SecretInformer, listerCoreV1.SecretNamespaceLister, error) {
	factory, err := getSharedInformerFactory(ctx, &getSharedInformerFactoryOption{
		cluster:   kubeCluster,
		namespace: &namespace,
	})
	if err != nil {
		return nil, nil, err
	}
	secretInformer := factory.Core().V1().Secrets()
	err = startAndSyncInformer(ctx, secretInformer.Informer())
	if err != nil {
		return nil, nil, err
	}
	return secretInformer, secretInformer.Lister().Secrets(namespace), nil
}

func GetConfigMapInformer(ctx context.Context, kubeCluster *models.Cluster, namespace string) (informerCoreV1.ConfigMapInformer, listerCoreV1.ConfigMapNamespaceLister, error) {
	factory, err := getSharedInformerFactory(ctx, &getSharedInformerFactoryOption{
		cluster:   kubeCluster,
		namespace: &namespace,
	})
	if err != nil {
		return nil, nil, err
	}
	configMapInformer := factory.Core().V1().ConfigMaps()
	err = startAndSyncInformer(ctx, configMapInformer.Informer())
	if err != nil {
		return nil, nil, err
	}
	return configMapInformer, configMapInformer.Lister().ConfigMaps(namespace), nil
}

func GetNodeInformer(ctx context.Context, kubeCluster *models.Cluster) (informerCoreV1.NodeInformer, listerCoreV1.NodeLister, error) {
	factory, err := getSharedInformerFactory(ctx, &getSharedInformerFactoryOption{
		cluster:   kubeCluster,
		namespace: nil,
	})
	if err != nil {
		return nil, nil, err
	}
	nodeInformer := factory.Core().V1().Nodes()
	err = startAndSyncInformer(ctx, nodeInformer.Informer())
	if err != nil {
		return nil, nil, err
	}
	return nodeInformer, nodeInformer.Lister(), nil
}

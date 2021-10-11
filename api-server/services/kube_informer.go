package services

import (
	"context"
	"fmt"
	"time"

	"github.com/bentoml/yatai/api-server/models"

	"github.com/viney-shih/go-lock"

	"github.com/bentoml/yatai/common/utils"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	informerAppsV1 "k8s.io/client-go/informers/apps/v1"
	informerCoreV1 "k8s.io/client-go/informers/core/v1"
	informerNetworkingV1 "k8s.io/client-go/informers/networking/v1"
	listerAppsV1 "k8s.io/client-go/listers/apps/v1"
	listerCoreV1 "k8s.io/client-go/listers/core/v1"
	listerNetworkingV1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
)

type CacheKey string

var (
	informerSyncTimeout = 30 * time.Second

	podInformerCache        = make(map[CacheKey]informerCoreV1.PodInformer)
	podNamespaceListerCache = make(map[CacheKey]listerCoreV1.PodNamespaceLister)
	podInformerMuCache      = make(map[CacheKey]*lock.CASMutex)
	podInformerCacheRW      = lock.NewCASMutex()

	deploymentInformerCache        = make(map[CacheKey]informerAppsV1.DeploymentInformer)
	deploymentNamespaceListerCache = make(map[CacheKey]listerAppsV1.DeploymentNamespaceLister)
	deploymentInformerMuCache      = make(map[CacheKey]*lock.CASMutex)
	deploymentInformerCacheRW      = lock.NewCASMutex()

	statefulSetInformerCache        = make(map[CacheKey]informerAppsV1.StatefulSetInformer)
	statefulSetNamespaceListerCache = make(map[CacheKey]listerAppsV1.StatefulSetNamespaceLister)
	statefulSetInformerMuCache      = make(map[CacheKey]*lock.CASMutex)
	statefulSetInformerCacheRW      = lock.NewCASMutex()

	ingressInformerCache        = make(map[CacheKey]informerNetworkingV1.IngressInformer)
	ingressNamespaceListerCache = make(map[CacheKey]listerNetworkingV1.IngressNamespaceLister)
	ingressInformerMuCache      = make(map[CacheKey]*lock.CASMutex)
	ingressInformerCacheRW      = lock.NewCASMutex()

	daemonSetInformerCache        = make(map[CacheKey]informerAppsV1.DaemonSetInformer)
	daemonSetNamespaceListerCache = make(map[CacheKey]listerAppsV1.DaemonSetNamespaceLister)
	daemonSetInformerMuCache      = make(map[CacheKey]*lock.CASMutex)
	daemonSetInformerCacheRW      = lock.NewCASMutex()

	eventInformerCache        = make(map[CacheKey]informerCoreV1.EventInformer)
	eventNamespaceListerCache = make(map[CacheKey]listerCoreV1.EventNamespaceLister)
	eventInformerMuCache      = make(map[CacheKey]*lock.CASMutex)
	eventInformerCacheRW      = lock.NewCASMutex()

	kubeClusterNodeEventInformerCache   = make(map[CacheKey]informerCoreV1.EventInformer)
	kubeClusterNodeEventListerCache     = make(map[CacheKey]listerCoreV1.EventLister)
	kubeClusterNodeEventInformerMuCache = make(map[CacheKey]*lock.CASMutex)
	kubeClusterNodeEventInformerCacheRW = lock.NewCASMutex()

	secretInformerCache        = make(map[CacheKey]informerCoreV1.SecretInformer)
	secretNamespaceListerCache = make(map[CacheKey]listerCoreV1.SecretNamespaceLister)
	secretInformerMuCache      = make(map[CacheKey]*lock.CASMutex)
	secretInformerCacheRW      = lock.NewCASMutex()

	configMapInformerCache        = make(map[CacheKey]informerCoreV1.ConfigMapInformer)
	configMapNamespaceListerCache = make(map[CacheKey]listerCoreV1.ConfigMapNamespaceLister)
	configMapInformerMuCache      = make(map[CacheKey]*lock.CASMutex)
	configMapInformerCacheRW      = lock.NewCASMutex()

	nodeInformerCache   = make(map[CacheKey]informerCoreV1.NodeInformer)
	nodeListerCache     = make(map[CacheKey]listerCoreV1.NodeLister)
	nodeInformerMuCache = make(map[CacheKey]*lock.CASMutex)
	nodeInformerCacheRW = lock.NewCASMutex()
)

type makeGetInformerOption struct {
	cluster              *models.Cluster
	namespace            *string
	informerCacheRW      *lock.CASMutex
	getInformerFromCache func(cacheKey CacheKey) (interface{}, bool)
	getListerFromCache   func(cacheKey CacheKey) (interface{}, bool)
	setInformerToCache   func(cacheKey CacheKey, v interface{})
	setListerToCache     func(cacheKey CacheKey, v interface{})
	informerMuCache      map[CacheKey]*lock.CASMutex
	informerGetter       func(factory informers.SharedInformerFactory) interface {
		Informer() cache.SharedIndexInformer
	}
	listerGetter func(informer interface{}) interface{}
}

func makeGetInformer(ctx context.Context, option *makeGetInformerOption) (interface{}, interface{}, error) {
	var cacheKey CacheKey
	if option.namespace != nil {
		cacheKey = CacheKey(fmt.Sprintf("%s::%s", option.cluster.Name, *option.namespace))
	} else {
		cacheKey = CacheKey(option.cluster.Name)
	}

	option.informerCacheRW.Lock()
	informer, informerOk := option.getInformerFromCache(cacheKey)
	lister, _ := option.getListerFromCache(cacheKey)
	informerMu, informerMuOk := option.informerMuCache[cacheKey]
	if !informerMuOk {
		informerMu = lock.NewCASMutex()
		option.informerMuCache[cacheKey] = informerMu
	}
	option.informerCacheRW.Unlock()

	if !informerOk {
		var err error
		informer, lister, err = func() (interface{}, interface{}, error) {
			if !informerMu.TryLockWithContext(ctx) {
				return nil, nil, errors.New("informer locker is busy")
			}

			defer informerMu.Unlock()

			option.informerCacheRW.RLock()
			informer, informerOk := option.getInformerFromCache(cacheKey)
			lister, _ := option.getListerFromCache(cacheKey)
			option.informerCacheRW.RUnlock()

			if informerOk {
				return informer, lister, nil
			}

			clientset, _, err := ClusterService.GetKubeCliSet(ctx, option.cluster)
			if err != nil {
				return nil, nil, err
			}

			var factory informers.SharedInformerFactory
			if option.namespace != nil {
				factory = informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithNamespace(*option.namespace))
			} else {
				factory = informers.NewSharedInformerFactoryWithOptions(clientset, 0)
			}
			sthInformer := option.informerGetter(factory)

			stopper := make(chan struct{})
			go factory.Start(stopper)
			defer func() {
				if err != nil {
					close(stopper)
				}
			}()

			syncStopper := make(chan struct{})
			go func() {
				select {
				case <-ctx.Done():
					close(syncStopper)
					return
				case <-time.After(informerSyncTimeout):
					close(syncStopper)
					return
				}
			}()

			informer_ := sthInformer.Informer()
			if !cache.WaitForCacheSync(syncStopper, informer_.HasSynced) {
				err = errors.New("Timed out waiting for caches to sync informer")
				runtime.HandleError(err)
				return nil, nil, err
			}

			lister = option.listerGetter(sthInformer)

			option.informerCacheRW.Lock()
			option.setInformerToCache(cacheKey, sthInformer)
			option.setListerToCache(cacheKey, lister)
			option.informerCacheRW.Unlock()

			return sthInformer, lister, nil
		}()

		if err != nil {
			return nil, nil, err
		}
	}

	return informer, lister, nil
}

func GetPodInformer(ctx context.Context, cluster *models.Cluster, namespace string) (informerCoreV1.PodInformer, listerCoreV1.PodNamespaceLister, error) {
	informer, lister, err := makeGetInformer(ctx, &makeGetInformerOption{
		cluster:         cluster,
		namespace:       utils.StringPtr(namespace),
		informerCacheRW: podInformerCacheRW,
		getInformerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			informer, ok := podInformerCache[cacheKey]
			return informer, ok
		},
		getListerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			lister, ok := podNamespaceListerCache[cacheKey]
			return lister, ok
		},
		setInformerToCache: func(cacheKey CacheKey, v interface{}) {
			podInformerCache[cacheKey] = v.(informerCoreV1.PodInformer)
		},
		setListerToCache: func(cacheKey CacheKey, v interface{}) {
			podNamespaceListerCache[cacheKey] = v.(listerCoreV1.PodNamespaceLister)
		},
		informerMuCache: podInformerMuCache,
		informerGetter: func(factory informers.SharedInformerFactory) interface {
			Informer() cache.SharedIndexInformer
		} {
			return factory.Core().V1().Pods()
		},
		listerGetter: func(informer interface{}) interface{} {
			return informer.(informerCoreV1.PodInformer).Lister().Pods(namespace)
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return informer.(informerCoreV1.PodInformer), lister.(listerCoreV1.PodNamespaceLister), nil
}

func GetDeploymentInformer(ctx context.Context, kubeCluster *models.Cluster, namespace string) (informerAppsV1.DeploymentInformer, listerAppsV1.DeploymentNamespaceLister, error) {
	informer, lister, err := makeGetInformer(ctx, &makeGetInformerOption{
		cluster:         kubeCluster,
		namespace:       utils.StringPtr(namespace),
		informerCacheRW: deploymentInformerCacheRW,
		getInformerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			informer, ok := deploymentInformerCache[cacheKey]
			return informer, ok
		},
		getListerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			lister, ok := deploymentNamespaceListerCache[cacheKey]
			return lister, ok
		},
		setInformerToCache: func(cacheKey CacheKey, v interface{}) {
			deploymentInformerCache[cacheKey] = v.(informerAppsV1.DeploymentInformer)
		},
		setListerToCache: func(cacheKey CacheKey, v interface{}) {
			deploymentNamespaceListerCache[cacheKey] = v.(listerAppsV1.DeploymentNamespaceLister)
		},
		informerMuCache: deploymentInformerMuCache,
		informerGetter: func(factory informers.SharedInformerFactory) interface {
			Informer() cache.SharedIndexInformer
		} {
			return factory.Apps().V1().Deployments()
		},
		listerGetter: func(informer interface{}) interface{} {
			return informer.(informerAppsV1.DeploymentInformer).Lister().Deployments(namespace)
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return informer.(informerAppsV1.DeploymentInformer), lister.(listerAppsV1.DeploymentNamespaceLister), nil
}

func GetStatefulSetInformer(ctx context.Context, kubeCluster *models.Cluster, namespace string) (informerAppsV1.StatefulSetInformer, listerAppsV1.StatefulSetNamespaceLister, error) {
	informer, lister, err := makeGetInformer(ctx, &makeGetInformerOption{
		cluster:         kubeCluster,
		namespace:       utils.StringPtr(namespace),
		informerCacheRW: statefulSetInformerCacheRW,
		getInformerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			informer, ok := statefulSetInformerCache[cacheKey]
			return informer, ok
		},
		getListerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			lister, ok := statefulSetNamespaceListerCache[cacheKey]
			return lister, ok
		},
		setInformerToCache: func(cacheKey CacheKey, v interface{}) {
			statefulSetInformerCache[cacheKey] = v.(informerAppsV1.StatefulSetInformer)
		},
		setListerToCache: func(cacheKey CacheKey, v interface{}) {
			statefulSetNamespaceListerCache[cacheKey] = v.(listerAppsV1.StatefulSetNamespaceLister)
		},
		informerMuCache: statefulSetInformerMuCache,
		informerGetter: func(factory informers.SharedInformerFactory) interface {
			Informer() cache.SharedIndexInformer
		} {
			return factory.Apps().V1().StatefulSets()
		},
		listerGetter: func(informer interface{}) interface{} {
			return informer.(informerAppsV1.StatefulSetInformer).Lister().StatefulSets(namespace)
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return informer.(informerAppsV1.StatefulSetInformer), lister.(listerAppsV1.StatefulSetNamespaceLister), nil
}

func GetIngressInformer(ctx context.Context, kubeCluster *models.Cluster, namespace string) (informerNetworkingV1.IngressInformer, listerNetworkingV1.IngressNamespaceLister, error) {
	informer, lister, err := makeGetInformer(ctx, &makeGetInformerOption{
		cluster:         kubeCluster,
		namespace:       utils.StringPtr(namespace),
		informerCacheRW: ingressInformerCacheRW,
		getInformerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			informer, ok := ingressInformerCache[cacheKey]
			return informer, ok
		},
		getListerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			lister, ok := ingressNamespaceListerCache[cacheKey]
			return lister, ok
		},
		setInformerToCache: func(cacheKey CacheKey, v interface{}) {
			ingressInformerCache[cacheKey] = v.(informerNetworkingV1.IngressInformer)
		},
		setListerToCache: func(cacheKey CacheKey, v interface{}) {
			ingressNamespaceListerCache[cacheKey] = v.(listerNetworkingV1.IngressNamespaceLister)
		},
		informerMuCache: ingressInformerMuCache,
		informerGetter: func(factory informers.SharedInformerFactory) interface {
			Informer() cache.SharedIndexInformer
		} {
			return factory.Networking().V1().Ingresses()
		},
		listerGetter: func(informer interface{}) interface{} {
			return informer.(informerNetworkingV1.IngressInformer).Lister().Ingresses(namespace)
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return informer.(informerNetworkingV1.IngressInformer), lister.(listerNetworkingV1.IngressNamespaceLister), nil
}

func GetDaemonSetInformer(ctx context.Context, kubeCluster *models.Cluster, namespace string) (informerAppsV1.DaemonSetInformer, listerAppsV1.DaemonSetNamespaceLister, error) {
	informer, lister, err := makeGetInformer(ctx, &makeGetInformerOption{
		cluster:         kubeCluster,
		namespace:       utils.StringPtr(namespace),
		informerCacheRW: daemonSetInformerCacheRW,
		getInformerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			informer, ok := daemonSetInformerCache[cacheKey]
			return informer, ok
		},
		getListerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			lister, ok := daemonSetNamespaceListerCache[cacheKey]
			return lister, ok
		},
		setInformerToCache: func(cacheKey CacheKey, v interface{}) {
			daemonSetInformerCache[cacheKey] = v.(informerAppsV1.DaemonSetInformer)
		},
		setListerToCache: func(cacheKey CacheKey, v interface{}) {
			daemonSetNamespaceListerCache[cacheKey] = v.(listerAppsV1.DaemonSetNamespaceLister)
		},
		informerMuCache: daemonSetInformerMuCache,
		informerGetter: func(factory informers.SharedInformerFactory) interface {
			Informer() cache.SharedIndexInformer
		} {
			return factory.Apps().V1().DaemonSets()
		},
		listerGetter: func(informer interface{}) interface{} {
			return informer.(informerAppsV1.DaemonSetInformer).Lister().DaemonSets(namespace)
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return informer.(informerAppsV1.DaemonSetInformer), lister.(listerAppsV1.DaemonSetNamespaceLister), nil
}

func GetEventInformer(ctx context.Context, cluster *models.Cluster, namespace string) (informerCoreV1.EventInformer, listerCoreV1.EventNamespaceLister, error) {
	informer, lister, err := makeGetInformer(ctx, &makeGetInformerOption{
		cluster:         cluster,
		namespace:       utils.StringPtr(namespace),
		informerCacheRW: eventInformerCacheRW,
		getInformerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			informer, ok := eventInformerCache[cacheKey]
			return informer, ok
		},
		getListerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			lister, ok := eventNamespaceListerCache[cacheKey]
			return lister, ok
		},
		setInformerToCache: func(cacheKey CacheKey, v interface{}) {
			eventInformerCache[cacheKey] = v.(informerCoreV1.EventInformer)
		},
		setListerToCache: func(cacheKey CacheKey, v interface{}) {
			eventNamespaceListerCache[cacheKey] = v.(listerCoreV1.EventNamespaceLister)
		},
		informerMuCache: eventInformerMuCache,
		informerGetter: func(factory informers.SharedInformerFactory) interface {
			Informer() cache.SharedIndexInformer
		} {
			return factory.Core().V1().Events()
		},
		listerGetter: func(informer interface{}) interface{} {
			return informer.(informerCoreV1.EventInformer).Lister().Events(namespace)
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return informer.(informerCoreV1.EventInformer), lister.(listerCoreV1.EventNamespaceLister), nil
}

func GetNodeEventInformer(ctx context.Context, kubeCluster *models.Cluster) (informerCoreV1.EventInformer, listerCoreV1.EventLister, error) {
	informer, lister, err := makeGetInformer(ctx, &makeGetInformerOption{
		cluster:         kubeCluster,
		informerCacheRW: kubeClusterNodeEventInformerCacheRW,
		getInformerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			informer, ok := kubeClusterNodeEventInformerCache[cacheKey]
			return informer, ok
		},
		getListerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			lister, ok := kubeClusterNodeEventListerCache[cacheKey]
			return lister, ok
		},
		setInformerToCache: func(cacheKey CacheKey, v interface{}) {
			kubeClusterNodeEventInformerCache[cacheKey] = v.(informerCoreV1.EventInformer)
		},
		setListerToCache: func(cacheKey CacheKey, v interface{}) {
			kubeClusterNodeEventListerCache[cacheKey] = v.(listerCoreV1.EventLister)
		},
		informerMuCache: kubeClusterNodeEventInformerMuCache,
		informerGetter: func(factory informers.SharedInformerFactory) interface {
			Informer() cache.SharedIndexInformer
		} {
			return factory.Core().V1().Events()
		},
		listerGetter: func(informer interface{}) interface{} {
			return informer.(informerCoreV1.EventInformer).Lister()
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return informer.(informerCoreV1.EventInformer), lister.(listerCoreV1.EventLister), nil
}

func GetSecretInformer(ctx context.Context, kubeCluster *models.Cluster, namespace string) (informerCoreV1.SecretInformer, listerCoreV1.SecretNamespaceLister, error) {
	informer, lister, err := makeGetInformer(ctx, &makeGetInformerOption{
		cluster:         kubeCluster,
		namespace:       utils.StringPtr(namespace),
		informerCacheRW: secretInformerCacheRW,
		getInformerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			informer, ok := secretInformerCache[cacheKey]
			return informer, ok
		},
		getListerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			lister, ok := secretNamespaceListerCache[cacheKey]
			return lister, ok
		},
		setInformerToCache: func(cacheKey CacheKey, v interface{}) {
			secretInformerCache[cacheKey] = v.(informerCoreV1.SecretInformer)
		},
		setListerToCache: func(cacheKey CacheKey, v interface{}) {
			secretNamespaceListerCache[cacheKey] = v.(listerCoreV1.SecretNamespaceLister)
		},
		informerMuCache: secretInformerMuCache,
		informerGetter: func(factory informers.SharedInformerFactory) interface {
			Informer() cache.SharedIndexInformer
		} {
			return factory.Core().V1().Secrets()
		},
		listerGetter: func(informer interface{}) interface{} {
			return informer.(informerCoreV1.SecretInformer).Lister().Secrets(namespace)
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return informer.(informerCoreV1.SecretInformer), lister.(listerCoreV1.SecretNamespaceLister), nil
}

func GetConfigMapInformer(ctx context.Context, kubeCluster *models.Cluster, namespace string) (informerCoreV1.ConfigMapInformer, listerCoreV1.ConfigMapNamespaceLister, error) {
	informer, lister, err := makeGetInformer(ctx, &makeGetInformerOption{
		cluster:         kubeCluster,
		namespace:       utils.StringPtr(namespace),
		informerCacheRW: configMapInformerCacheRW,
		getInformerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			informer, ok := configMapInformerCache[cacheKey]
			return informer, ok
		},
		getListerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			lister, ok := configMapNamespaceListerCache[cacheKey]
			return lister, ok
		},
		setInformerToCache: func(cacheKey CacheKey, v interface{}) {
			configMapInformerCache[cacheKey] = v.(informerCoreV1.ConfigMapInformer)
		},
		setListerToCache: func(cacheKey CacheKey, v interface{}) {
			configMapNamespaceListerCache[cacheKey] = v.(listerCoreV1.ConfigMapNamespaceLister)
		},
		informerMuCache: configMapInformerMuCache,
		informerGetter: func(factory informers.SharedInformerFactory) interface {
			Informer() cache.SharedIndexInformer
		} {
			return factory.Core().V1().ConfigMaps()
		},
		listerGetter: func(informer interface{}) interface{} {
			return informer.(informerCoreV1.ConfigMapInformer).Lister().ConfigMaps(namespace)
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return informer.(informerCoreV1.ConfigMapInformer), lister.(listerCoreV1.ConfigMapNamespaceLister), nil
}

func GetNodeInformer(ctx context.Context, kubeCluster *models.Cluster) (informerCoreV1.NodeInformer, listerCoreV1.NodeLister, error) {
	informer, lister, err := makeGetInformer(ctx, &makeGetInformerOption{
		cluster:         kubeCluster,
		informerCacheRW: nodeInformerCacheRW,
		getInformerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			informer, ok := nodeInformerCache[cacheKey]
			return informer, ok
		},
		getListerFromCache: func(cacheKey CacheKey) (interface{}, bool) {
			lister, ok := nodeListerCache[cacheKey]
			return lister, ok
		},
		setInformerToCache: func(cacheKey CacheKey, v interface{}) {
			nodeInformerCache[cacheKey] = v.(informerCoreV1.NodeInformer)
		},
		setListerToCache: func(cacheKey CacheKey, v interface{}) {
			nodeListerCache[cacheKey] = v.(listerCoreV1.NodeLister)
		},
		informerMuCache: nodeInformerMuCache,
		informerGetter: func(factory informers.SharedInformerFactory) interface {
			Informer() cache.SharedIndexInformer
		} {
			return factory.Core().V1().Nodes()
		},
		listerGetter: func(informer interface{}) interface{} {
			return informer.(informerCoreV1.NodeInformer).Lister()
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return informer.(informerCoreV1.NodeInformer), lister.(listerCoreV1.NodeLister), nil
}

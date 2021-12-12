package controllersv1

import (
	"sort"
	"sync/atomic"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bentoml/yatai/common/consts"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type kubeController struct {
	baseController
}

var KubeController = kubeController{}

func (c *kubeController) GetPodKubeEvents(ctx *gin.Context, schema *GetClusterSchema) error {
	var err error

	ctx.Request.Header.Del("Origin")
	conn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logrus.Errorf("ws connect failed: %q", err.Error())
		return err
	}
	defer conn.Close()

	defer func() {
		if err != nil {
			msg := schemasv1.WsRespSchema{
				Type:    schemasv1.WsRespTypeError,
				Message: err.Error(),
				Payload: nil,
			}
			_ = conn.WriteJSON(&msg)
		}
	}()

	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return err
	}

	err = ClusterController.canView(ctx, cluster)
	if err != nil {
		return err
	}

	closeCh := make(chan struct{})
	go func() {
		for {
			mt, _, _ := conn.ReadMessage()

			if mt == websocket.CloseMessage || mt == -1 {
				close(closeCh)
				break
			}
		}
	}()

	filter := func(event *apiv1.Event) bool {
		return true
	}

	kubeNs := ctx.Query("namespace")
	podName := ctx.Query("pod_name")
	if podName != "" {
		cliset, _, err := services.ClusterService.GetKubeCliSet(ctx, cluster)
		if err != nil {
			return err
		}

		podsCli := cliset.CoreV1().Pods(kubeNs)

		pod, err := podsCli.Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		filter = func(event *apiv1.Event) bool {
			return event.InvolvedObject.Kind == consts.KubeEventResourceKindPod && event.InvolvedObject.UID == pod.UID
		}
	}

	eventInformer, eventLister, err := services.GetEventInformer(ctx, cluster, kubeNs)
	if err != nil {
		err = errors.Wrap(err, "get event informer")
		return err
	}

	informer := eventInformer.Informer()
	defer runtime.HandleCrash()

	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	var failedCount int32 = 0
	var maxFailed int32 = 10

	failed := func() {
		atomic.AddInt32(&failedCount, 1)
		time.Sleep(time.Second * 10)
	}

	send := func() {
		events, err := eventLister.List(labels.Everything())
		if err != nil {
			logrus.Errorf("list events failed: %s", err.Error())
			failed()
			return
		}

		_events := make([]*apiv1.Event, 0)

		for _, event := range events {
			if !filter(event) {
				continue
			}
			_events = append(_events, event)
		}

		sort.SliceStable(_events, func(i, j int) bool {
			it := _events[i].LastTimestamp
			jt := _events[j].LastTimestamp

			return it.Before(&jt)
		})

		err = conn.WriteJSON(&schemasv1.WsRespSchema{
			Type:    schemasv1.WsRespTypeSuccess,
			Message: "",
			Payload: _events,
		})
		if err != nil {
			logrus.Errorf("ws write json failed: %s", err.Error())
			failed()
			return
		}
	}

	send()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if event, ok := obj.(*apiv1.Event); ok {
				if !filter(event) {
					return
				}
				send()
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if event, ok := newObj.(*apiv1.Event); ok {
				if !filter(event) {
					return
				}
				send()
			}
		},
		DeleteFunc: func(obj interface{}) {
			if event, ok := obj.(*apiv1.Event); ok {
				if !filter(event) {
					return
				}
				send()
			}
		},
	})

	func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		for {
			select {
			case <-closeCh:
				return
			default:
			}

			if failedCount > maxFailed {
				logrus.Error("ws pods failed too frequently!")
				break
			}

			<-ticker.C
		}
	}()

	return nil
}

func (c *kubeController) GetDeploymentKubeEvents(ctx *gin.Context, schema *GetDeploymentSchema) error {
	var err error

	ctx.Request.Header.Del("Origin")
	conn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logrus.Errorf("ws connect failed: %q", err.Error())
		return err
	}
	defer conn.Close()

	defer func() {
		if err != nil {
			msg := schemasv1.WsRespSchema{
				Type:    schemasv1.WsRespTypeError,
				Message: err.Error(),
				Payload: nil,
			}
			_ = conn.WriteJSON(&msg)
		}
	}()

	deployment, err := schema.GetDeployment(ctx)
	if err != nil {
		return err
	}

	err = DeploymentController.canView(ctx, deployment)
	if err != nil {
		return err
	}

	closeCh := make(chan struct{})
	go func() {
		for {
			mt, _, _ := conn.ReadMessage()

			if mt == websocket.CloseMessage || mt == -1 {
				close(closeCh)
				break
			}
		}
	}()

	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return err
	}

	filter, err := services.KubeEventService.MakeDeploymentKubeEventFilter(ctx, deployment, nil)
	if err != nil {
		return err
	}

	podName := ctx.Query("pod_name")
	if podName != "" {
		cliset, _, err := services.ClusterService.GetKubeCliSet(ctx, cluster)
		if err != nil {
			return err
		}

		kubeNs := services.DeploymentService.GetKubeNamespace(deployment)
		podsCli := cliset.CoreV1().Pods(kubeNs)

		pod, err := podsCli.Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if pod.Labels[consts.KubeLabelYataiDeployment] != deployment.Name {
			return errors.Errorf("pod %s not in this deployment", podName)
		}

		filter = func(event *apiv1.Event) bool {
			return event.InvolvedObject.Kind == consts.KubeEventResourceKindPod && event.InvolvedObject.UID == pod.UID
		}
	}

	eventInformer, eventLister, err := services.GetEventInformer(ctx, cluster, services.DeploymentService.GetKubeNamespace(deployment))
	if err != nil {
		err = errors.Wrap(err, "get event informer")
		return err
	}

	informer := eventInformer.Informer()
	defer runtime.HandleCrash()

	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	var failedCount int32 = 0
	var maxFailed int32 = 10

	failed := func() {
		atomic.AddInt32(&failedCount, 1)
		time.Sleep(time.Second * 10)
	}

	send := func() {
		events, err := eventLister.List(labels.Everything())
		if err != nil {
			logrus.Errorf("list events failed: %s", err.Error())
			failed()
			return
		}

		_events := make([]*apiv1.Event, 0)

		for _, event := range events {
			if !filter(event) {
				continue
			}
			_events = append(_events, event)
		}

		sort.SliceStable(_events, func(i, j int) bool {
			it := _events[i].LastTimestamp
			jt := _events[j].LastTimestamp

			return it.Before(&jt)
		})

		err = conn.WriteJSON(&schemasv1.WsRespSchema{
			Type:    schemasv1.WsRespTypeSuccess,
			Message: "",
			Payload: _events,
		})
		if err != nil {
			logrus.Errorf("ws write json failed: %s", err.Error())
			failed()
			return
		}
	}

	send()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if event, ok := obj.(*apiv1.Event); ok {
				if !filter(event) {
					return
				}
				send()
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if event, ok := newObj.(*apiv1.Event); ok {
				if !filter(event) {
					return
				}
				send()
			}
		},
		DeleteFunc: func(obj interface{}) {
			if event, ok := obj.(*apiv1.Event); ok {
				if !filter(event) {
					return
				}
				send()
			}
		},
	})

	func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		for {
			select {
			case <-closeCh:
				return
			default:
			}

			if failedCount > maxFailed {
				logrus.Error("ws pods failed too frequently!")
				break
			}

			<-ticker.C
		}
	}()

	return nil
}

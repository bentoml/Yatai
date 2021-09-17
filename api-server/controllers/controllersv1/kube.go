package controllersv1

import (
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type kubeController struct {
	baseController
}

var KubeController = kubeController{}

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

	eventInformer, eventLister, err := services.GetEventInformer(ctx, cluster, services.DeploymentService.GetKubeNamespace(deployment))
	if err != nil {
		err = errors.Wrap(err, "get app pool event informer")
		return err
	}

	filter, err := services.KubeEventService.MakeKubeEventFilter(ctx, deployment, nil)
	if err != nil {
		return err
	}

	informer := eventInformer.Informer()
	defer runtime.HandleCrash()

	seen := make(map[types.UID]struct{})
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()
	returned := false

	var failedCount int32 = 0
	var maxFailed int32 = 10

	failed := func() {
		atomic.AddInt32(&failedCount, 1)
		time.Sleep(time.Second * 10)
	}

	send := func() {
		events, err := eventLister.List(labels.Everything())
		if err != nil {
			logrus.Errorf("list events failed: %s", errors.Wrap(err, "list events from app pool event informer").Error())
			failed()
			return
		}

		for _, event := range events {
			if !filter(event) {
				continue
			}
			if _, ok := seen[event.UID]; ok {
				continue
			}
			seen[event.UID] = struct{}{}
			err := conn.WriteJSON(&schemasv1.WsRespSchema{
				Type:    schemasv1.WsRespTypeSuccess,
				Message: "",
				Payload: event,
			})
			if err != nil {
				logrus.Errorf("ws write json failed: %s", err.Error())
				failed()
				continue
			}
		}

		if !returned && len(events) == 0 {
			returned = true
			err := conn.WriteJSON(&schemasv1.WsRespSchema{
				Type:    schemasv1.WsRespTypeSuccess,
				Message: "",
				Payload: nil,
			})
			if err != nil {
				logrus.Errorf("ws write json failed: %s", err.Error())
				failed()
				return
			}
		}
	}

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

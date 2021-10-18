package controllersv1

import (
	"context"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"helm.sh/helm/v3/pkg/chart"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/bentoml/yatai/common/consts"

	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/schemas/modelschemas"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type yataiComponentController struct {
	clusterController
}

var YataiComponentController = yataiComponentController{}

func (c *yataiComponentController) ListOperatorHelmCharts(ctx *gin.Context) ([]*chart.Chart, error) {
	return services.YataiComponentService.ListOperatorHelmCharts(ctx)
}

func (c *yataiComponentController) List(ctx *gin.Context, schema *GetClusterSchema) ([]*schemasv1.YataiComponentSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, cluster); err != nil {
		return nil, err
	}
	comps, err := services.YataiComponentService.List(ctx, cluster.ID)
	if err != nil {
		return nil, errors.Wrap(err, "list cluster yatai comps")
	}
	return transformersv1.ToYataiComponentSchemas(ctx, comps)
}

type CreateYataiComponentSchema struct {
	schemasv1.CreateYataiComponentSchema
	GetClusterSchema
}

func (c *yataiComponentController) Create(ctx *gin.Context, schema *CreateYataiComponentSchema) (*schemasv1.YataiComponentSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canOperate(ctx, cluster); err != nil {
		return nil, err
	}
	comp, err := services.YataiComponentService.Create(ctx, services.CreateYataiComponentReleaseOption{
		ClusterId: cluster.ID,
		Type:      schema.Type,
	})
	if err != nil {
		return nil, err
	}
	err = clearGrafanaCache(ctx, schema.OrgName, schema.ClusterName)
	if err != nil {
		return nil, errors.Wrap(err, "clear grafana cache")
	}
	return transformersv1.ToYataiComponentSchema(ctx, comp)
}

type GetYataiComponentSchema struct {
	GetClusterSchema
	Type modelschemas.YataiComponentType `path:"componentType"`
}

func (c *yataiComponentController) Get(ctx *gin.Context, schema *GetYataiComponentSchema) (*schemasv1.YataiComponentSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get cluster")
	}
	if err = c.canOperate(ctx, cluster); err != nil {
		return nil, err
	}
	comp, err := services.YataiComponentService.Get(ctx, services.GetYataiComponentReleaseOption{
		ClusterId: cluster.ID,
		Type:      schema.Type,
	})
	if err != nil {
		return nil, errors.Wrap(err, "delete yatai component")
	}
	return transformersv1.ToYataiComponentSchema(ctx, comp)
}

func (c *yataiComponentController) Delete(ctx *gin.Context, schema *GetYataiComponentSchema) (*schemasv1.YataiComponentSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get cluster")
	}
	if err = c.canOperate(ctx, cluster); err != nil {
		return nil, err
	}
	comp, err := services.YataiComponentService.Delete(ctx, services.DeleteYataiComponentReleaseOption{
		ClusterId: cluster.ID,
		Type:      schema.Type,
	})
	if err != nil {
		return nil, errors.Wrap(err, "delete yatai component")
	}
	err = clearGrafanaCache(ctx, schema.OrgName, schema.ClusterName)
	if err != nil {
		return nil, errors.Wrap(err, "clear grafana cache")
	}
	return transformersv1.ToYataiComponentSchema(ctx, comp)
}

func (c *yataiComponentController) ListHelmChartReleaseResources(ctx *gin.Context, schema *GetYataiComponentSchema) (err error) {
	ctx.Request.Header.Del("Origin")
	conn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logrus.Errorf("ws connect failed: %q", err.Error())
		return
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
		err = errors.Wrap(err, "get cluster")
		return
	}

	if err = c.canOperate(ctx, cluster); err != nil {
		return
	}

	pollingCtx, cancel := context.WithCancel(ctx)

	go func() {
		for {
			mt, _, err := conn.ReadMessage()

			if err != nil || mt == websocket.CloseMessage || mt == -1 {
				cancel()
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logrus.Printf("error: %v", err)
				}
				break
			}
		}
	}()

	var previousData []*schemasv1.KubeResourceSchema

	send_ := func() error {
		resources, err := services.YataiComponentService.ListHelmChartReleaseResources(ctx, services.ListYataiComponentHelmChartReleaseResourcesOption{
			ClusterId: cluster.ID,
			Type:      schema.Type,
		})
		if err != nil {
			err = errors.Wrap(err, "list yatai component helm chart release resources")
			return err
		}

		ss, err := transformersv1.ToKubeResourceSchemas(ctx, resources)
		if err != nil {
			return err
		}

		if reflect.DeepEqual(ss, previousData) {
			return nil
		}

		previousData = ss
		err = conn.WriteJSON(schemasv1.WsRespSchema{
			Type:    schemasv1.WsRespTypeSuccess,
			Message: "",
			Payload: ss,
		})
		return err
	}

	failedCount := atomic.NewInt64(0)
	maxFailed := int64(10)

	fail := func() {
		failedCount.Inc()
	}

	send := func() {
		err := send_()
		if err != nil {
			fail()
		}
	}

	secretInformer, _, err := services.GetSecretInformer(ctx, cluster, consts.KubeNamespaceYataiComponents)
	if err != nil {
		return
	}

	selector, err := labels.Parse("owner=helm")
	if err != nil {
		return
	}

	informer := secretInformer.Informer()
	defer runtime.HandleCrash()

	checkSecret := func(obj interface{}) bool {
		secret, ok := obj.(*apiv1.Secret)
		if !ok {
			return false
		}
		return selector.Matches(labels.Set(secret.Labels))
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if !checkSecret(obj) {
				return
			}
			send()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !checkSecret(newObj) {
				return
			}
			send()
		},
		DeleteFunc: func(obj interface{}) {
			if !checkSecret(obj) {
				return
			}
			send()
		},
	})

	err = send_()
	if err != nil {
		return
	}

	func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		for {
			select {
			case <-pollingCtx.Done():
				return
			default:
			}

			if failedCount.Load() > maxFailed {
				logrus.Error("ws pods failed too frequently!")
				break
			}

			<-ticker.C
		}
	}()

	return nil
}

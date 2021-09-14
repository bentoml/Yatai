package controllersv1

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/sync/errsgroup"
	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type deploymentController struct {
	baseController
}

var DeploymentController = deploymentController{}

type GetDeploymentSchema struct {
	GetClusterSchema
	DeploymentName string `path:"deploymentName"`
}

func (s *GetDeploymentSchema) GetDeployment(ctx context.Context) (*models.Deployment, error) {
	cluster, err := s.GetCluster(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get cluster")
	}
	deployment, err := services.DeploymentService.GetByName(ctx, cluster.ID, s.DeploymentName)
	if err != nil {
		return nil, errors.Wrapf(err, "get deployment %s", s.DeploymentName)
	}
	return deployment, nil
}

func (c *deploymentController) canView(ctx context.Context, deployment *models.Deployment) error {
	cluster, err := services.ClusterService.GetAssociatedCluster(ctx, deployment)
	if err != nil {
		return errors.Wrap(err, "get associated cluster")
	}
	return ClusterController.canView(ctx, cluster)
}

func (c *deploymentController) canUpdate(ctx context.Context, deployment *models.Deployment) error {
	cluster, err := services.ClusterService.GetAssociatedCluster(ctx, deployment)
	if err != nil {
		return errors.Wrap(err, "get associated cluster")
	}
	return ClusterController.canUpdate(ctx, cluster)
}

func (c *deploymentController) canOperate(ctx context.Context, deployment *models.Deployment) error {
	cluster, err := services.ClusterService.GetAssociatedCluster(ctx, deployment)
	if err != nil {
		return errors.Wrap(err, "get associated cluster")
	}
	return ClusterController.canOperate(ctx, cluster)
}

type CreateDeploymentSchema struct {
	schemasv1.CreateDeploymentSchema
	GetClusterSchema
}

func (c *deploymentController) Create(ctx *gin.Context, schema *CreateDeploymentSchema) (*schemasv1.DeploymentSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = ClusterController.canUpdate(ctx, cluster); err != nil {
		return nil, err
	}

	deployment, err := services.DeploymentService.Create(ctx, services.CreateDeploymentOption{
		CreatorId:   user.ID,
		ClusterId:   cluster.ID,
		Name:        schema.Name,
		Description: schema.Description,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create deployment")
	}

	bento, err := services.BentoService.GetByName(ctx, org.ID, schema.BentoName)
	if err != nil {
		return nil, errors.Wrapf(err, "get bento %s from organization %s", schema.BentoName, org.Name)
	}

	bentoVersion, err := services.BentoVersionService.GetByVersion(ctx, bento.ID, schema.BentoVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "get bento %s version %s from organization %s", schema.BentoName, schema.BentoVersion, org.Name)
	}

	deploymentSnapshot, err := services.DeploymentSnapshotService.Create(ctx, services.CreateDeploymentSnapshotOption{
		CreatorId:      user.ID,
		DeploymentId:   deployment.ID,
		BentoVersionId: bentoVersion.ID,
		Type:           schema.Type,
		Status:         modelschemas.DeploymentSnapshotStatusActive,
		CanaryRules:    schema.CanaryRules,
		Config:         schema.Config,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create deployment snapshot")
	}

	err = services.DeploymentSnapshotService.Deploy(ctx, deploymentSnapshot, false)
	if err != nil {
		return nil, errors.Wrap(err, "deploy deployment snapshot")
	}

	return transformersv1.ToDeploymentSchema(ctx, deployment)
}

type UpdateDeploymentSchema struct {
	schemasv1.UpdateDeploymentSchema
	GetDeploymentSchema
}

func (c *deploymentController) Update(ctx *gin.Context, schema *UpdateDeploymentSchema) (*schemasv1.DeploymentSchema, error) {
	deployment, err := schema.GetDeployment(ctx)
	if err != nil {
		return nil, err
	}
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, deployment); err != nil {
		return nil, err
	}

	bento, err := services.BentoService.GetByName(ctx, org.ID, schema.BentoName)
	if err != nil {
		return nil, errors.Wrapf(err, "get bento %s from organization %s", schema.BentoName, org.Name)
	}

	bentoVersion, err := services.BentoVersionService.GetByVersion(ctx, bento.ID, schema.BentoVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "get bento %s version %s from organization %s", schema.BentoName, schema.BentoVersion, org.Name)
	}

	deploymentSnapshot, err := services.DeploymentSnapshotService.Create(ctx, services.CreateDeploymentSnapshotOption{
		CreatorId:      user.ID,
		DeploymentId:   deployment.ID,
		BentoVersionId: bentoVersion.ID,
		Type:           schema.Type,
		Status:         modelschemas.DeploymentSnapshotStatusActive,
		CanaryRules:    schema.CanaryRules,
		Config:         schema.Config,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create deployment snapshot")
	}

	err = services.DeploymentSnapshotService.Deploy(ctx, deploymentSnapshot, false)
	if err != nil {
		return nil, errors.Wrap(err, "deploy deployment snapshot")
	}

	return transformersv1.ToDeploymentSchema(ctx, deployment)
}

func (c *deploymentController) Get(ctx *gin.Context, schema *GetDeploymentSchema) (*schemasv1.DeploymentSchema, error) {
	deployment, err := schema.GetDeployment(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, deployment); err != nil {
		return nil, err
	}
	return transformersv1.ToDeploymentSchema(ctx, deployment)
}

type ListDeploymentSchema struct {
	schemasv1.ListQuerySchema
	GetClusterSchema
}

func (c *deploymentController) List(ctx *gin.Context, schema *ListDeploymentSchema) (*schemasv1.DeploymentListSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}

	if err = ClusterController.canView(ctx, cluster); err != nil {
		return nil, err
	}

	deployments, total, err := services.DeploymentService.List(ctx, services.ListDeploymentOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		ClusterId: cluster.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "list deployments")
	}

	deploymentSchemas, err := transformersv1.ToDeploymentSchemas(ctx, deployments)
	return &schemasv1.DeploymentListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: deploymentSchemas,
	}, err
}

type ListTerminalRecordSchema struct {
	schemasv1.ListQuerySchema
	GetDeploymentSchema
}

func (c *deploymentController) ListTerminalRecords(ctx *gin.Context, schema *ListTerminalRecordSchema) (*schemasv1.TerminalRecordListSchema, error) {
	deployment, err := schema.GetDeployment(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, deployment); err != nil {
		return nil, err
	}

	terminalRecords, total, err := services.TerminalRecordService.List(ctx, services.ListTerminalRecordOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		DeploymentId: utils.UintPtr(deployment.ID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list terminal records")
	}

	terminalRecordSchemas, err := transformersv1.ToTerminalRecordSchemas(ctx, terminalRecords)
	return &schemasv1.TerminalRecordListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: terminalRecordSchemas,
	}, err
}

var (
	deploymentPodsWsConns       sync.Map
	deploymentPodsWsConnRws     = make(map[string]*sync.RWMutex)
	deploymentPodsWsHasManagers = make(map[string]bool)
	deploymentPodsWsConnRwsRw   sync.RWMutex
)

type connWrapper struct {
	Conn     *websocket.Conn
	IsNew    bool
	IsClosed bool
}

func (c *deploymentController) WsPods(ctx *gin.Context, schema *GetDeploymentSchema) error {
	ctx.Request.Header.Del("Origin")
	conn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logrus.Errorf("ws connect failed: %q", err.Error())
		return err
	}
	defer conn.Close()

	deployment, err := schema.GetDeployment(ctx)
	if err != nil {
		return err
	}
	if err = c.canView(ctx, deployment); err != nil {
		return err
	}

	cachedKey := fmt.Sprintf("%d", deployment.ID)

	deploymentPodsWsConnRwsRw.Lock()
	rw := deploymentPodsWsConnRws[cachedKey]
	if rw == nil {
		rw = &sync.RWMutex{}
	}
	deploymentPodsWsConnRws[cachedKey] = rw
	deploymentPodsWsConnRwsRw.Unlock()

	rw.Lock()
	conns := make([]*connWrapper, 0)
	conns_, ok := deploymentPodsWsConns.Load(cachedKey)
	if ok {
		conns = conns_.([]*connWrapper)
	}
	connW := &connWrapper{
		Conn:     conn,
		IsNew:    false,
		IsClosed: false,
	}
	conns = append(conns, connW)
	deploymentPodsWsConns.Store(cachedKey, conns)
	rw.Unlock()

	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return err
	}

	podInformer, podLister, err := services.GetPodInformer(ctx, cluster, consts.KubeNamespaceYataiDeployment)
	if err != nil {
		return err
	}

	pods, err := services.KubePodService.ListPodsByDeployment(ctx, podLister, deployment)
	if err != nil {
		return err
	}

	var podSchemas []*schemasv1.KubePodSchema

	podSchemas, err = transformersv1.ToPodSchemas(ctx, pods)
	if err != nil {
		err = errors.Wrap(err, "get app all components with pods")
		return err
	}

	err = connW.Conn.WriteJSON(&schemasv1.WsRespSchema{
		Type:    schemasv1.WsRespTypeSuccess,
		Message: "",
		Payload: podSchemas,
	})
	if err != nil {
		logrus.Errorf("ws write json failed: %q", err.Error())
	}
	connW.IsNew = false

	pollingCtx, cancel := context.WithCancel(ctx)
	go func() {
		for {
			mt, _, err := conn.ReadMessage()

			if err != nil || mt == websocket.CloseMessage || mt == -1 {
				connW.IsClosed = true
				cancel()
				break
			}
		}
	}()

	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-pollingCtx.Done():
			return nil
		default:
		}

		rw.RLock()
		hasManager := deploymentPodsWsHasManagers[cachedKey]
		rw.RUnlock()

		if hasManager {
			<-ticker.C
		} else {
			break
		}
	}

	rw.Lock()
	deploymentPodsWsHasManagers[cachedKey] = true
	defer func() {
		rw.Lock()
		defer rw.Unlock()
		deploymentPodsWsHasManagers[cachedKey] = false
	}()
	rw.Unlock()

	failedCount := 0
	maxFailed := 10

	failed := func() {
		failedCount += 1
		time.Sleep(time.Second * 10)
	}

	send := func(podLister v1.PodNamespaceLister) error {
		rw.Lock()
		defer rw.Unlock()

		conns := make([]*connWrapper, 0)
		conns_, ok := deploymentPodsWsConns.Load(cachedKey)
		if ok {
			conns = conns_.([]*connWrapper)
		}

		newConns := make([]*connWrapper, 0, len(conns))

		pods, err := services.KubePodService.ListPodsByDeployment(ctx, podLister, deployment)
		if err != nil {
			logrus.Errorf("get app pods failed: %q", err.Error())
			failed()
			return err
		}

		newPodSchemas, err := transformersv1.ToPodSchemas(ctx, pods)
		if err != nil {
			logrus.Errorf("get app pods failed: %q", err.Error())
			failed()
			return err
		}

		viewChanged := !reflect.DeepEqual(podSchemas, newPodSchemas)
		podSchemas = newPodSchemas

		var mu sync.Mutex
		var eg errsgroup.Group
		for _, conn := range conns {
			conn := conn

			if conn.IsClosed {
				continue
			}

			if !conn.IsNew && !viewChanged {
				newConns = append(newConns, conn)
				continue
			}

			eg.Go(func() error {
				err = conn.Conn.WriteJSON(&schemasv1.WsRespSchema{
					Type:    schemasv1.WsRespTypeSuccess,
					Message: "",
					Payload: newPodSchemas,
				})
				if err != nil {
					_ = conn.Conn.Close()
					conn.IsClosed = true
				} else {
					mu.Lock()
					conn.IsNew = false
					newConns = append(newConns, conn)
					mu.Unlock()
				}
				return nil
			})
		}
		err = eg.Wait()
		if err != nil {
			logrus.Errorf("eg wait failed: %q", err.Error())
			return err
		}
		deploymentPodsWsConns.Store(cachedKey, newConns)
		failedCount = 0
		return nil
	}

	send_ := func() {
		_ = send(podLister)
	}

	informer := podInformer.Informer()
	defer runtime.HandleCrash()

	deploymentId := fmt.Sprintf("%d", deployment.ID)

	checkPod := func(obj interface{}) bool {
		pod, ok := obj.(*apiv1.Pod)
		if !ok {
			return false
		}
		if pod.Labels[consts.KubeLabelYataiDeploymentId] != deploymentId {
			return false
		}
		return true
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if !checkPod(obj) {
				return
			}
			send_()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !checkPod(newObj) {
				return
			}
			send_()
		},
		DeleteFunc: func(obj interface{}) {
			if !checkPod(obj) {
				return
			}
			send_()
		},
	})

	func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		for {
			select {
			case <-pollingCtx.Done():
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

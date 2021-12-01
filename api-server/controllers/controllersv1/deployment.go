package controllersv1

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/huandu/xstrings"

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

	return c.doUpdate(ctx, schema.UpdateDeploymentSchema, org, deployment, label)
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
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, deployment); err != nil {
		return nil, err
	}
	label, err := services.LabelService.Get(ctx, schema.UpdateDeploymentSchema)
	if err = LabelController.canUpdate(ctx, label); err != nil {
		return nil, err
	}
	if label, err := services.LabelService.Create(ctx, services.CreateLabelOption{
		OrganizationId: org.ID,
		CreatorId:      user.ID,
		Key:            schema.LabelKey,
		Value:          schema.LabelValue,
	}); err != nil {
		return nil, err
	}

	return c.doUpdate(ctx, schema.UpdateDeploymentSchema, org, deployment)
}

func (c *deploymentController) doUpdate(ctx *gin.Context, schema schemasv1.UpdateDeploymentSchema, org *models.Organization, deployment *models.Deployment, label *models.Label) (*schemasv1.DeploymentSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	bentoNames := make([]string, 0, len(schema.Targets))
	bentoNamesSeen := make(map[string]struct{}, len(schema.Targets))

	bentoVersionVersionsMapping := make(map[string][]string, len(schema.Targets))

	for _, createDeploymentTargetSchema := range schema.Targets {
		if _, ok := bentoNamesSeen[createDeploymentTargetSchema.BentoName]; !ok {
			bentoNames = append(bentoNames, createDeploymentTargetSchema.BentoName)
			bentoNamesSeen[createDeploymentTargetSchema.BentoName] = struct{}{}
		}
		bentoVersionVersions, ok := bentoVersionVersionsMapping[createDeploymentTargetSchema.BentoName]
		if !ok {
			bentoVersionVersions = make([]string, 0, 1)
		}
		bentoVersionVersions = append(bentoVersionVersions, createDeploymentTargetSchema.BentoVersion)
		bentoVersionVersionsMapping[createDeploymentTargetSchema.BentoName] = bentoVersionVersions
	}

	bentos, _, err := services.BentoService.List(ctx, services.ListBentoOption{
		OrganizationId: utils.UintPtr(org.ID),
		Names:          &bentoNames,
	})
	if err != nil {
		return nil, err
	}
	bentosMapping := make(map[string]*models.Bento, len(bentos))
	for _, bento := range bentos {
		bentosMapping[bento.Name] = bento
	}

	bentoVersionsMapping := make(map[string]*models.BentoVersion)
	for _, bento := range bentos {
		versions := bentoVersionVersionsMapping[bento.Name]
		bentoVersions, _, err := services.BentoVersionService.List(ctx, services.ListBentoVersionOption{
			BentoId:  utils.UintPtr(bento.ID),
			Versions: &versions,
		})
		if err != nil {
			return nil, err
		}
		for _, bentoVersion := range bentoVersions {
			bentoVersionsMapping[fmt.Sprintf("%s:%s", bento.Name, bentoVersion.Version)] = bentoVersion
		}
	}

	deploymentRevision, err := services.DeploymentRevisionService.Create(ctx, services.CreateDeploymentRevisionOption{
		CreatorId:    user.ID,
		DeploymentId: deployment.ID,
		Status:       modelschemas.DeploymentRevisionStatusActive,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create deployment revision")
	}

	deploymentTargets := make([]*models.DeploymentTarget, 0, len(schema.Targets))
	for _, createDeploymentTargetSchema := range schema.Targets {
		bentoVersion := bentoVersionsMapping[fmt.Sprintf("%s:%s", createDeploymentTargetSchema.BentoName, createDeploymentTargetSchema.BentoVersion)]
		if bentoVersion == nil {
			return nil, errors.Errorf("can't find bento version: %s:%s", createDeploymentTargetSchema.BentoName, createDeploymentTargetSchema.BentoVersion)
		}

		deploymentTarget, err := services.DeploymentTargetService.Create(ctx, services.CreateDeploymentTargetOption{
			CreatorId:            user.ID,
			DeploymentId:         deployment.ID,
			DeploymentRevisionId: deploymentRevision.ID,
			BentoVersionId:       bentoVersion.ID,
			Type:                 createDeploymentTargetSchema.Type,
			CanaryRules:          createDeploymentTargetSchema.CanaryRules,
			Config:               createDeploymentTargetSchema.Config,
		})
		if err != nil {
			return nil, errors.Wrap(err, "create deployment target")
		}
		deploymentTargets = append(deploymentTargets, deploymentTarget)
	}
	// TODO: create label here (call the label service methods)

	err = services.DeploymentRevisionService.Deploy(ctx, deploymentRevision, deploymentTargets, false)
	if err != nil {
		return nil, errors.Wrap(err, "deploy deployment revision")
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

func (c *deploymentController) Terminate(ctx *gin.Context, schema *GetDeploymentSchema) (*schemasv1.DeploymentSchema, error) {
	deployment, err := schema.GetDeployment(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canOperate(ctx, deployment); err != nil {
		return nil, err
	}
	deployment, err = services.DeploymentService.Terminate(ctx, deployment)
	if err != nil {
		return nil, err
	}
	return transformersv1.ToDeploymentSchema(ctx, deployment)
}

func (c *deploymentController) Delete(ctx *gin.Context, schema *GetDeploymentSchema) (*schemasv1.DeploymentSchema, error) {
	deployment, err := schema.GetDeployment(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canOperate(ctx, deployment); err != nil {
		return nil, err
	}
	deployment, err = services.DeploymentService.Delete(ctx, deployment)
	if err != nil {
		return nil, err
	}
	return transformersv1.ToDeploymentSchema(ctx, deployment)
}

type ListClusterDeploymentSchema struct {
	schemasv1.ListQuerySchema
	GetClusterSchema
}

func fillListDeploymentOption(ctx context.Context, listOpt *services.ListDeploymentOption, queryMap map[string]interface{}) error {
	for k, v := range queryMap {
		if k == schemasv1.KeyQIn {
			fieldNames := make([]string, 0, len(v.([]string)))
			for _, fieldName := range v.([]string) {
				if _, ok := map[string]struct{}{
					"name":        {},
					"description": {},
				}[fieldName]; !ok {
					continue
				}
				fieldNames = append(fieldNames, fieldName)
			}
			listOpt.KeywordFieldNames = &fieldNames
		}
		if k == schemasv1.KeyQKeywords {
			listOpt.Keywords = utils.StringSlicePtr(v.([]string))
		}
		if k == "cluster" {
			clusters, _, err := services.ClusterService.List(ctx, services.ListClusterOption{
				Names: utils.StringSlicePtr(v.([]string)),
			})
			if err != nil {
				return err
			}
			clusterIds := make([]uint, 0, len(clusters))
			for _, cluster := range clusters {
				clusterIds = append(clusterIds, cluster.ID)
			}
			listOpt.ClusterIds = &clusterIds
		}
		if k == "creator" {
			userNames, err := processUserNamesFromQ(ctx, v.([]string))
			if err != nil {
				return err
			}
			users, err := services.UserService.ListByNames(ctx, userNames)
			if err != nil {
				return err
			}
			userIds := make([]uint, 0, len(users))
			for _, user := range users {
				userIds = append(userIds, user.ID)
			}
			listOpt.CreatorIds = utils.UintSlicePtr(userIds)
		}
		if k == "last_updater" {
			userNames, err := processUserNamesFromQ(ctx, v.([]string))
			if err != nil {
				return err
			}
			users, err := services.UserService.ListByNames(ctx, userNames)
			if err != nil {
				return err
			}
			userIds := make([]uint, 0, len(users))
			for _, user := range users {
				userIds = append(userIds, user.ID)
			}
			listOpt.LastUpdaterIds = utils.UintSlicePtr(userIds)
		}
		if k == "status" {
			statuses := make([]modelschemas.DeploymentStatus, 0, len(v.([]string)))
			for _, status := range v.([]string) {
				statuses = append(statuses, modelschemas.DeploymentStatus(status))
			}
			listOpt.Statuses = &statuses
		}
		if k == "sort" {
			fieldName, _, order := xstrings.LastPartition(v.([]string)[0], "-")
			if _, ok := map[string]struct{}{
				"created_at": {},
				"updated_at": {},
			}[fieldName]; !ok {
				continue
			}
			if _, ok := map[string]struct{}{
				"desc": {},
				"asc":  {},
			}[order]; !ok {
				continue
			}
			if fieldName == "updated_at" {
				fieldName = "deployment_revision.created_at"
			}
			listOpt.Order = utils.StringPtr(fmt.Sprintf("%s %s", fieldName, strings.ToUpper(order)))
		}
	}
	return nil
}

func (c *deploymentController) ListClusterDeployments(ctx *gin.Context, schema *ListClusterDeploymentSchema) (*schemasv1.DeploymentListSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}

	if err = ClusterController.canView(ctx, cluster); err != nil {
		return nil, err
	}

	listOpt := services.ListDeploymentOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		ClusterId: utils.UintPtr(cluster.ID),
	}

	err = fillListDeploymentOption(ctx, &listOpt, schema.Q.ToMap())
	if err != nil {
		return nil, err
	}

	deployments, total, err := services.DeploymentService.List(ctx, listOpt)
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

type ListOrganizationDeploymentSchema struct {
	schemasv1.ListQuerySchema
	GetOrganizationSchema
}

func (c *deploymentController) ListOrganizationDeployments(ctx *gin.Context, schema *ListOrganizationDeploymentSchema) (*schemasv1.DeploymentListSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = OrganizationController.canView(ctx, organization); err != nil {
		return nil, err
	}

	listOpt := services.ListDeploymentOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		OrganizationId: utils.UintPtr(organization.ID),
	}

	err = fillListDeploymentOption(ctx, &listOpt, schema.Q.ToMap())
	if err != nil {
		return nil, err
	}

	deployments, total, err := services.DeploymentService.List(ctx, listOpt)
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

	kubeNs := services.DeploymentService.GetKubeNamespace(deployment)
	podInformer, podLister, err := services.GetPodInformer(ctx, cluster, kubeNs)
	if err != nil {
		return err
	}

	pods, err := services.KubePodService.ListPodsByDeployment(ctx, podLister, deployment)
	if err != nil {
		return err
	}

	var podSchemas []*schemasv1.KubePodSchema

	podSchemas, err = transformersv1.ToKubePodSchemas(ctx, pods)
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

		newPodSchemas, err := transformersv1.ToKubePodSchemas(ctx, pods)
		if err != nil {
			logrus.Errorf("get app pods failed: %q", err.Error())
			failed()
			return err
		}

		viewChanged := !reflect.DeepEqual(podSchemas, newPodSchemas)
		if viewChanged {
			go func() {
				deployment_, err := services.DeploymentService.Get(ctx, deployment.ID)
				if err != nil {
					return
				}
				_, _ = services.DeploymentService.SyncStatus(ctx, deployment_)
			}()
		}
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

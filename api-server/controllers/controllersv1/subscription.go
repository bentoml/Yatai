package controllersv1

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
)

type subscriptionController struct {
	// nolint: unused
	baseController
}

var SubscriptionController = subscriptionController{}

func (c *subscriptionController) SubscribeResource(ctx *gin.Context) error {
	ctx.Request.Header.Del("Origin")
	conn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logrus.Errorf("ws connect failed: %q", err.Error())
		return err
	}
	defer conn.Close()

	currentUser, err := services.GetCurrentUser(ctx)
	if err != nil {
		return err
	}

	var mu sync.RWMutex
	resourceUidsMap := make(map[modelschemas.ResourceType][]string)
	schemasCache := make(map[string]interface{})

	pollingCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		for {
			var req schemasv1.SubscriptionReqSchema
			_, msg, err := conn.ReadMessage()

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logrus.Errorf("ws read failed: %q", err.Error())
				}
				cancel()
				return
			}

			err = json.Unmarshal(msg, &req)
			if err != nil {
				writeWsError(conn, err)
				continue
			}

			if req.Type == schemasv1.WsReqTypeHeartbeat {
				continue
			}

			switch req.Payload.Action {
			case schemasv1.SubscriptionActionSubscribe:
				actualUids := make([]string, 0, len(req.Payload.ResourceUids))
				// nolint: exhaustive
				switch req.Payload.ResourceType {
				case modelschemas.ResourceTypeBento:
					bentos, err := services.BentoService.ListByUids(ctx, req.Payload.ResourceUids)
					if err != nil {
						writeWsError(conn, err)
						continue
					}
					for _, bento := range bentos {
						bentoRepository, err := services.BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
						if err != nil {
							writeWsError(conn, err)
							continue
						}
						if err = services.MemberService.CanView(ctx, &services.OrganizationMemberService, currentUser, bentoRepository.OrganizationId); err != nil {
							writeWsError(conn, err)
							continue
						}
						actualUids = append(actualUids, bento.Uid)
					}
				case modelschemas.ResourceTypeModel:
					models, err := services.ModelService.ListByUids(ctx, req.Payload.ResourceUids)
					if err != nil {
						writeWsError(conn, err)
						continue
					}
					for _, model := range models {
						modelRepository, err := services.ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
						if err != nil {
							writeWsError(conn, err)
							continue
						}
						if err = services.MemberService.CanView(ctx, &services.OrganizationMemberService, currentUser, modelRepository.OrganizationId); err != nil {
							writeWsError(conn, err)
							continue
						}
						actualUids = append(actualUids, model.Uid)
					}
				case modelschemas.ResourceTypeDeployment:
					deployments, err := services.DeploymentService.ListByUids(ctx, req.Payload.ResourceUids)
					if err != nil {
						writeWsError(conn, err)
						continue
					}
					for _, deployment := range deployments {
						cluster, err := services.ClusterService.GetAssociatedCluster(ctx, deployment)
						if err != nil {
							writeWsError(conn, err)
							continue
						}
						if err = services.MemberService.CanView(ctx, &services.ClusterMemberService, currentUser, cluster.ID); err != nil {
							writeWsError(conn, err)
							continue
						}
						actualUids = append(actualUids, deployment.Uid)
					}
				default:
					continue
				}
				mu.Lock()
				uids, ok := resourceUidsMap[req.Payload.ResourceType]
				if !ok {
					uids = make([]string, 0, len(actualUids))
				}
				uids = append(uids, actualUids...)
				resourceUidsMap[req.Payload.ResourceType] = uids
				mu.Unlock()
			case schemasv1.SubscriptionActionUnsubscribe:
				mu.Lock()
				uids, ok := resourceUidsMap[req.Payload.ResourceType]
				if ok {
					seen := make(map[string]struct{}, len(req.Payload.ResourceUids))
					for _, uid := range req.Payload.ResourceUids {
						seen[uid] = struct{}{}
					}
					newUids := make([]string, 0)
					for _, uid := range uids {
						if _, ok := seen[uid]; ok {
							continue
						}
						newUids = append(newUids, uid)
					}
					resourceUidsMap[req.Payload.ResourceType] = newUids
				}
				mu.Unlock()
			}
		}
	}()

	send := func() error {
		for resourceType, uids := range resourceUidsMap {
			// nolint: exhaustive
			switch resourceType {
			case modelschemas.ResourceTypeBento:
				bentos, err := services.BentoService.ListByUids(ctx, uids)
				if err != nil {
					return err
				}
				bentoSchemas, err := transformersv1.ToBentoSchemas(ctx, bentos)
				if err != nil {
					return err
				}
				for _, bentoSchema := range bentoSchemas {
					isEqual := func() bool {
						mu.Lock()
						defer func() {
							schemasCache[bentoSchema.Uid] = bentoSchema
						}()
						defer mu.Unlock()
						if oldSchema, ok := schemasCache[bentoSchema.Uid]; ok {
							return reflect.DeepEqual(oldSchema, bentoSchema)
						}
						return false
					}()

					if isEqual {
						continue
					}

					err = conn.WriteJSON(&schemasv1.WsRespSchema{
						Type:    schemasv1.WsRespTypeSuccess,
						Message: "",
						Payload: &schemasv1.SubscriptionRespSchema{
							ResourceType: bentoSchema.ResourceType,
							Payload:      bentoSchema,
						},
					})
					if err != nil {
						return err
					}
				}
			case modelschemas.ResourceTypeModel:
				models, err := services.ModelService.ListByUids(ctx, uids)
				if err != nil {
					return err
				}
				modelSchemas, err := transformersv1.ToModelSchemas(ctx, models)
				if err != nil {
					return err
				}
				for _, modelSchema := range modelSchemas {
					isEqual := func() bool {
						mu.Lock()
						defer func() {
							schemasCache[modelSchema.Uid] = modelSchema
						}()
						defer mu.Unlock()
						if oldSchema, ok := schemasCache[modelSchema.Uid]; ok {
							return reflect.DeepEqual(oldSchema, modelSchema)
						}
						return false
					}()

					if isEqual {
						continue
					}

					err = conn.WriteJSON(&schemasv1.WsRespSchema{
						Type:    schemasv1.WsRespTypeSuccess,
						Message: "",
						Payload: &schemasv1.SubscriptionRespSchema{
							ResourceType: modelSchema.ResourceType,
							Payload:      modelSchema,
						},
					})
					if err != nil {
						return err
					}
				}
			case modelschemas.ResourceTypeDeployment:
				deployments, err := services.DeploymentService.ListByUids(ctx, uids)
				if err != nil {
					return err
				}
				deploymentSchemas, err := transformersv1.ToDeploymentSchemas(ctx, deployments)
				if err != nil {
					return err
				}
				for _, deploymentSchema := range deploymentSchemas {
					isEqual := func() bool {
						mu.Lock()
						defer func() {
							schemasCache[deploymentSchema.Uid] = deploymentSchema
						}()
						defer mu.Unlock()
						if oldSchema, ok := schemasCache[deploymentSchema.Uid]; ok {
							return reflect.DeepEqual(oldSchema, deploymentSchema)
						}
						return false
					}()

					if isEqual {
						continue
					}

					err = conn.WriteJSON(&schemasv1.WsRespSchema{
						Type:    schemasv1.WsRespTypeSuccess,
						Message: "",
						Payload: &schemasv1.SubscriptionRespSchema{
							ResourceType: deploymentSchema.ResourceType,
							Payload:      deploymentSchema,
						},
					})
					if err != nil {
						return err
					}
				}
			default:
				continue
			}
		}

		return nil
	}

	func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		for {
			select {
			case <-pollingCtx.Done():
				return
			default:
				err = send()
				if err != nil {
					writeWsError(conn, err)
				}
			}

			<-ticker.C
		}
	}()

	return nil
}

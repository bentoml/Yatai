package controllersv1

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/schemas/modelschemas"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type subscriptionController struct {
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

	go func() {
		for {
			var req schemasv1.SubscriptionReqSchema
			mt, msg, err := conn.ReadMessage()

			if err != nil || mt == websocket.CloseMessage || mt == -1 {
				cancel()
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logrus.Printf("error: %v", err)
				}
				break
			}

			err = json.Unmarshal(msg, &req)
			if err != nil {
				_ = conn.WriteJSON(&schemasv1.WsRespSchema{
					Type:    schemasv1.WsRespTypeError,
					Message: fmt.Sprintf("cannot read json: %s", err.Error()),
				})
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
				case modelschemas.ResourceTypeBentoVersion:
					bentoVersions, err := services.BentoVersionService.ListByUids(ctx, req.Payload.ResourceUids)
					if err != nil {
						_ = conn.WriteJSON(&schemasv1.WsRespSchema{
							Type:    schemasv1.WsRespTypeError,
							Message: err.Error(),
						})
						continue
					}
					for _, bentoVersion := range bentoVersions {
						bento, err := services.BentoService.GetAssociatedBento(ctx, bentoVersion)
						if err != nil {
							_ = conn.WriteJSON(&schemasv1.WsRespSchema{
								Type:    schemasv1.WsRespTypeError,
								Message: err.Error(),
							})
							continue
						}
						if err = services.MemberService.CanView(ctx, &services.OrganizationMemberService, currentUser.ID, bento.OrganizationId); err != nil {
							_ = conn.WriteJSON(&schemasv1.WsRespSchema{
								Type:    schemasv1.WsRespTypeError,
								Message: err.Error(),
							})
							continue
						}
						actualUids = append(actualUids, bentoVersion.Uid)
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
			case modelschemas.ResourceTypeBentoVersion:
				bentoVersions, err := services.BentoVersionService.ListByUids(ctx, uids)
				if err != nil {
					return err
				}
				bentoVersionSchemas, err := transformersv1.ToBentoVersionSchemas(ctx, bentoVersions)
				if err != nil {
					return err
				}
				for _, bentoVersionSchema := range bentoVersionSchemas {
					isEqual := func() bool {
						mu.Lock()
						defer func() {
							schemasCache[bentoVersionSchema.Uid] = bentoVersionSchema
						}()
						defer mu.Unlock()
						if oldSchema, ok := schemasCache[bentoVersionSchema.Uid]; ok {
							return reflect.DeepEqual(oldSchema, bentoVersionSchema)
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
							ResourceType: bentoVersionSchema.ResourceType,
							Payload:      bentoVersionSchema,
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
					_ = conn.WriteJSON(&schemasv1.WsRespSchema{
						Type:    schemasv1.WsRespTypeError,
						Message: err.Error(),
					})
				}
			}

			<-ticker.C
		}
	}()

	return nil
}

package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type SubscriptionAction string

const (
	SubscriptionActionSubscribe   SubscriptionAction = "subscribe"
	SubscriptionActionUnsubscribe SubscriptionAction = "unsubscribe"
)

type SubscriptionRespSchema struct {
	ResourceType modelschemas.ResourceType `json:"resource_type"`
	Payload      interface{}               `json:"payload"`
}

type SubscriptionReqSchema struct {
	WsReqSchema
	Payload *struct {
		Action       SubscriptionAction        `json:"action"`
		ResourceType modelschemas.ResourceType `json:"resource_type"`
		ResourceUids []string                  `json:"resource_uids"`
	} `json:"payload"`
}

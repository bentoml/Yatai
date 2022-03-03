package models

import (
	apiv1 "k8s.io/api/core/v1"

	"github.com/bentoml/yatai-schemas/modelschemas"
)

type KubePodWithStatus struct {
	Pod      apiv1.Pod
	Status   modelschemas.KubePodStatus
	Warnings []apiv1.Event
}

package models

import apiv1 "k8s.io/api/core/v1"

type KubePodStatus struct {
	Status          string                 `json:"status"`
	Phase           apiv1.PodPhase         `json:"phase"`
	ContainerStates []apiv1.ContainerState `json:"container_states"`
}

type KubePodWithStatus struct {
	Pod      apiv1.Pod
	Status   KubePodStatus
	Warnings []apiv1.Event
}

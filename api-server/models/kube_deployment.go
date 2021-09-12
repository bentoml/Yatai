package models

import appsv1 "k8s.io/api/apps/v1"

type KubeDeploymentWithPods struct {
	Deployment appsv1.Deployment
	Pods       []*KubePodWithStatus
}

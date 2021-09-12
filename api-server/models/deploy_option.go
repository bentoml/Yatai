package models

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeployOption struct {
	Force           bool
	OwnerReferences []metav1.OwnerReference
}

package main

import (
	"k8s.io/apimachinery/pkg/runtime"

	grafanav1alpha1 "github.com/bentoml/grafana-operator/api/integreatly/v1alpha1"
	"github.com/bentoml/yatai/api-server/cmd"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = grafanav1alpha1.AddToScheme(scheme)
}

func main() {
	cmd.Execute()
}

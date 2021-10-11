package main

import (
	grafanav1alpha1 "github.com/bentoml/grafana-operator/api/integreatly/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"

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

package models

import (
	"helm.sh/helm/v3/pkg/release"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type YataiComponent struct {
	Type    modelschemas.YataiComponentType `json:"type"`
	Release *release.Release                `json:"release"`
}

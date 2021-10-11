package schemasv1

import (
	"helm.sh/helm/v3/pkg/release"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type YataiComponentSchema struct {
	Type    modelschemas.YataiComponentType `json:"type"`
	Release *release.Release                `json:"release"`
}

type CreateYataiComponentSchema struct {
	Type modelschemas.YataiComponentType `json:"type"`
}

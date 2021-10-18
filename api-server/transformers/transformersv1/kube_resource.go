package transformersv1

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToKubeResourceSchemas(ctx context.Context, resources []*models.KubeResource) (ss []*schemasv1.KubeResourceSchema, err error) {
	for _, r := range resources {
		ss = append(ss, &schemasv1.KubeResourceSchema{
			APIVersion:  r.APIVersion,
			Kind:        r.Kind,
			Name:        r.Name,
			Namespace:   r.Namespace,
			MatchLabels: r.MatchLabels,
		})
	}
	return
}

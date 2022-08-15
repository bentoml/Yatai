package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/models"
)

func ToYataiComponentSchema(ctx context.Context, yataiComponent *models.YataiComponent) (*schemasv1.YataiComponentSchema, error) {
	if yataiComponent == nil {
		return nil, nil
	}
	ss, err := ToYataiComponentSchemas(ctx, []*models.YataiComponent{yataiComponent})
	if err != nil {
		return nil, errors.Wrap(err, "ToYataiComponentSchemas")
	}
	return ss[0], nil
}

func ToYataiComponentSchemas(ctx context.Context, yataiComponents []*models.YataiComponent) ([]*schemasv1.YataiComponentSchema, error) {
	resourceSchemaMap, err := ToResourceSchemasMap(ctx, yataiComponents)
	if err != nil {
		return nil, errors.Wrap(err, "ToResourceSchemasMap")
	}

	res := make([]*schemasv1.YataiComponentSchema, 0, len(yataiComponents))
	for _, yataiComponent := range yataiComponents {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, yataiComponent)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		clusterSchema, err := GetAssociatedClusterFullSchema(ctx, yataiComponent)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedClusterSchema")
		}
		resourceSchema, ok := resourceSchemaMap[yataiComponent.GetUid()]
		if !ok {
			return nil, errors.Errorf("resource schema not found for yataiComponent %s", yataiComponent.GetUid())
		}
		res = append(res, &schemasv1.YataiComponentSchema{
			ResourceSchema:    resourceSchema,
			Creator:           creatorSchema,
			Cluster:           clusterSchema,
			Manifest:          yataiComponent.Manifest,
			Version:           yataiComponent.Version,
			KubeNamespace:     yataiComponent.KubeNamespace,
			LatestHeartbeatAt: yataiComponent.LatestHeartbeatAt,
			LatestInstalledAt: yataiComponent.LatestInstalledAt,
		})
	}
	return res, nil
}

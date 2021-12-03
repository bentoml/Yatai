package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToModelSchema(ctx context.Context, model *models.Model) (*schemasv1.ModelSchema, error) {
	if model == nil {
		return nil, nil
	}
	ss, err := ToModelSchemas(ctx, []*models.Model{model})
	if err != nil {
		return nil, errors.Wrap(err, "ToModelSchemas")
	}
	return ss[0], nil
}

func ToModelSchemas(ctx context.Context, models []*models.Model) ([]*schemasv1.ModelSchema, error) {
	modelIds := make([]uint, 0, len(models))
	for _, model := range models {
		modelIds = append(modelIds, model.ID)
	}
	versions, err := services.ModelVersionService.ListLatestByModelIds(ctx, modelIds)
	if err != nil {
		return nil, errors.Wrap(err, "List Latest version ByModel Ids")
	}
	versionSchemas, err := ToModelVersionSchemas(ctx, versions)
	if err != nil {
		return nil, errors.Wrap(err, "ToModelVersionSchemas")
	}
	versionSchemasMap := make(map[string]*schemasv1.ModelVersionSchema, len(versions))
	for _, s := range versionSchemas {
		versionSchemasMap[s.ModelUid] = s
	}
	resourceSchemasMap, err := ToResourceSchemasMap(ctx, models)
	if err != nil {
		return nil, errors.Wrap(err, "ToResourceSchemasMap")
	}
	res := make([]*schemasv1.ModelSchema, 0, len(models))
	for _, model := range models {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, model)
		if err != nil {
			return nil, errors.Wrap(err, "get associated creator schema")
		}
		organizationSchema, err := GetAssociatedOrganizationSchema(ctx, model)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedOrganizationSchema")
		}
		resourceSchema, ok := resourceSchemasMap[model.GetUid()]
		if !ok {
			return nil, errors.Errorf("resource schema not found for model %s", model.GetUid())
		}
		res = append(res, &schemasv1.ModelSchema{
			ResourceSchema: resourceSchema,
			Creator:        creatorSchema,
			Organization:   organizationSchema,
			Description:    model.Description,
			LatestVersion:  versionSchemasMap[model.GetUid()],
		})
	}
	return res, nil
}

type IModelAssociate interface {
	services.IModelAssociate
	models.IResource
}

func GetAssociatedModelSchema(ctx context.Context, associate IModelAssociate) (*schemasv1.ModelSchema, error) {
	user, err := services.ModelService.GetAssociatedModel(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated cluster", associate.GetResourceType(), associate.GetName())
	}
	userSchema, err := ToModelSchema(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "ToModelSchema")
	}
	return userSchema, nil
}

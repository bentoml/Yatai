package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToBentoSchema(ctx context.Context, bento *models.Bento) (*schemasv1.BentoSchema, error) {
	if bento == nil {
		return nil, nil
	}
	ss, err := ToBentoSchemas(ctx, []*models.Bento{bento})
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoSchemas")
	}
	return ss[0], nil
}

func ToBentoSchemas(ctx context.Context, bentos []*models.Bento) ([]*schemasv1.BentoSchema, error) {
	bentoIds := make([]uint, 0, len(bentos))
	for _, bento := range bentos {
		bentoIds = append(bentoIds, bento.ID)
	}

	versions, err := services.BentoVersionService.ListLatestByBentoIds(ctx, bentoIds)
	if err != nil {
		return nil, errors.Wrap(err, "list latest versions by bento ids")
	}
	versionSchemas, err := ToBentoVersionSchemas(ctx, versions)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoVersionSchemas")
	}
	versionSchemasMap := make(map[string]*schemasv1.BentoVersionSchema, len(versions))
	for _, s := range versionSchemas {
		versionSchemasMap[s.BentoUid] = s
	}

	res := make([]*schemasv1.BentoSchema, 0, len(bentos))
	for _, bento := range bentos {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, bento)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		organizationSchema, err := GetAssociatedOrganizationSchema(ctx, bento)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedClusterSchema")
		}
		res = append(res, &schemasv1.BentoSchema{
			ResourceSchema: ToResourceSchema(bento),
			Creator:        creatorSchema,
			Organization:   organizationSchema,
			Description:    bento.Description,
			LatestVersion:  versionSchemasMap[bento.GetUid()],
		})
	}
	return res, nil
}

type IBentoAssociate interface {
	services.IBentoAssociate
	models.IResource
}

func GetAssociatedBentoSchema(ctx context.Context, associate IBentoAssociate) (*schemasv1.BentoSchema, error) {
	user, err := services.BentoService.GetAssociatedBento(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated cluster", associate.GetResourceType(), associate.GetName())
	}
	userSchema, err := ToBentoSchema(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoSchema")
	}
	return userSchema, nil
}

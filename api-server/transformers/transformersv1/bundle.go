package transformersv1

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
	"github.com/pkg/errors"
)

func ToBundleSchema(ctx context.Context, bundle *models.Bundle) (*schemasv1.BundleSchema, error) {
	if bundle == nil {
		return nil, nil
	}
	ss, err := ToBundleSchemas(ctx, []*models.Bundle{bundle})
	if err != nil {
		return nil, errors.Wrap(err, "ToBundleSchemas")
	}
	return ss[0], nil
}

func ToBundleSchemas(ctx context.Context, bundles []*models.Bundle) ([]*schemasv1.BundleSchema, error) {
	bundleIds := make([]uint, 0, len(bundles))
	for _, bundle := range bundles {
		bundleIds = append(bundleIds, bundle.ID)
	}

	versions, err := services.BundleVersionService.ListLatestByBundleIds(ctx, bundleIds)
	if err != nil {
		return nil, errors.Wrap(err, "list latest versions by bundle ids")
	}
	versionSchemas, err := ToBundleVersionSchemas(ctx, versions)
	if err != nil {
		return nil, errors.Wrap(err, "ToBundleVersionSchemas")
	}
	versionSchemasMap := make(map[string]*schemasv1.BundleVersionSchema, len(versions))
	for _, s := range versionSchemas {
		versionSchemasMap[s.BundleUid] = s
	}

	res := make([]*schemasv1.BundleSchema, 0, len(bundles))
	for _, bundle := range bundles {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, bundle)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		organizationSchema, err := GetAssociatedOrganizationSchema(ctx, bundle)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedClusterSchema")
		}
		res = append(res, &schemasv1.BundleSchema{
			ResourceSchema: ToResourceSchema(bundle),
			Creator:        creatorSchema,
			Organization:   organizationSchema,
			Description:    bundle.Description,
			LatestVersion:  versionSchemasMap[bundle.GetUid()],
		})
	}
	return res, nil
}

type IBundleAssociate interface {
	services.IBundleAssociate
	models.IResource
}

func GetAssociatedBundleSchema(ctx context.Context, associate IBundleAssociate) (*schemasv1.BundleSchema, error) {
	user, err := services.BundleService.GetAssociatedBundle(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated cluster", associate.GetResourceType(), associate.GetName())
	}
	userSchema, err := ToBundleSchema(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "ToBundleSchema")
	}
	return userSchema, nil
}

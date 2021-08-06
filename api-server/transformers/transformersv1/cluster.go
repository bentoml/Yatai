package transformersv1

import (
	"context"

	"github.com/bentoml/yatai/schemas/modelschemas"
	jujuerrors "github.com/juju/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
	"github.com/pkg/errors"
)

func ToClusterSchema(ctx context.Context, cluster *models.Cluster) (*schemasv1.ClusterSchema, error) {
	if cluster == nil {
		return nil, nil
	}
	ss, err := ToClusterSchemas(ctx, []*models.Cluster{cluster})
	if err != nil {
		return nil, errors.Wrap(err, "ToClusterSchemas")
	}
	return ss[0], nil
}

func ToClusterSchemas(ctx context.Context, clusters []*models.Cluster) ([]*schemasv1.ClusterSchema, error) {
	res := make([]*schemasv1.ClusterSchema, 0, len(clusters))
	for _, cluster := range clusters {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, cluster)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		res = append(res, &schemasv1.ClusterSchema{
			ResourceSchema: ToResourceSchema(cluster),
			Creator:        creatorSchema,
			Description:    cluster.Description,
		})
	}
	return res, nil
}

func ToClusterFullSchema(ctx context.Context, cluster *models.Cluster) (*schemasv1.ClusterFullSchema, error) {
	if cluster == nil {
		return nil, nil
	}
	s, err := ToClusterSchema(ctx, cluster)
	if err != nil {
		return nil, errors.Wrap(err, "ToClusterSchema")
	}
	org, err := services.OrganizationService.GetAssociatedOrganization(ctx, cluster)
	if err != nil {
		return nil, errors.Wrap(err, "get organization")
	}
	orgSchema, err := ToOrganizationSchema(ctx, org)
	if err != nil {
		return nil, errors.Wrap(err, "to organization schema")
	}
	var kubeConfig *string
	var config **modelschemas.ClusterConfigSchema
	currentUser, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get current user")
	}
	if err = services.MemberService.CanUpdate(ctx, &services.ClusterMemberService, currentUser.ID, cluster.ID); err != nil {
		if !jujuerrors.IsForbidden(err) {
			return nil, err
		}
	} else {
		kubeConfig = &cluster.KubeConfig
		config = &cluster.Config
	}
	return &schemasv1.ClusterFullSchema{
		ClusterSchema: *s,
		Organization:  orgSchema,
		KubeConfig:    kubeConfig,
		Config:        config,
	}, nil
}

type IClusterAssociate interface {
	services.IClusterAssociate
	models.IResource
}

func GetAssociatedClusterSchema(ctx context.Context, associate IClusterAssociate) (*schemasv1.ClusterSchema, error) {
	user, err := services.ClusterService.GetAssociatedCluster(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated cluster", associate.GetResourceType(), associate.GetName())
	}
	userSchema, err := ToClusterSchema(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "ToClusterSchema")
	}
	return userSchema, nil
}

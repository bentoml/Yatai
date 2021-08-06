package controllersv1

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type clusterController struct {
	baseController
}

var ClusterController = clusterController{}

type GetClusterSchema struct {
	GetOrganizationSchema
	ClusterName string `path:"clusterName"`
}

func (s *GetClusterSchema) GetCluster(ctx context.Context) (*models.Cluster, error) {
	org, err := s.GetOrganization(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "get organization %s", org.Name)
	}
	cluster, err := services.ClusterService.GetByName(ctx, org.ID, s.ClusterName)
	if err != nil {
		return nil, errors.Wrapf(err, "get cluster %s", s.ClusterName)
	}
	return cluster, nil
}

func (c *clusterController) canView(ctx context.Context, cluster *models.Cluster) error {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return err
	}
	return services.MemberService.CanView(ctx, &services.ClusterMemberService, user.ID, cluster.ID)
}

func (c *clusterController) canUpdate(ctx context.Context, cluster *models.Cluster) error {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return err
	}
	return services.MemberService.CanUpdate(ctx, &services.ClusterMemberService, user.ID, cluster.ID)
}

func (c *clusterController) canOperate(ctx context.Context, cluster *models.Cluster) error {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return err
	}
	return services.MemberService.CanOperate(ctx, &services.ClusterMemberService, user.ID, cluster.ID)
}

type CreateClusterSchema struct {
	schemasv1.CreateClusterSchema
	GetOrganizationSchema
}

func (c *clusterController) Create(ctx *gin.Context, schema *CreateClusterSchema) (*schemasv1.ClusterSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = OrganizationController.canOperate(ctx, org); err != nil {
		return nil, err
	}

	cluster, err := services.ClusterService.Create(ctx, services.CreateClusterOption{
		CreatorId:      user.ID,
		OrganizationId: org.ID,
		Name:           schema.Name,
		Description:    schema.Description,
		KubeConfig:     schema.KubeConfig,
		Config:         schema.Config,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create cluster")
	}
	return transformersv1.ToClusterSchema(ctx, cluster)
}

type UpdateClusterSchema struct {
	schemasv1.UpdateClusterSchema
	GetClusterSchema
}

func (c *clusterController) Update(ctx *gin.Context, schema *UpdateClusterSchema) (*schemasv1.ClusterSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, cluster); err != nil {
		return nil, err
	}
	cluster, err = services.ClusterService.Update(ctx, cluster, services.UpdateClusterOption{
		Description: schema.Description,
		Config:      schema.Config,
		KubeConfig:  schema.KubeConfig,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update cluster")
	}
	return transformersv1.ToClusterSchema(ctx, cluster)
}

func (c *clusterController) Get(ctx *gin.Context, schema *GetClusterSchema) (*schemasv1.ClusterSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, cluster); err != nil {
		return nil, err
	}
	return transformersv1.ToClusterSchema(ctx, cluster)
}

type ListClusterSchema struct {
	schemasv1.ListQuerySchema
	GetOrganizationSchema
}

func (c *clusterController) List(ctx *gin.Context, schema *ListClusterSchema) (*schemasv1.ClusterListSchema, error) {
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = OrganizationController.canView(ctx, org); err != nil {
		return nil, err
	}

	clusters, total, err := services.ClusterService.List(ctx, services.ListClusterOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		OrganizationId: utils.UintPtr(org.ID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list clusters")
	}

	clusterSchemas, err := transformersv1.ToClusterSchemas(ctx, clusters)
	return &schemasv1.ClusterListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: clusterSchemas,
	}, err
}

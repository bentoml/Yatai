package controllersv1

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"

	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/schemas/modelschemas"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type yataiComponentController struct {
	clusterController
}

var YataiComponentController = yataiComponentController{}

func (c *yataiComponentController) ListOperatorHelmCharts(ctx *gin.Context) ([]*chart.Chart, error) {
	return services.YataiComponentService.ListOperatorHelmCharts(ctx)
}

func (c *yataiComponentController) List(ctx *gin.Context, schema *GetClusterSchema) ([]*schemasv1.YataiComponentSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, cluster); err != nil {
		return nil, err
	}
	comps, err := services.YataiComponentService.List(ctx, cluster.ID)
	if err != nil {
		return nil, errors.Wrap(err, "list cluster yatai comps")
	}
	return transformersv1.ToYataiComponentSchemas(ctx, comps)
}

type CreateYataiComponentSchema struct {
	schemasv1.CreateYataiComponentSchema
	GetClusterSchema
}

func (c *yataiComponentController) Create(ctx *gin.Context, schema *CreateYataiComponentSchema) (*schemasv1.YataiComponentSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canOperate(ctx, cluster); err != nil {
		return nil, err
	}
	comp, err := services.YataiComponentService.Create(ctx, services.CreateYataiComponentReleaseOption{
		ClusterId: cluster.ID,
		Type:      schema.Type,
	})
	if err != nil {
		return nil, err
	}
	return transformersv1.ToYataiComponentSchema(ctx, comp)
}

type GetYataiComponentSchema struct {
	GetClusterSchema
	Type modelschemas.YataiComponentType `path:"componentType"`
}

func (c *yataiComponentController) Delete(ctx *gin.Context, schema *GetYataiComponentSchema) (*schemasv1.YataiComponentSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get cluster")
	}
	if err = c.canOperate(ctx, cluster); err != nil {
		return nil, err
	}
	comp, err := services.YataiComponentService.Delete(ctx, services.DeleteYataiComponentReleaseOption{
		ClusterId: cluster.ID,
		Type:      schema.Type,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create yatai component")
	}
	return transformersv1.ToYataiComponentSchema(ctx, comp)
}

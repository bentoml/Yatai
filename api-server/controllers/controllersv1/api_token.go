package controllersv1

import (
	"context"
	"time"

	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type apiTokenController struct {
	baseController
}

var ApiTokenController = apiTokenController{}

type GetApiTokenSchema struct {
	GetOrganizationSchema
	ApiTokenUid string `path:"apiTokenUid"`
}

func (s *GetApiTokenSchema) GetApiToken(ctx context.Context) (*models.ApiToken, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	apiToken, err := services.ApiTokenService.GetByUid(ctx, s.ApiTokenUid)
	if err != nil {
		return nil, errors.Wrapf(err, "get apiToken %s", s.ApiTokenUid)
	}
	if apiToken.UserId != user.ID {
		return nil, consts.ErrNotFound
	}
	return apiToken, nil
}

type CreateApiTokenSchema struct {
	schemasv1.CreateApiTokenSchema
	GetOrganizationSchema
}

func (c *apiTokenController) Create(ctx *gin.Context, schema *CreateApiTokenSchema) (*schemasv1.ApiTokenFullSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	apiToken, err := services.ApiTokenService.Create(ctx, services.CreateApiTokenOption{
		UserId:         user.ID,
		OrganizationId: org.ID,
		Name:           schema.Name,
		Description:    schema.Description,
		Scopes:         schema.Scopes,
		ExpiredAt:      schema.ExpiredAt,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create apiToken")
	}
	return transformersv1.ToApiTokenFullSchema(ctx, apiToken)
}

type UpdateApiTokenSchema struct {
	schemasv1.UpdateApiTokenSchema
	GetApiTokenSchema
}

func (c *apiTokenController) Update(ctx *gin.Context, schema *UpdateApiTokenSchema) (*schemasv1.ApiTokenSchema, error) {
	apiToken, err := schema.GetApiToken(ctx)
	if err != nil {
		return nil, err
	}
	var scopes **modelschemas.ApiTokenScopes
	if schema.Scopes != nil {
		scopes = &schema.Scopes
	}
	var expiredAt **time.Time
	if schema.ExpiredAt != nil {
		expiredAt = &schema.ExpiredAt
	}
	apiToken, err = services.ApiTokenService.Update(ctx, apiToken, services.UpdateApiTokenOption{
		Description: schema.Description,
		Scopes:      scopes,
		ExpiredAt:   expiredAt,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update apiToken")
	}
	return transformersv1.ToApiTokenSchema(ctx, apiToken)
}

func (c *apiTokenController) Get(ctx *gin.Context, schema *GetApiTokenSchema) (*schemasv1.ApiTokenSchema, error) {
	apiToken, err := schema.GetApiToken(ctx)
	if err != nil {
		return nil, err
	}
	return transformersv1.ToApiTokenSchema(ctx, apiToken)
}

func (c *apiTokenController) Delete(ctx *gin.Context, schema *GetApiTokenSchema) (*schemasv1.ApiTokenSchema, error) {
	apiToken, err := schema.GetApiToken(ctx)
	if err != nil {
		return nil, err
	}
	apiToken, err = services.ApiTokenService.Delete(ctx, apiToken)
	if err != nil {
		return nil, err
	}
	return transformersv1.ToApiTokenSchema(ctx, apiToken)
}

type ListApiTokenSchema struct {
	schemasv1.ListQuerySchema
	GetOrganizationSchema
}

func (c *apiTokenController) List(ctx *gin.Context, schema *ListApiTokenSchema) (*schemasv1.ApiTokenListSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	apiTokens, total, err := services.ApiTokenService.List(ctx, services.ListApiTokenOption{
		VisitorId:      utils.UintPtr(user.ID),
		OrganizationId: utils.UintPtr(org.ID),
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "list apiTokens")
	}

	apiTokenSchemas, err := transformersv1.ToApiTokenSchemas(ctx, apiTokens)
	return &schemasv1.ApiTokenListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: apiTokenSchemas,
	}, err
}

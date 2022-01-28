package controllersv1

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/huandu/xstrings"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type modelRepositoryController struct {
	baseController
}

var ModelRepositoryController = modelRepositoryController{}

type GetModelRepositorySchema struct {
	GetOrganizationSchema
	ModelRepositoryName string `path:"modelRepositoryName"`
}

func (s *GetModelRepositorySchema) GetModelRepository(ctx context.Context) (*models.ModelRepository, error) {
	organization, err := s.GetOrganization(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get organization")
	}
	modelRepository, err := services.ModelRepositoryService.GetByName(ctx, organization.ID, s.ModelRepositoryName)
	if err != nil {
		return nil, errors.Wrapf(err, "get modelRepository %s", s.ModelRepositoryName)
	}
	return modelRepository, nil
}

func (c *modelRepositoryController) canView(ctx context.Context, modelRepository *models.ModelRepository) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canView(ctx, organization)
}

func (c *modelRepositoryController) canUpdate(ctx context.Context, modelRepository *models.ModelRepository) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canUpdate(ctx, organization)
}

func (c *modelRepositoryController) canOperate(ctx context.Context, modelRepository *models.ModelRepository) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canOperate(ctx, organization)
}

type CreateModelRepositorySchema struct {
	schemasv1.CreateModelRepositorySchema
	GetOrganizationSchema
}

func (c *modelRepositoryController) Create(ctx *gin.Context, schema *CreateModelRepositorySchema) (*schemasv1.ModelRepositorySchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = OrganizationController.canUpdate(ctx, organization); err != nil {
		return nil, err
	}
	modelRepository, err := services.ModelRepositoryService.Create(ctx, services.CreateModelRepositoryOption{
		OrganizationId: organization.ID,
		CreatorId:      user.ID,
		Name:           schema.Name,
		Labels:         schema.Labels,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create modelRepository")
	}
	return transformersv1.ToModelRepositorySchema(ctx, modelRepository)
}

type UpdateModelRepositorySchema struct {
	schemasv1.UpdateModelRepositorySchema
	GetModelRepositorySchema
}

func (c *modelRepositoryController) Update(ctx *gin.Context, schema *UpdateModelRepositorySchema) (*schemasv1.ModelRepositorySchema, error) {
	modelRepository, err := schema.GetModelRepository(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, modelRepository); err != nil {
		return nil, err
	}
	modelRepository, err = services.ModelRepositoryService.Update(ctx, modelRepository, services.UpdateModelRepositoryOption{
		Description: schema.Description,
		Labels:      schema.Labels,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update modelRepository")
	}
	return transformersv1.ToModelRepositorySchema(ctx, modelRepository)
}

func (c *modelRepositoryController) Get(ctx *gin.Context, schema *GetModelRepositorySchema) (*schemasv1.ModelRepositorySchema, error) {
	modelRepository, err := schema.GetModelRepository(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, modelRepository); err != nil {
		return nil, err
	}
	return transformersv1.ToModelRepositorySchema(ctx, modelRepository)
}

type ListModelRepositorySchema struct {
	schemasv1.ListQuerySchema
	GetOrganizationSchema
}

func (c *modelRepositoryController) List(ctx *gin.Context, schema *ListModelRepositorySchema) (*schemasv1.ModelRepositoryListSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = OrganizationController.canView(ctx, organization); err != nil {
		return nil, err
	}

	listOpt := services.ListModelRepositoryOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		OrganizationId: utils.UintPtr(organization.ID),
	}

	queryMap := schema.Q.ToMap()
	for k, v := range queryMap {
		if k == schemasv1.KeyQIn {
			fieldNames := make([]string, 0, len(v.([]string)))
			for _, fieldName := range v.([]string) {
				if _, ok := map[string]struct{}{
					"name":        {},
					"description": {},
				}[fieldName]; !ok {
					continue
				}
				fieldNames = append(fieldNames, fieldName)
			}
			listOpt.KeywordFieldNames = &fieldNames
		}
		if k == schemasv1.KeyQKeywords {
			listOpt.Keywords = utils.StringSlicePtr(v.([]string))
		}
		if k == "creator" {
			userNames, err := processUserNamesFromQ(ctx, v.([]string))
			if err != nil {
				return nil, err
			}
			users, err := services.UserService.ListByNames(ctx, userNames)
			if err != nil {
				return nil, err
			}
			userIds := make([]uint, 0, len(users))
			for _, user := range users {
				userIds = append(userIds, user.ID)
			}
			listOpt.CreatorIds = utils.UintSlicePtr(userIds)
		}
		if k == "last_updater" {
			userNames, err := processUserNamesFromQ(ctx, v.([]string))
			if err != nil {
				return nil, err
			}
			users, err := services.UserService.ListByNames(ctx, userNames)
			if err != nil {
				return nil, err
			}
			userIds := make([]uint, 0, len(users))
			for _, user := range users {
				userIds = append(userIds, user.ID)
			}
			listOpt.LastUpdaterIds = utils.UintSlicePtr(userIds)
		}
		if k == "sort" {
			fieldName, _, order := xstrings.LastPartition(v.([]string)[0], "-")
			if _, ok := map[string]struct{}{
				"created_at": {},
				"updated_at": {},
			}[fieldName]; !ok {
				continue
			}
			if _, ok := map[string]struct{}{
				"desc": {},
				"asc":  {},
			}[order]; !ok {
				continue
			}
			if fieldName == "updated_at" {
				fieldName = "model.created_at"
			}
			listOpt.Order = utils.StringPtr(fmt.Sprintf("%s %s", fieldName, strings.ToUpper(order)))
		}
		if k == "label" {
			labelsSchema := services.ParseQueryLabelsToLabelsList(v.([]string))
			listOpt.LabelsList = &labelsSchema
		}
		if k == "-label" {
			labelsSchema := services.ParseQueryLabelsToLabelsList(v.([]string))
			listOpt.LackLabelsList = &labelsSchema
		}
	}

	modelRepositories, total, err := services.ModelRepositoryService.List(ctx, listOpt)
	if err != nil {
		return nil, errors.Wrap(err, "list modelRepositories")
	}

	modelRepositorySchemas, err := transformersv1.ToModelRepositorySchemas(ctx, modelRepositories)
	return &schemasv1.ModelRepositoryListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: modelRepositorySchemas,
	}, err
}

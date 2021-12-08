// nolint: goconst
package controllersv1

import (
	"context"
	"fmt"
	"strings"

	"github.com/huandu/xstrings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type bentoRepositoryController struct {
	baseController
}

var BentoRepositoryController = bentoRepositoryController{}

type GetBentoRepositorySchema struct {
	GetOrganizationSchema
	BentoRepositoryName string `path:"bentoRepositoryName"`
}

func (s *GetBentoRepositorySchema) GetBentoRepository(ctx context.Context) (*models.BentoRepository, error) {
	organization, err := s.GetOrganization(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get organization")
	}
	bentoRepository, err := services.BentoRepositoryService.GetByName(ctx, organization.ID, s.BentoRepositoryName)
	if err != nil {
		return nil, errors.Wrapf(err, "get bentoRepository %s", s.BentoRepositoryName)
	}
	return bentoRepository, nil
}

func (c *bentoRepositoryController) canView(ctx context.Context, bentoRepository *models.BentoRepository) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canView(ctx, organization)
}

func (c *bentoRepositoryController) canUpdate(ctx context.Context, bentoRepository *models.BentoRepository) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canUpdate(ctx, organization)
}

func (c *bentoRepositoryController) canOperate(ctx context.Context, bentoRepository *models.BentoRepository) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canOperate(ctx, organization)
}

type CreateBentoRepositorySchema struct {
	schemasv1.CreateBentoRepositorySchema
	GetOrganizationSchema
}

func (c *bentoRepositoryController) Create(ctx *gin.Context, schema *CreateBentoRepositorySchema) (*schemasv1.BentoRepositorySchema, error) {
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

	bentoRepository, err := services.BentoRepositoryService.Create(ctx, services.CreateBentoRepositoryOption{
		CreatorId:      user.ID,
		OrganizationId: organization.ID,
		Name:           schema.Name,
		Labels:         schema.Labels,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create bentoRepository")
	}
	return transformersv1.ToBentoRepositorySchema(ctx, bentoRepository)
}

type UpdateBentoRepositorySchema struct {
	schemasv1.UpdateBentoRepositorySchema
	GetBentoRepositorySchema
}

func (c *bentoRepositoryController) Update(ctx *gin.Context, schema *UpdateBentoRepositorySchema) (*schemasv1.BentoRepositorySchema, error) {
	bentoRepository, err := schema.GetBentoRepository(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, bentoRepository); err != nil {
		return nil, err
	}
	bentoRepository, err = services.BentoRepositoryService.Update(ctx, bentoRepository, services.UpdateBentoRepositoryOption{
		Description: schema.Description,
		Labels:      schema.Labels,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update bentoRepository")
	}
	return transformersv1.ToBentoRepositorySchema(ctx, bentoRepository)
}

func (c *bentoRepositoryController) Get(ctx *gin.Context, schema *GetBentoRepositorySchema) (*schemasv1.BentoRepositorySchema, error) {
	bentoRepository, err := schema.GetBentoRepository(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, bentoRepository); err != nil {
		return nil, err
	}
	return transformersv1.ToBentoRepositorySchema(ctx, bentoRepository)
}

type ListBentoRepositorySchema struct {
	schemasv1.ListQuerySchema
	GetOrganizationSchema
}

func processUserNamesFromQ(ctx context.Context, userNames []string) ([]string, error) {
	currentUser, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, len(userNames))
	for _, userName := range userNames {
		if userName == schemasv1.ValueQMe {
			userName = currentUser.Name
		}
		res = append(res, userName)
	}
	return res, nil
}

func (c *bentoRepositoryController) List(ctx *gin.Context, schema *ListBentoRepositorySchema) (*schemasv1.BentoRepositoryListSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = OrganizationController.canView(ctx, organization); err != nil {
		return nil, err
	}

	listOpt := services.ListBentoRepositoryOption{
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
				fieldName = "bento.created_at"
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

	bentoRepositories, total, err := services.BentoRepositoryService.List(ctx, listOpt)
	if err != nil {
		return nil, errors.Wrap(err, "list bentoRepositories")
	}

	bentoRepositorySchemas, err := transformersv1.ToBentoRepositorySchemas(ctx, bentoRepositories)
	return &schemasv1.BentoRepositoryListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: bentoRepositorySchemas,
	}, err
}

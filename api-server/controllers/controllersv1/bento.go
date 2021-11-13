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

type bentoController struct {
	baseController
}

var BentoController = bentoController{}

type GetBentoSchema struct {
	GetOrganizationSchema
	BentoName string `path:"bentoName"`
}

func (s *GetBentoSchema) GetBento(ctx context.Context) (*models.Bento, error) {
	organization, err := s.GetOrganization(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get organization")
	}
	bento, err := services.BentoService.GetByName(ctx, organization.ID, s.BentoName)
	if err != nil {
		return nil, errors.Wrapf(err, "get bento %s", s.BentoName)
	}
	return bento, nil
}

func (c *bentoController) canView(ctx context.Context, bento *models.Bento) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, bento)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canView(ctx, organization)
}

func (c *bentoController) canUpdate(ctx context.Context, bento *models.Bento) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, bento)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canUpdate(ctx, organization)
}

func (c *bentoController) canOperate(ctx context.Context, bento *models.Bento) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, bento)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canOperate(ctx, organization)
}

type CreateBentoSchema struct {
	schemasv1.CreateBentoSchema
	GetOrganizationSchema
}

func (c *bentoController) Create(ctx *gin.Context, schema *CreateBentoSchema) (*schemasv1.BentoSchema, error) {
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

	bento, err := services.BentoService.Create(ctx, services.CreateBentoOption{
		CreatorId:      user.ID,
		OrganizationId: organization.ID,
		Name:           schema.Name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create bento")
	}
	return transformersv1.ToBentoSchema(ctx, bento)
}

type UpdateBentoSchema struct {
	schemasv1.UpdateBentoSchema
	GetBentoSchema
}

func (c *bentoController) Update(ctx *gin.Context, schema *UpdateBentoSchema) (*schemasv1.BentoSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, bento); err != nil {
		return nil, err
	}
	bento, err = services.BentoService.Update(ctx, bento, services.UpdateBentoOption{
		Description: schema.Description,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update bento")
	}
	return transformersv1.ToBentoSchema(ctx, bento)
}

func (c *bentoController) Get(ctx *gin.Context, schema *GetBentoSchema) (*schemasv1.BentoSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, bento); err != nil {
		return nil, err
	}
	return transformersv1.ToBentoSchema(ctx, bento)
}

type ListBentoSchema struct {
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

func (c *bentoController) List(ctx *gin.Context, schema *ListBentoSchema) (*schemasv1.BentoListSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = OrganizationController.canView(ctx, organization); err != nil {
		return nil, err
	}

	listOpt := services.ListBentoOption{
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
				fieldName = "bento_version.created_at"
			}
			listOpt.Order = utils.StringPtr(fmt.Sprintf("%s %s", fieldName, strings.ToUpper(order)))
		}
	}

	bentos, total, err := services.BentoService.List(ctx, listOpt)
	if err != nil {
		return nil, errors.Wrap(err, "list bentos")
	}

	bentoSchemas, err := transformersv1.ToBentoSchemas(ctx, bentos)
	return &schemasv1.BentoListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: bentoSchemas,
	}, err
}

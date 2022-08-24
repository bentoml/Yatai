package controllersv1

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huandu/xstrings"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
)

type organizationController struct {
	baseController
}

var OrganizationController = organizationController{}

type GetOrganizationSchema struct {
	OrgName string `header:"X-Yatai-Organization"`
}

func (s *GetOrganizationSchema) GetOrganization(ctx context.Context) (*models.Organization, error) {
	return services.GetCurrentOrganization(ctx)
}

func (c *organizationController) canView(ctx context.Context, organization *models.Organization) error {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return err
	}
	return services.MemberService.CanView(ctx, &services.OrganizationMemberService, user, organization.ID)
}

func (c *organizationController) canUpdate(ctx context.Context, organization *models.Organization) error {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return err
	}
	return services.MemberService.CanUpdate(ctx, &services.OrganizationMemberService, user, organization.ID)
}

func (c *organizationController) canOperate(ctx context.Context, organization *models.Organization) error {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return err
	}
	return services.MemberService.CanOperate(ctx, &services.OrganizationMemberService, user, organization.ID)
}

func (c *organizationController) Create(ctx *gin.Context, schema *schemasv1.CreateOrganizationSchema) (*schemasv1.OrganizationFullSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	organization, err := services.OrganizationService.Create(ctx, services.CreateOrganizationOption{
		CreatorId:   user.ID,
		Name:        schema.Name,
		Description: schema.Description,
		Config:      schema.Config,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create organization")
	}
	_, err = services.OrganizationMemberService.Create(ctx, user.ID, services.CreateOrganizationMemberOption{
		CreatorId:      user.ID,
		OrganizationId: organization.ID,
		UserId:         user.ID,
		Role:           modelschemas.MemberRoleAdmin,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create organization member")
	}
	return transformersv1.ToOrganizationFullSchema(ctx, organization)
}

type UpdateOrganizationSchema struct {
	schemasv1.UpdateOrganizationSchema
	GetOrganizationSchema
}

func (c *organizationController) Update(ctx *gin.Context, schema *UpdateOrganizationSchema) (*schemasv1.OrganizationFullSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, organization); err != nil {
		return nil, err
	}
	organization, err = services.OrganizationService.Update(ctx, organization, services.UpdateOrganizationOption{
		Description: schema.Description,
		Config:      schema.Config,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update organization")
	}
	return transformersv1.ToOrganizationFullSchema(ctx, organization)
}

func (c *organizationController) Get(ctx *gin.Context, schema *GetOrganizationSchema) (*schemasv1.OrganizationFullSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, organization); err != nil {
		return nil, err
	}
	return transformersv1.ToOrganizationFullSchema(ctx, organization)
}

func (c *organizationController) GetMajorCluster(ctx *gin.Context, schema *GetOrganizationSchema) (*schemasv1.ClusterFullSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, organization); err != nil {
		return nil, err
	}
	cluster, err := services.OrganizationService.GetMajorCluster(ctx, organization)
	if err != nil {
		return nil, errors.Wrap(err, "get major cluster")
	}
	return transformersv1.ToClusterFullSchema(ctx, cluster)
}

type ListEventOperationNames struct {
	GetOrganizationSchema
	ResourceType modelschemas.ResourceType `query:"resource_type"`
}

func (c *organizationController) ListEventOperationNames(ctx *gin.Context, schema *ListEventOperationNames) ([]string, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, organization); err != nil {
		return nil, err
	}
	return services.EventService.ListOperationNames(ctx, organization.ID, schema.ResourceType)
}

type ListOrginizationEventsSchema struct {
	schemasv1.ListQuerySchema
	GetOrganizationSchema
}

func (c *organizationController) ListEvents(ctx *gin.Context, schema *ListOrginizationEventsSchema) (*schemasv1.EventListSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, organization); err != nil {
		return nil, err
	}

	listOpt := services.ListEventOption{
		BaseListOption: services.BaseListOption{
			Start: &schema.Start,
			Count: &schema.Count,
		},
		OrganizationId: &organization.ID,
		Status:         modelschemas.EventStatusSuccess.Ptr(),
	}

	queryMap := schema.Q.ToMap()
	for k, v := range queryMap {
		if k == "status" {
			listOpt.Status = modelschemas.EventStatus(v.([]string)[0]).Ptr()
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
		if k == "resource_type" {
			listOpt.ResourceType = modelschemas.ResourceType(v.([]string)[0]).Ptr()
		}
		if k == "started_at" {
			startedAtStr := v.([]string)[0]
			startedAt, err := time.Parse("2006-01-02", startedAtStr)
			if err != nil {
				return nil, errors.Wrap(err, "parse started_at")
			}
			listOpt.StartedAt = &startedAt
		}
		if k == "ended_at" {
			endedAtStr := v.([]string)[0]
			endedAt, err := time.Parse("2006-01-02", endedAtStr)
			if err != nil {
				return nil, errors.Wrap(err, "parse ended_at")
			}
			listOpt.EndedAt = &endedAt
		}
		if k == "sort" {
			fieldName, _, order := xstrings.LastPartition(v.([]string)[0], "-")
			if _, ok := map[string]struct{}{
				"created_at": {},
			}[fieldName]; !ok {
				continue
			}
			if _, ok := map[string]struct{}{
				"desc": {},
				"asc":  {},
			}[order]; !ok {
				continue
			}
			listOpt.Order = utils.StringPtr(fmt.Sprintf("event.%s %s", fieldName, strings.ToUpper(order)))
		}
	}

	events, total, err := services.EventService.List(ctx, listOpt)
	if err != nil {
		return nil, errors.Wrap(err, "list events")
	}
	eventSchemas, err := transformersv1.ToEventSchemas(ctx, events)
	if err != nil {
		return nil, errors.Wrap(err, "transform events")
	}
	return &schemasv1.EventListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: eventSchemas,
	}, nil
}

func (c *organizationController) ListModelModules(ctx *gin.Context, schema *GetOrganizationSchema) ([]string, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, organization); err != nil {
		return nil, err
	}
	modules, err := services.ModelService.ListAllModules(ctx, organization.ID)
	return modules, err
}

func (c *organizationController) List(ctx *gin.Context, schema *schemasv1.ListQuerySchema) (*schemasv1.OrganizationListSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	organizations, total, err := services.OrganizationService.List(ctx, services.ListOrganizationOption{BaseListOption: services.BaseListOption{
		Start:  utils.UintPtr(schema.Start),
		Count:  utils.UintPtr(schema.Count),
		Search: schema.Search,
	},
		VisitorId: utils.UintPtr(user.ID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list organizations")
	}

	organizationSchemas, err := transformersv1.ToOrganizationSchemas(ctx, organizations)
	return &schemasv1.OrganizationListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: organizationSchemas,
	}, err
}

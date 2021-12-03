package controllersv1

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/huandu/xstrings"

	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type bentoVersionController struct {
	baseController
}

var BentoVersionController = bentoVersionController{}

type GetBentoVersionSchema struct {
	GetBentoSchema
	Version string `path:"version"`
}

func (s *GetBentoVersionSchema) GetBentoVersion(ctx context.Context) (*models.BentoVersion, error) {
	bento, err := s.GetBento(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "get bento %s", bento.Name)
	}
	version, err := services.BentoVersionService.GetByVersion(ctx, bento.ID, s.Version)
	if err != nil {
		return nil, errors.Wrapf(err, "get bento version %s", s.Version)
	}
	return version, nil
}

func (c *bentoVersionController) canView(ctx context.Context, version *models.BentoVersion) error {
	bento, err := services.BentoService.GetAssociatedBento(ctx, version)
	if err != nil {
		return errors.Wrap(err, "get associated bento")
	}
	return BentoController.canView(ctx, bento)
}

func (c *bentoVersionController) canUpdate(ctx context.Context, version *models.BentoVersion) error {
	bento, err := services.BentoService.GetAssociatedBento(ctx, version)
	if err != nil {
		return errors.Wrap(err, "get associated bento")
	}
	return BentoController.canUpdate(ctx, bento)
}

func (c *bentoVersionController) canOperate(ctx context.Context, version *models.BentoVersion) error {
	bento, err := services.BentoService.GetAssociatedBento(ctx, version)
	if err != nil {
		return errors.Wrap(err, "get associated bento")
	}
	return BentoController.canOperate(ctx, bento)
}

type CreateBentoVersionSchema struct {
	schemasv1.CreateBentoVersionSchema
	GetBentoSchema
}

func (c *bentoVersionController) Create(ctx *gin.Context, schema *CreateBentoVersionSchema) (*schemasv1.BentoVersionSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = BentoController.canUpdate(ctx, bento); err != nil {
		return nil, err
	}
	buildAt, err := time.Parse("2006-01-02 15:04:05.000000", schema.BuildAt)
	if err != nil {
		return nil, errors.Wrapf(err, "parse build_at")
	}
	version, err := services.BentoVersionService.Create(ctx, services.CreateBentoVersionOption{
		CreatorId:   user.ID,
		BentoId:     bento.ID,
		Version:     schema.Version,
		Description: schema.Description,
		Manifest:    schema.Manifest,
		BuildAt:     buildAt,
		Labels:      schema.Labels,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create version")
	}
	return transformersv1.ToBentoVersionSchema(ctx, version)
}

type UpdateBentoVersionSchema struct {
	schemasv1.UpdateBentoVersionSchema
	GetBentoVersionSchema
}

func (c *bentoVersionController) Update(ctx *gin.Context, schema *UpdateBentoVersionSchema) (*schemasv1.BentoVersionSchema, error) {
	version, err := schema.GetBentoVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, version); err != nil {
		return nil, err
	}
	version, err = services.BentoVersionService.Update(ctx, version, services.UpdateBentoVersionOption{
		Labels: schema.Labels,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update bentoVersion")
	}
	return transformersv1.ToBentoVersionSchema(ctx, version)
}

func (c *bentoVersionController) PreSignS3UploadUrl(ctx *gin.Context, schema *GetBentoVersionSchema) (*schemasv1.BentoVersionSchema, error) {
	version, err := schema.GetBentoVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, version); err != nil {
		return nil, err
	}
	url, err := services.BentoVersionService.PreSignS3UploadUrl(ctx, version)
	if err != nil {
		return nil, errors.Wrap(err, "pre sign s3 upload url")
	}
	bentoVersionSchema, err := transformersv1.ToBentoVersionSchema(ctx, version)
	if err != nil {
		return nil, err
	}
	bentoVersionSchema.PresignedS3Url = url.String()
	return bentoVersionSchema, nil
}

func (c *bentoVersionController) PreSignS3DownloadUrl(ctx *gin.Context, schema *GetBentoVersionSchema) (*schemasv1.BentoVersionSchema, error) {
	version, err := schema.GetBentoVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, version); err != nil {
		return nil, err
	}
	url, err := services.BentoVersionService.PreSignS3DownloadUrl(ctx, version)
	if err != nil {
		return nil, errors.Wrap(err, "pre sign s3 upload url")
	}
	bentoVersionSchema, err := transformersv1.ToBentoVersionSchema(ctx, version)
	if err != nil {
		return nil, err
	}
	bentoVersionSchema.PresignedS3Url = url.String()
	return bentoVersionSchema, nil
}

func (c *bentoVersionController) StartUpload(ctx *gin.Context, schema *GetBentoVersionSchema) (*schemasv1.BentoVersionSchema, error) {
	version, err := schema.GetBentoVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, version); err != nil {
		return nil, err
	}
	uploadStatus := modelschemas.BentoVersionUploadStatusUploading
	now := time.Now()
	nowPtr := &now
	version, err = services.BentoVersionService.Update(ctx, version, services.UpdateBentoVersionOption{
		UploadStatus:    &uploadStatus,
		UploadStartedAt: &nowPtr,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update version")
	}
	return transformersv1.ToBentoVersionSchema(ctx, version)
}

type FinishUploadBentoVersionSchema struct {
	schemasv1.FinishUploadBentoVersionSchema
	GetBentoVersionSchema
}

func (c *bentoVersionController) FinishUpload(ctx *gin.Context, schema *FinishUploadBentoVersionSchema) (*schemasv1.BentoVersionSchema, error) {
	version, err := schema.GetBentoVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, version); err != nil {
		return nil, err
	}
	now := time.Now()
	nowPtr := &now
	version, err = services.BentoVersionService.Update(ctx, version, services.UpdateBentoVersionOption{
		UploadStatus:         schema.Status,
		UploadFinishedAt:     &nowPtr,
		UploadFinishedReason: schema.Reason,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update version")
	}
	return transformersv1.ToBentoVersionSchema(ctx, version)
}

func (c *bentoVersionController) Get(ctx *gin.Context, schema *GetBentoVersionSchema) (*schemasv1.BentoVersionSchema, error) {
	version, err := schema.GetBentoVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, version); err != nil {
		return nil, err
	}
	return transformersv1.ToBentoVersionSchema(ctx, version)
}

type ListBentoVersionSchema struct {
	schemasv1.ListQuerySchema
	GetBentoSchema
}

func (c *bentoVersionController) List(ctx *gin.Context, schema *ListBentoVersionSchema) (*schemasv1.BentoVersionListSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}

	if err = BentoController.canView(ctx, bento); err != nil {
		return nil, err
	}

	bentos, total, err := services.BentoVersionService.List(ctx, services.ListBentoVersionOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		BentoId: utils.UintPtr(bento.ID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list bentos")
	}

	bentoSchemas, err := transformersv1.ToBentoVersionSchemas(ctx, bentos)
	return &schemasv1.BentoVersionListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: bentoSchemas,
	}, err
}

type ListAllBentoVersionSchema struct {
	schemasv1.ListQuerySchema
	GetOrganizationSchema
}

func (c *bentoVersionController) ListAll(ctx *gin.Context, schema *ListAllBentoVersionSchema) (*schemasv1.BentoVersionWithBentoListSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = OrganizationController.canView(ctx, organization); err != nil {
		return nil, err
	}

	listOpt := services.ListBentoVersionOption{
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
		if k == "sort" {
			fieldName, _, order := xstrings.LastPartition(v.([]string)[0], "-")
			if _, ok := map[string]struct{}{
				"created_at": {},
				"build_at":   {},
			}[fieldName]; !ok {
				continue
			}
			if _, ok := map[string]struct{}{
				"desc": {},
				"asc":  {},
			}[order]; !ok {
				continue
			}
			listOpt.Order = utils.StringPtr(fmt.Sprintf("bento_version.%s %s", fieldName, strings.ToUpper(order)))
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
	bentos, total, err := services.BentoVersionService.List(ctx, listOpt)
	if err != nil {
		return nil, errors.Wrap(err, "list bentos")
	}

	bentoSchemas, err := transformersv1.ToBentoVersionWithBentoSchemas(ctx, bentos)
	return &schemasv1.BentoVersionWithBentoListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: bentoSchemas,
	}, err
}

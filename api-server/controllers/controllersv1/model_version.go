package controllersv1

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/huandu/xstrings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/modelschemas"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type modelVersionController struct {
	baseController
}

var ModelVersionController = modelVersionController{}

type GetModelVersionSchema struct {
	GetModelSchema
	Version string `path:"version"`
}

func (s *GetModelVersionSchema) GetModelVersion(ctx context.Context) (*models.ModelVersion, error) {
	model, err := s.GetModel(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "get model %s", model.Name)
	}
	version, err := services.ModelVersionService.GetByVersion(ctx, model.ID, s.Version)
	if err != nil {
		return nil, errors.Wrapf(err, "get model version %s", s.Version)
	}
	return version, nil
}

func (c *modelVersionController) canView(ctx context.Context, version *models.ModelVersion) error {
	model, err := services.ModelService.GetAssociatedModel(ctx, version)
	if err != nil {
		return errors.Wrap(err, "get associated model")
	}
	return ModelController.canView(ctx, model)
}

func (c *modelVersionController) canUpdate(ctx context.Context, version *models.ModelVersion) error {
	model, err := services.ModelService.GetAssociatedModel(ctx, version)
	if err != nil {
		return errors.Wrap(err, "get associated model")
	}
	return ModelController.canUpdate(ctx, model)
}

func (c *modelVersionController) canOperate(ctx context.Context, version *models.ModelVersion) error {
	model, err := services.ModelService.GetAssociatedModel(ctx, version)
	if err != nil {
		return errors.Wrap(err, "get associated model")
	}
	return ModelController.canOperate(ctx, model)
}

type CreateModelVersionSchema struct {
	schemasv1.CreateModelVersionSchema
	GetModelSchema
}

func (c *modelVersionController) Create(ctx *gin.Context, schema *CreateModelVersionSchema) (*schemasv1.ModelVersionSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	model, err := schema.GetModel(ctx)
	if err != nil {
		return nil, err
	}
	if err = ModelController.canUpdate(ctx, model); err != nil {
		return nil, err
	}
	buildAt, err := time.Parse("2006-01-02 15:04:05.000000", schema.BuildAt)
	if err != nil {
		return nil, errors.Wrap(err, "parse build at")
	}
	version, err := services.ModelVersionService.Create(ctx, services.CreateModelVersionOption{
		CreatorId:   user.ID,
		ModelId:     model.ID,
		Version:     schema.Version,
		Description: schema.Description,
		Manifest:    schema.Manifest,
		BuildAt:     buildAt,
		Labels:      schema.Labels,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create model version")
	}
	return transformersv1.ToModelVersionSchema(ctx, version)
}

func (c *modelVersionController) PreSignS3UploadUrl(ctx *gin.Context, schema *GetModelVersionSchema) (*schemasv1.ModelVersionSchema, error) {
	version, err := schema.GetModelVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, version); err != nil {
		return nil, err
	}
	url, err := services.ModelVersionService.PreSignS3UploadUrl(ctx, version)
	if err != nil {
		return nil, errors.Wrap(err, "pre sign s3 upload url")
	}
	bentoVersionSchema, err := transformersv1.ToModelVersionSchema(ctx, version)
	if err != nil {
		return nil, err
	}
	bentoVersionSchema.PresignedS3Url = url.String()
	return bentoVersionSchema, nil
}

func (c *modelVersionController) PreSignS3DownloadUrl(ctx *gin.Context, schema *GetModelVersionSchema) (*schemasv1.ModelVersionSchema, error) {
	version, err := schema.GetModelVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, version); err != nil {
		return nil, err
	}
	url, err := services.ModelVersionService.PreSignS3DownloadUrl(ctx, version)
	if err != nil {
		return nil, errors.Wrap(err, "pre sign s3 download url")
	}
	bentoVersionSchema, err := transformersv1.ToModelVersionSchema(ctx, version)
	if err != nil {
		return nil, err
	}
	bentoVersionSchema.PresignedS3Url = url.String()
	return bentoVersionSchema, nil
}

type UpdateModelVersionSchema struct {
	schemasv1.UpdateModelVersionSchema
	GetModelVersionSchema
}

func (c *modelVersionController) Update(ctx *gin.Context, schema *UpdateModelVersionSchema) (*schemasv1.ModelVersionSchema, error) {
	modelVersion, err := schema.GetModelVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, modelVersion); err != nil {
		return nil, err
	}
	modelVersion, err = services.ModelVersionService.Update(ctx, modelVersion, services.UpdateModelVersionOption{
		Labels: schema.Labels,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Update modelVersion")
	}
	return transformersv1.ToModelVersionSchema(ctx, modelVersion)
}

func (c *modelVersionController) StartUpload(ctx *gin.Context, schema *GetModelVersionSchema) (*schemasv1.ModelVersionSchema, error) {
	version, err := schema.GetModelVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, version); err != nil {
		return nil, err
	}
	uploadStatus := modelschemas.ModelVersionUploadStatusUploading
	now := time.Now()
	nowPtr := &now
	version, err = services.ModelVersionService.Update(ctx, version, services.UpdateModelVersionOption{
		UploadStatus:    &uploadStatus,
		UploadStartedAt: &nowPtr,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update version")
	}
	return transformersv1.ToModelVersionSchema(ctx, version)
}

type FinishUploadModelVersionSchema struct {
	schemasv1.FinishUploadModelVersionSchema
	GetModelVersionSchema
}

func (c *modelVersionController) FinishUpload(ctx *gin.Context, schema *FinishUploadModelVersionSchema) (*schemasv1.ModelVersionSchema, error) {
	version, err := schema.GetModelVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, version); err != nil {
		return nil, err
	}
	now := time.Now()
	nowPtr := &now
	version, err = services.ModelVersionService.Update(ctx, version, services.UpdateModelVersionOption{
		UploadStatus:         schema.Status,
		UploadFinishedAt:     &nowPtr,
		UploadFinishedReason: schema.Reason,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update version")
	}
	return transformersv1.ToModelVersionSchema(ctx, version)
}

func (c *modelVersionController) Get(ctx *gin.Context, schema *GetModelVersionSchema) (*schemasv1.ModelVersionSchema, error) {
	version, err := schema.GetModelVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, version); err != nil {
		return nil, err
	}
	return transformersv1.ToModelVersionSchema(ctx, version)
}

type ListModelVersionSchema struct {
	schemasv1.ListQuerySchema
	GetModelSchema
}

func (c *modelVersionController) List(ctx *gin.Context, schema *ListModelVersionSchema) (*schemasv1.ModelVersionListSchema, error) {
	model, err := schema.GetModel(ctx)
	if err != nil {
		return nil, err
	}
	if err = ModelController.canView(ctx, model); err != nil {
		return nil, err
	}

	models_, total, err := services.ModelVersionService.List(ctx, services.ListModelVersionOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		ModelId: utils.UintPtr(model.ID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list models")
	}

	modelSchemas, err := transformersv1.ToModelVersionSchemas(ctx, models_)
	return &schemasv1.ModelVersionListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: modelSchemas,
	}, err
}

type ListAllModelVersionSchema struct {
	schemasv1.ListQuerySchema
	GetOrganizationSchema
}

func (c *modelVersionController) ListAll(ctx *gin.Context, schema *ListAllModelVersionSchema) (*schemasv1.ModelVersionWithModelListSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = OrganizationController.canView(ctx, organization); err != nil {
		return nil, err
	}

	listOpt := services.ListModelVersionOption{
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
			listOpt.Order = utils.StringPtr(fmt.Sprintf("model_version.%s %s", fieldName, strings.ToUpper(order)))
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
	models_, total, err := services.ModelVersionService.List(ctx, listOpt)
	if err != nil {
		return nil, errors.Wrap(err, "list models")
	}

	modelSchemas, err := transformersv1.ToModelVersionWithModelSchemas(ctx, models_)
	return &schemasv1.ModelVersionWithModelListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: modelSchemas,
	}, err
}

package controllersv1

import (
	"context"
	"time"

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
	org, err := services.OrganizationService.GetAssociatedOrganization(ctx, model)
	if err != nil {
		return nil, err
	}
	label, err := LabelController.GetLabel(ctx); err != nil {
		return nil, err
	}
	if err = LabelController.canUpdate(ctx, label)
	err != nil {
		label, err = services.LabelService.Create(ctx, services.CreateLabelOption{
			OrganizationId: org.ID,
			CreatorId: user.ID,
			Key: schema.LabelKey,
			Value: schema.LabelValue
		}) 
	}
	else {
		label, err = services.LabelService.Update(ctx, services.UpdateLabelOption{
			OrganizationId: org.ID,
			CreatorId: user.ID,
			Value: schema.LabelValue,
		})
	}
	if err != nil {
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
		LabelKey: label.LabelKey,
		LabelValue: label.LabelValue
	})
	if err != nil {
		return nil, errors.Wrap(err, "create model version")
	}
	return transformersv1.ToModelVersionSchema(ctx, version)
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

	models, total, err := services.ModelVersionService.List(ctx, services.ListModelVersionOption{
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

	modelSchemas, err := transformersv1.ToModelVersionSchemas(ctx, models)
	return &schemasv1.ModelVersionListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: modelSchemas,
	}, err
}

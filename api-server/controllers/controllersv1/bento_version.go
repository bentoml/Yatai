package controllersv1

import (
	"context"
	"time"

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
	})
	if err != nil {
		return nil, errors.Wrap(err, "create version")
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

package controllersv1

import (
	"context"
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type bundleVersionController struct {
	baseController
}

var BundleVersionController = bundleVersionController{}

type GetBundleVersionSchema struct {
	GetBundleSchema
	Version string `path:"version"`
}

func (s *GetBundleVersionSchema) GetBundleVersion(ctx context.Context) (*models.BundleVersion, error) {
	bundle, err := s.GetBundle(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "get bundle %s", bundle.Name)
	}
	version, err := services.BundleVersionService.GetByVersion(ctx, bundle.ID, s.Version)
	if err != nil {
		return nil, errors.Wrapf(err, "get bundle version %s", s.Version)
	}
	return version, nil
}

func (c *bundleVersionController) canView(ctx context.Context, version *models.BundleVersion) error {
	bundle, err := services.BundleService.GetAssociatedBundle(ctx, version)
	if err != nil {
		return errors.Wrap(err, "get associated bundle")
	}
	return BundleController.canView(ctx, bundle)
}

func (c *bundleVersionController) canUpdate(ctx context.Context, version *models.BundleVersion) error {
	bundle, err := services.BundleService.GetAssociatedBundle(ctx, version)
	if err != nil {
		return errors.Wrap(err, "get associated bundle")
	}
	return BundleController.canUpdate(ctx, bundle)
}

func (c *bundleVersionController) canOperate(ctx context.Context, version *models.BundleVersion) error {
	bundle, err := services.BundleService.GetAssociatedBundle(ctx, version)
	if err != nil {
		return errors.Wrap(err, "get associated bundle")
	}
	return BundleController.canOperate(ctx, bundle)
}

type CreateBundleVersionSchema struct {
	schemasv1.CreateBundleVersionSchema
	GetBundleSchema
}

func (c *bundleVersionController) Create(ctx *gin.Context, schema *CreateBundleVersionSchema) (*schemasv1.BundleVersionSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	bundle, err := schema.GetBundle(ctx)
	if err != nil {
		return nil, err
	}
	if err = BundleController.canUpdate(ctx, bundle); err != nil {
		return nil, err
	}
	version, err := services.BundleVersionService.Create(ctx, services.CreateBundleVersionOption{
		CreatorId:   user.ID,
		BundleId:    bundle.ID,
		Version:     schema.Version,
		Description: schema.Description,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create version")
	}
	return transformersv1.ToBundleVersionSchema(ctx, version)
}

func (c *bundleVersionController) StartUpload(ctx *gin.Context, schema *GetBundleVersionSchema) (*schemasv1.BundleVersionSchema, error) {
	version, err := schema.GetBundleVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, version); err != nil {
		return nil, err
	}
	uploadStatus := modelschemas.BundleVersionUploadStatusUploading
	now := time.Now()
	nowPtr := &now
	version, err = services.BundleVersionService.Update(ctx, version, services.UpdateBundleVersionOption{
		UploadStatus:    &uploadStatus,
		UploadStartedAt: &nowPtr,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update version")
	}
	return transformersv1.ToBundleVersionSchema(ctx, version)
}

type FinishUploadBundleVersionSchema struct {
	schemasv1.FinishUploadBundleVersionSchema
	GetBundleVersionSchema
}

func (c *bundleVersionController) FinishUpload(ctx *gin.Context, schema *FinishUploadBundleVersionSchema) (*schemasv1.BundleVersionSchema, error) {
	version, err := schema.GetBundleVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, version); err != nil {
		return nil, err
	}
	now := time.Now()
	nowPtr := &now
	version, err = services.BundleVersionService.Update(ctx, version, services.UpdateBundleVersionOption{
		UploadStatus:         schema.Status,
		UploadFinishedAt:     &nowPtr,
		UploadFinishedReason: schema.Reason,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update version")
	}
	return transformersv1.ToBundleVersionSchema(ctx, version)
}

func (c *bundleVersionController) Get(ctx *gin.Context, schema *GetBundleVersionSchema) (*schemasv1.BundleVersionSchema, error) {
	version, err := schema.GetBundleVersion(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, version); err != nil {
		return nil, err
	}
	return transformersv1.ToBundleVersionSchema(ctx, version)
}

type ListBundleVersionSchema struct {
	schemasv1.ListQuerySchema
	GetBundleSchema
}

func (c *bundleVersionController) List(ctx *gin.Context, schema *ListBundleVersionSchema) (*schemasv1.BundleVersionListSchema, error) {
	bundle, err := schema.GetBundle(ctx)
	if err != nil {
		return nil, err
	}

	if err = BundleController.canView(ctx, bundle); err != nil {
		return nil, err
	}

	bundles, total, err := services.BundleVersionService.List(ctx, services.ListBundleVersionOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		BundleId: utils.UintPtr(bundle.ID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list bundles")
	}

	bundleSchemas, err := transformersv1.ToBundleVersionSchemas(ctx, bundles)
	return &schemasv1.BundleVersionListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: bundleSchemas,
	}, err
}

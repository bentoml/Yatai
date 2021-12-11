// nolint: goconst
package controllersv1

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

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

type bentoController struct {
	baseController
}

var BentoController = bentoController{}

type GetBentoSchema struct {
	GetBentoRepositorySchema
	Version string `path:"version"`
}

func (s *GetBentoSchema) GetBento(ctx context.Context) (*models.Bento, error) {
	bentoRepository, err := s.GetBentoRepository(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "get bentoRepository %s", bentoRepository.Name)
	}
	bento, err := services.BentoService.GetByVersion(ctx, bentoRepository.ID, s.Version)
	if err != nil {
		return nil, errors.Wrapf(err, "get bentoRepository %s bento %s", bentoRepository.Name, s.Version)
	}
	return bento, nil
}

func (c *bentoController) canView(ctx context.Context, bento *models.Bento) error {
	bentoRepository, err := services.BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return errors.Wrap(err, "get associated bentoRepository")
	}
	return BentoRepositoryController.canView(ctx, bentoRepository)
}

func (c *bentoController) canUpdate(ctx context.Context, bento *models.Bento) error {
	bentoRepository, err := services.BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return errors.Wrap(err, "get associated bentoRepository")
	}
	return BentoRepositoryController.canUpdate(ctx, bentoRepository)
}

func (c *bentoController) canOperate(ctx context.Context, bento *models.Bento) error {
	bentoRepository, err := services.BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return errors.Wrap(err, "get associated bentoRepository")
	}
	return BentoRepositoryController.canOperate(ctx, bentoRepository)
}

type CreateBentoSchema struct {
	schemasv1.CreateBentoSchema
	GetBentoRepositorySchema
}

func (c *bentoController) Create(ctx *gin.Context, schema *CreateBentoSchema) (*schemasv1.BentoSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	bentoRepository, err := schema.GetBentoRepository(ctx)
	if err != nil {
		return nil, err
	}
	if err = BentoRepositoryController.canUpdate(ctx, bentoRepository); err != nil {
		return nil, err
	}
	buildAt, err := time.Parse("2006-01-02 15:04:05.000000", schema.BuildAt)
	if err != nil {
		return nil, errors.Wrapf(err, "parse build_at")
	}
	bento, err := services.BentoService.Create(ctx, services.CreateBentoOption{
		CreatorId:         user.ID,
		BentoRepositoryId: bentoRepository.ID,
		Version:           schema.Version,
		Description:       schema.Description,
		Manifest:          schema.Manifest,
		BuildAt:           buildAt,
		Labels:            schema.Labels,
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
		Labels: schema.Labels,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update bento")
	}
	return transformersv1.ToBentoSchema(ctx, bento)
}

func (c *bentoController) PreSignUploadUrl(ctx *gin.Context, schema *GetBentoSchema) (*schemasv1.BentoSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, bento); err != nil {
		return nil, err
	}
	url, err := services.BentoService.PreSignUploadUrl(ctx, bento)
	if err != nil {
		return nil, errors.Wrap(err, "pre sign upload url")
	}
	bentoSchema, err := transformersv1.ToBentoSchema(ctx, bento)
	if err != nil {
		return nil, err
	}
	bentoSchema.PresignedUploadUrl = url.String()
	return bentoSchema, nil
}

func (c *bentoController) PreSignDownloadUrl(ctx *gin.Context, schema *GetBentoSchema) (*schemasv1.BentoSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, bento); err != nil {
		return nil, err
	}
	url, err := services.BentoService.PreSignDownloadUrl(ctx, bento)
	if err != nil {
		return nil, errors.Wrap(err, "pre sign s3 upload url")
	}
	bentoSchema, err := transformersv1.ToBentoSchema(ctx, bento)
	if err != nil {
		return nil, err
	}
	bentoSchema.PresignedDownloadUrl = url.String()
	return bentoSchema, nil
}

func (c *bentoController) StartUpload(ctx *gin.Context, schema *GetBentoSchema) (*schemasv1.BentoSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, bento); err != nil {
		return nil, err
	}
	uploadStatus := modelschemas.BentoUploadStatusUploading
	now := time.Now()
	nowPtr := &now
	bento, err = services.BentoService.Update(ctx, bento, services.UpdateBentoOption{
		UploadStatus:    &uploadStatus,
		UploadStartedAt: &nowPtr,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update bento")
	}
	return transformersv1.ToBentoSchema(ctx, bento)
}

type FinishUploadBentoSchema struct {
	schemasv1.FinishUploadBentoSchema
	GetBentoSchema
}

func (c *bentoController) FinishUpload(ctx *gin.Context, schema *FinishUploadBentoSchema) (*schemasv1.BentoSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, bento); err != nil {
		return nil, err
	}
	now := time.Now()
	nowPtr := &now
	bento, err = services.BentoService.Update(ctx, bento, services.UpdateBentoOption{
		UploadStatus:         schema.Status,
		UploadFinishedAt:     &nowPtr,
		UploadFinishedReason: schema.Reason,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update bento")
	}
	return transformersv1.ToBentoSchema(ctx, bento)
}

func (c *bentoController) RecreateImageBuilderJob(ctx *gin.Context, schema *GetBentoSchema) (*schemasv1.BentoSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, bento); err != nil {
		return nil, err
	}
	models_, err := services.BentoService.ListModelsFromManifests(ctx, bento)
	if err != nil {
		return nil, err
	}
	for _, model := range models_ {
		model := model
		go func() {
			_, err := services.ModelService.CreateImageBuilderJob(ctx, model)
			if err != nil {
				logrus.Errorf("failed to create image builder job for model %s: %v", model.Version, err)
			}
		}()
	}
	bento, err = services.BentoService.CreateImageBuilderJob(ctx, bento)
	if err != nil {
		return nil, err
	}
	return transformersv1.ToBentoSchema(ctx, bento)
}

func (c *bentoController) ListImageBuilderPods(ctx *gin.Context, schema *GetBentoSchema) ([]*schemasv1.KubePodSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, bento); err != nil {
		return nil, err
	}
	pods, err := services.BentoService.ListImageBuilderPods(ctx, bento)
	if err != nil {
		return nil, err
	}
	return transformersv1.ToKubePodSchemas(ctx, pods)
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
	GetBentoRepositorySchema
}

func (c *bentoController) List(ctx *gin.Context, schema *ListBentoSchema) (*schemasv1.BentoListSchema, error) {
	bentoRepository, err := schema.GetBentoRepository(ctx)
	if err != nil {
		return nil, err
	}

	if err = BentoRepositoryController.canView(ctx, bentoRepository); err != nil {
		return nil, err
	}

	bentos, total, err := services.BentoService.List(ctx, services.ListBentoOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		BentoRepositoryId: utils.UintPtr(bentoRepository.ID),
	})
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

type ListAllBentoSchema struct {
	schemasv1.ListQuerySchema
	GetOrganizationSchema
}

func (c *bentoController) ListAll(ctx *gin.Context, schema *ListAllBentoSchema) (*schemasv1.BentoWithRepositoryListSchema, error) {
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
			listOpt.Order = utils.StringPtr(fmt.Sprintf("bento.%s %s", fieldName, strings.ToUpper(order)))
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
	bentos, total, err := services.BentoService.List(ctx, listOpt)
	if err != nil {
		return nil, errors.Wrap(err, "list bentos")
	}

	bentoSchemas, err := transformersv1.ToBentoWithRepositorySchemas(ctx, bentos)
	return &schemasv1.BentoWithRepositoryListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: bentoSchemas,
	}, err
}

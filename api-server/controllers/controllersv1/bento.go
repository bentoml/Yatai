// nolint: goconst
package controllersv1

import (
	"context"
	"fmt"
	"strings"
	"time"

	pep440version "github.com/aquasecurity/go-pep440-version"
	"github.com/gin-gonic/gin"
	"github.com/huandu/xstrings"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
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
		return nil, errors.Wrapf(err, "get bentoRepository %s", s.BentoRepositoryName)
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
		Labels:   schema.Labels,
		Manifest: schema.Manifest,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update bento")
	}
	return transformersv1.ToBentoSchema(ctx, bento)
}

func abortWithError(ctx *gin.Context, err error) {
	ctx.AbortWithStatusJSON(500, map[string]string{
		"error": err.Error(),
	})
}

func (c *bentoController) Upload(ctx *gin.Context) {
	schema := GetBentoSchema{
		GetBentoRepositorySchema: GetBentoRepositorySchema{
			BentoRepositoryName: ctx.Param("bentoRepositoryName"),
		},
		Version: ctx.Param("version"),
	}

	bento, err := schema.GetBento(ctx)
	if err != nil {
		abortWithError(ctx, err)
		return
	}

	if err = c.canUpdate(ctx, bento); err != nil {
		abortWithError(ctx, err)
		return
	}

	uploadStatus := modelschemas.BentoUploadStatusUploading

	defer func() {
		org, err := schema.GetOrganization(ctx)
		if err != nil {
			abortWithError(ctx, err)
			return
		}

		user, err := services.GetCurrentUser(ctx)
		if err != nil {
			abortWithError(ctx, err)
			return
		}

		apiTokenName := ""
		if user.ApiToken != nil {
			apiTokenName = user.ApiToken.Name
		}
		createEventOpt := services.CreateEventOption{
			CreatorId:      user.ID,
			ApiTokenName:   apiTokenName,
			OrganizationId: &org.ID,
			ResourceType:   modelschemas.ResourceTypeBento,
			ResourceId:     bento.ID,
			Status:         modelschemas.EventStatusSuccess,
			OperationName:  "pushed",
		}
		if uploadStatus != modelschemas.BentoUploadStatusSuccess {
			createEventOpt.Status = modelschemas.EventStatusFailed
		}

		if _, err = services.EventService.Create(ctx, createEventOpt); err != nil {
			abortWithError(ctx, err)
			return
		}
	}()

	now := time.Now()
	nowPtr := &now
	bento, err = services.BentoService.Update(ctx, bento, services.UpdateBentoOption{
		UploadStatus:    &uploadStatus,
		UploadStartedAt: &nowPtr,
	})
	if err != nil {
		abortWithError(ctx, err)
		return
	}

	bodySize := ctx.Request.ContentLength

	err = services.BentoService.Upload(ctx, bento, ctx.Request.Body, bodySize)
	if err != nil {
		uploadStatus = modelschemas.BentoUploadStatusFailed
		now = time.Now()
		nowPtr = &now
		bento, err = services.BentoService.Update(ctx, bento, services.UpdateBentoOption{
			UploadStatus:    &uploadStatus,
			UploadStartedAt: &nowPtr,
		})
		if err != nil {
			abortWithError(ctx, err)
			return
		}
	}

	uploadStatus = modelschemas.BentoUploadStatusSuccess
	now = time.Now()
	nowPtr = &now
	_, err = services.BentoService.Update(ctx, bento, services.UpdateBentoOption{
		UploadStatus:    &uploadStatus,
		UploadStartedAt: &nowPtr,
	})
	if err != nil {
		abortWithError(ctx, err)
		return
	}
}

const BentomlVersionHeader = "X-Bentoml-Version"

func getBentomlVersion(ctx *gin.Context) string {
	return ctx.GetHeader(BentomlVersionHeader)
}

func clientSupportProxyTransmission(ctx *gin.Context) bool {
	return getBentomlVersion(ctx) != ""
}

func clientSupportTransmissionStrategy(ctx *gin.Context) (bool, error) {
	ver := getBentomlVersion(ctx)
	if ver == "" {
		return false, nil
	}
	currentVersion, err := pep440version.Parse(ver)
	if err != nil {
		err = errors.Wrapf(err, "parse bentoml version %s from request header", ver)
		return false, err
	}
	minVersion := pep440version.MustParse("1.0.5")
	return currentVersion.GreaterThan(minVersion), nil
}

func (c *bentoController) StartMultipartUpload(ctx *gin.Context, schema *GetBentoSchema) (*schemasv1.BentoSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, bento); err != nil {
		return nil, err
	}
	bentoSchema, err := transformersv1.ToBentoSchema(ctx, bento)
	if err != nil {
		return nil, err
	}
	uploadId, err := services.BentoService.StartMultipartUpload(ctx, bento)
	if err != nil {
		err = errors.Wrap(err, "failed to start multipart upload")
		return nil, err
	}
	bentoSchema.UploadId = uploadId
	return bentoSchema, nil
}

type PreSignBentoMultipartUploadUrl struct {
	GetBentoSchema
	schemasv1.PreSignMultipartUploadSchema
}

func (c *bentoController) PreSignMultipartUploadUrl(ctx *gin.Context, schema *PreSignBentoMultipartUploadUrl) (*schemasv1.BentoSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, bento); err != nil {
		return nil, err
	}
	bentoSchema, err := transformersv1.ToBentoSchema(ctx, bento)
	if err != nil {
		return nil, err
	}
	url_, err := services.BentoService.PreSignMultipartUploadUrl(ctx, bento, schema.UploadId, schema.PartNumber)
	if err != nil {
		err = errors.Wrap(err, "failed to pre sign multipart upload url")
		return nil, err
	}
	bentoSchema.PresignedUploadUrl = url_.String()
	return bentoSchema, nil
}

type CompleteBentoMultipartUpload struct {
	GetBentoSchema
	schemasv1.CompleteMultipartUploadSchema
}

func (c *bentoController) CompleteMultipartUpload(ctx *gin.Context, schema *CompleteBentoMultipartUpload) (*schemasv1.BentoSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, bento); err != nil {
		return nil, err
	}
	bentoSchema, err := transformersv1.ToBentoSchema(ctx, bento)
	if err != nil {
		return nil, err
	}
	parts := make([]minio.CompletePart, 0, len(schema.Parts))
	for _, part := range schema.Parts {
		parts = append(parts, minio.CompletePart{
			ETag:       part.ETag,
			PartNumber: part.PartNumber,
		})
	}
	err = services.BentoService.CompleteMultipartUpload(ctx, bento, schema.UploadId, parts)
	if err != nil {
		err = errors.Wrap(err, "failed to complete multipart upload")
		return nil, err
	}
	return bentoSchema, nil
}

func (c *bentoController) PreSignUploadUrl(ctx *gin.Context, schema *GetBentoSchema) (*schemasv1.BentoSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, bento); err != nil {
		return nil, err
	}
	bentoSchema, err := transformersv1.ToBentoSchema(ctx, bento)
	if err != nil {
		return nil, err
	}
	supportTransmissionStrategy, err := clientSupportTransmissionStrategy(ctx)
	if err != nil {
		return nil, err
	}
	if supportTransmissionStrategy {
		url_, err := services.BentoService.PreSignUploadUrl(ctx, bento)
		if err != nil {
			return nil, errors.Wrap(err, "pre sign upload url")
		}
		bentoSchema.PresignedUploadUrl = url_.String()
		return bentoSchema, nil
	}
	supportProxy := clientSupportProxyTransmission(ctx)
	if !supportProxy || bentoSchema.TransmissionStrategy == modelschemas.TransmissionStrategyPresignedURL {
		url_, err := services.BentoService.PreSignUploadUrl(ctx, bento)
		if err != nil {
			return nil, errors.Wrap(err, "pre sign upload url")
		}
		bentoSchema.PresignedUploadUrl = url_.String()
	} else {
		bentoSchema.PresignedUrlsDeprecated = true
	}
	return bentoSchema, nil
}

func (c *bentoController) Download(ctx *gin.Context) {
	schema := GetBentoSchema{
		GetBentoRepositorySchema: GetBentoRepositorySchema{
			BentoRepositoryName: ctx.Param("bentoRepositoryName"),
		},
		Version: ctx.Param("version"),
	}

	bento, err := schema.GetBento(ctx)
	if err != nil {
		abortWithError(ctx, err)
		return
	}
	if err = c.canUpdate(ctx, bento); err != nil {
		abortWithError(ctx, err)
		return
	}
	if err = services.BentoService.Download(ctx, bento, ctx.Writer); err != nil {
		abortWithError(ctx, err)
		return
	}
}

func (c *bentoController) PreSignDownloadUrl(ctx *gin.Context, schema *GetBentoSchema) (*schemasv1.BentoSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, bento); err != nil {
		return nil, err
	}
	bentoSchema, err := transformersv1.ToBentoSchema(ctx, bento)
	if err != nil {
		return nil, err
	}
	supportTransmissionStrategy, err := clientSupportTransmissionStrategy(ctx)
	if err != nil {
		return nil, err
	}
	if supportTransmissionStrategy {
		url_, err := services.BentoService.PreSignDownloadUrl(ctx, bento)
		if err != nil {
			return nil, errors.Wrap(err, "pre sign download url")
		}
		bentoSchema.PresignedUploadUrl = url_.String()
		return bentoSchema, nil
	}
	supportProxy := clientSupportProxyTransmission(ctx)
	if !supportProxy || bentoSchema.TransmissionStrategy == modelschemas.TransmissionStrategyPresignedURL {
		url_, err := services.BentoService.PreSignDownloadUrl(ctx, bento)
		if err != nil {
			return nil, errors.Wrap(err, "pre sign download url")
		}
		bentoSchema.PresignedDownloadUrl = url_.String()
	} else {
		bentoSchema.PresignedUrlsDeprecated = true
	}
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
	if schema.Status != nil {
		user, err := services.GetCurrentUser(ctx)
		if err != nil {
			return nil, err
		}
		bentoRepository, err := services.BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
		if err != nil {
			return nil, err
		}
		org, err := services.OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
		if err != nil {
			return nil, err
		}
		apiTokenName := ""
		if user.ApiToken != nil {
			apiTokenName = user.ApiToken.Name
		}
		createEventOpt := services.CreateEventOption{
			CreatorId:      user.ID,
			ApiTokenName:   apiTokenName,
			OrganizationId: &org.ID,
			ResourceType:   modelschemas.ResourceTypeBento,
			ResourceId:     bento.ID,
			Status:         modelschemas.EventStatusSuccess,
			OperationName:  "pushed",
		}
		if *schema.Status != modelschemas.BentoUploadStatusSuccess {
			createEventOpt.Status = modelschemas.EventStatusFailed
		}
		if _, err = services.EventService.Create(ctx, createEventOpt); err != nil {
			return nil, errors.Wrap(err, "create event")
		}
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
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	majorCluster, err := services.OrganizationService.GetMajorCluster(ctx, org)
	if err != nil {
		return nil, err
	}
	return transformersv1.ToKubePodSchemas(ctx, majorCluster.ID, pods)
}

func (c *bentoController) Get(ctx *gin.Context, schema *GetBentoSchema) (*schemasv1.BentoFullSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, bento); err != nil {
		return nil, err
	}
	return transformersv1.ToBentoFullSchema(ctx, bento)
}

type ListBentoDeploymentSchema struct {
	schemasv1.ListQuerySchema
	GetBentoSchema
}

func (c *bentoController) ListDeployment(ctx *gin.Context, schema *ListBentoDeploymentSchema) (*schemasv1.DeploymentListSchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, bento); err != nil {
		return nil, err
	}
	deployments, total, err := services.DeploymentService.List(ctx, services.ListDeploymentOption{
		BaseListOption: services.BaseListOption{
			Start:  &schema.Start,
			Count:  &schema.Count,
			Search: schema.Search,
		},
		BentoIds: &[]uint{bento.ID},
	})
	if err != nil {
		return nil, err
	}
	deploymentSchemas, err := transformersv1.ToDeploymentSchemas(ctx, deployments)
	if err != nil {
		return nil, err
	}
	return &schemasv1.DeploymentListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: deploymentSchemas,
	}, nil
}

type ListBentoModelSchema struct {
	schemasv1.ListQuerySchema
	GetBentoSchema
}

func (c *bentoController) ListModel(ctx *gin.Context, schema *ListBentoModelSchema) ([]*schemasv1.ModelWithRepositorySchema, error) {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, bento); err != nil {
		return nil, err
	}
	models, err := services.BentoService.ListModelsFromManifests(ctx, bento)
	if err != nil {
		return nil, err
	}
	return transformersv1.ToModelWithRepositorySchemas(ctx, models)
}

type ListBentoSchema struct {
	schemasv1.ListQuerySchema
	GetBentoRepositorySchema
}

func (c *bentoController) buildListOpt(ctx context.Context, listOpt *services.ListBentoOption, q schemasv1.Q) error {
	queryMap := q.ToMap()
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
				return err
			}
			users, err := services.UserService.ListByNames(ctx, userNames)
			if err != nil {
				return err
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
				"size":       {},
			}[fieldName]; !ok {
				continue
			}
			if fieldName == "size" {
				fieldName = "manifest->'size_bytes'"
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
	return nil
}

func (c *bentoController) List(ctx *gin.Context, schema *ListBentoSchema) (*schemasv1.BentoWithRepositoryListSchema, error) {
	bentoRepository, err := schema.GetBentoRepository(ctx)
	if err != nil {
		return nil, err
	}

	if err = BentoRepositoryController.canView(ctx, bentoRepository); err != nil {
		return nil, err
	}

	listOpt := services.ListBentoOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		BentoRepositoryId: utils.UintPtr(bentoRepository.ID),
	}

	err = c.buildListOpt(ctx, &listOpt, schema.Q)
	if err != nil {
		return nil, err
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

	err = c.buildListOpt(ctx, &listOpt, schema.Q)
	if err != nil {
		return nil, err
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

type ListAllImageBuildStatusUnsyncedBentoSchema struct {
	GetOrganizationSchema
}

func (c *bentoController) ListImageBuildStatusUnsynced(ctx *gin.Context, schema *ListAllImageBuildStatusUnsyncedBentoSchema) ([]*schemasv1.BentoWithRepositorySchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = OrganizationController.canView(ctx, organization); err != nil {
		return nil, err
	}

	bentos, err := services.BentoService.ListImageBuildStatusUnsynced(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list bentos")
	}

	return transformersv1.ToBentoWithRepositorySchemas(ctx, bentos)
}

type UpdateBentoImageBuildStatusSyncingAtSchema struct {
	GetBentoSchema
}

func (c *bentoController) UpdateBentoImageBuildStatusSyncingAt(ctx *gin.Context, schema *UpdateBentoImageBuildStatusSyncingAtSchema) error {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return err
	}

	if err = BentoController.canUpdate(ctx, bento); err != nil {
		return err
	}

	now := time.Now()
	nowPtr := &now
	_, err = services.BentoService.Update(ctx, bento, services.UpdateBentoOption{
		ImageBuildStatusSyncingAt: &nowPtr,
	})
	if err != nil {
		return errors.Wrap(err, "update bento")
	}

	return nil
}

type UpdateBentoImageBuildStatusSchema struct {
	GetBentoSchema
	ImageBuildStatus modelschemas.ImageBuildStatus `json:"image_build_status"`
}

func (c *bentoController) UpdateBentoImageBuildStatus(ctx *gin.Context, schema *UpdateBentoImageBuildStatusSchema) error {
	bento, err := schema.GetBento(ctx)
	if err != nil {
		return err
	}

	if err = BentoController.canUpdate(ctx, bento); err != nil {
		return err
	}

	now := time.Now()
	nowPtr := &now
	_, err = services.BentoService.Update(ctx, bento, services.UpdateBentoOption{
		ImageBuildStatus:          &schema.ImageBuildStatus,
		ImageBuildStatusUpdatedAt: &nowPtr,
	})
	if err != nil {
		return errors.Wrap(err, "update bento")
	}

	return nil
}

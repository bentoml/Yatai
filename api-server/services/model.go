package services

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"gorm.io/gorm"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/schemas/modelschemas"
)

type modelService struct{}

var ModelService = &modelService{}

func (s *modelService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Model{})
}

type CreateModelOption struct {
	CreatorId         uint
	ModelRepositoryId uint
	Version           string
	Description       string
	BuildAt           time.Time
	Manifest          *modelschemas.ModelManifestSchema
	Labels            modelschemas.LabelItemsSchema
}

type UpdateModelOption struct {
	ImageBuildStatus          *modelschemas.ImageBuildStatus
	ImageBuildStatusSyncingAt **time.Time
	ImageBuildStatusUpdatedAt **time.Time
	UploadStatus              *modelschemas.ModelUploadStatus
	UploadStartedAt           **time.Time
	UploadFinishedAt          **time.Time
	UploadFinishedReason      *string
	Labels                    *modelschemas.LabelItemsSchema
}

type ListModelOption struct {
	BaseListOption
	BaseListByLabelsOption
	ModelRepositoryId *uint
	Ids               *[]uint
	Versions          *[]string
	BentoIds          *[]uint
	OrganizationId    *uint
	CreatorId         *uint
	CreatorIds        *[]uint
	Order             *string
	Names             *[]string
	Modules           *[]string
}

func (s *modelService) Create(ctx context.Context, opt CreateModelOption) (model *models.Model, err error) {
	// nolint: ineffassign, staticcheck
	db, ctx, df, err := startTransaction(ctx)
	if err != nil {
		return
	}
	defer func() { df(err) }()
	model = &models.Model{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		ModelRepositoryAssociate: models.ModelRepositoryAssociate{
			ModelRepositoryId: opt.ModelRepositoryId,
		},
		Version:          opt.Version,
		Description:      opt.Description,
		ImageBuildStatus: modelschemas.ImageBuildStatusPending,
		UploadStatus:     modelschemas.ModelUploadStatusPending,
		BuildAt:          opt.BuildAt,
		Manifest:         opt.Manifest,
	}

	err = db.Create(model).Error
	if err != nil {
		return
	}

	var modelRepository *models.ModelRepository
	modelRepository, err = ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
	if err != nil {
		return
	}
	var org *models.Organization
	org, err = OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
	if err != nil {
		return
	}
	var user *models.User
	user, err = GetCurrentUser(ctx)
	if err != nil {
		return
	}
	err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, opt.Labels, user.ID, org.ID, model)
	return
}

func (s *modelService) PreSignUploadUrl(ctx context.Context, model *models.Model) (url *url.URL, err error) {
	modelRepository, err := ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
	if err != nil {
		return
	}
	s3Config, err := OrganizationService.GetS3Config(ctx, org)
	if err != nil {
		return
	}
	minioClient, err := s3Config.GetMinioClient()
	if err != nil {
		err = errors.Wrap(err, "create s3 client")
		return
	}

	bucketName, err := s.GetS3BucketName(ctx, model)
	if err != nil {
		return
	}

	err = s3Config.MakeSureBucket(ctx, bucketName)
	if err != nil {
		return
	}

	objectName, err := s.getS3ObjectName(ctx, model)
	if err != nil {
		return
	}

	url, err = minioClient.PresignedPutObject(ctx, bucketName, objectName, time.Hour)
	if err != nil {
		err = errors.Wrap(err, "presigned put object")
		return
	}
	if s3Config.Endpoint != s3Config.EndpointInCluster {
		url.Host = s3Config.Endpoint
	}
	return
}

func (s *modelService) PreSignDownloadUrl(ctx context.Context, model *models.Model) (url *url.URL, err error) {
	modelRepository, err := ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
	if err != nil {
		return
	}
	s3Config, err := OrganizationService.GetS3Config(ctx, org)
	if err != nil {
		return
	}
	minioClient, err := s3Config.GetMinioClient()
	if err != nil {
		err = errors.Wrap(err, "create s3 client")
		return
	}

	bucketName, err := s.GetS3BucketName(ctx, model)
	if err != nil {
		return
	}

	err = s3Config.MakeSureBucket(ctx, bucketName)
	if err != nil {
		return
	}

	objectName, err := s.getS3ObjectName(ctx, model)
	if err != nil {
		return
	}

	url, err = minioClient.PresignedGetObject(ctx, bucketName, objectName, time.Hour, nil)
	if err != nil {
		err = errors.Wrap(err, "presigned get object")
		return
	}
	if s3Config.Endpoint != s3Config.EndpointInCluster {
		url.Host = s3Config.Endpoint
	}
	return
}

func (s *modelService) getS3ObjectName(ctx context.Context, model *models.Model) (string, error) {
	modelRepository, err := ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
	if err != nil {
		return "", err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
	if err != nil {
		return "", err
	}
	objectName := fmt.Sprintf("models/%s/%s/%s.tar.gz", org.Name, modelRepository.Name, model.Version)
	return objectName, nil
}

func (s *modelService) GetImageName(ctx context.Context, model *models.Model, inCluster bool) (string, error) {
	modelRepository, err := ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
	if err != nil {
		return "", err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
	if err != nil {
		return "", err
	}
	dockerRegistry, err := OrganizationService.GetDockerRegistry(ctx, org)
	if err != nil {
		return "", err
	}
	var imageName string
	if inCluster {
		imageName = fmt.Sprintf("%s:yatai.%s.%s.%s", dockerRegistry.ModelsRepositoryURIInCluster, org.Name, modelRepository.Name, model.Version)
	} else {
		imageName = fmt.Sprintf("%s:yatai.%s.%s.%s", dockerRegistry.ModelsRepositoryURI, org.Name, modelRepository.Name, model.Version)
	}
	return imageName, nil
}

func (s *modelService) GetS3BucketName(ctx context.Context, model *models.Model) (string, error) {
	modelRepository, err := ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
	if err != nil {
		return "", err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
	if err != nil {
		return "", err
	}

	s3Config, err := OrganizationService.GetS3Config(ctx, org)
	if err != nil {
		return "", err
	}

	s3BucketName := s3Config.ModelsBucketName

	return s3BucketName, nil
}

func (s *modelService) GetTag(ctx context.Context, model *models.Model) (modelschemas.Tag, error) {
	modelRepository, err := ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
	if err != nil {
		return "", err
	}
	return modelschemas.Tag(fmt.Sprintf("%s:%s", modelRepository.Name, model.Version)), nil
}

func (s *modelService) Update(ctx context.Context, model *models.Model, opt UpdateModelOption) (*models.Model, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.ImageBuildStatus != nil {
		updaters["image_build_status"] = *opt.ImageBuildStatus
		defer func() {
			if err != nil {
				model.ImageBuildStatus = *opt.ImageBuildStatus
			}
		}()
	}
	if opt.ImageBuildStatusSyncingAt != nil {
		updaters["image_build_status_syncing_at"] = *opt.ImageBuildStatusSyncingAt
		defer func() {
			if err != nil {
				model.ImageBuildStatusSyncingAt = *opt.ImageBuildStatusSyncingAt
			}
		}()
	}
	if opt.ImageBuildStatusUpdatedAt != nil {
		updaters["image_build_status_updated_at"] = *opt.ImageBuildStatusUpdatedAt
		defer func() {
			if err != nil {
				model.ImageBuildStatusUpdatedAt = *opt.ImageBuildStatusUpdatedAt
			}
		}()
	}
	if opt.UploadStatus != nil {
		updaters["upload_status"] = *opt.UploadStatus
		defer func() {
			if err != nil {
				model.UploadStatus = *opt.UploadStatus
			}
		}()
	}
	if opt.UploadStartedAt != nil {
		updaters["upload_started_at"] = *opt.UploadStartedAt
		defer func() {
			if err != nil {
				model.UploadStartedAt = *opt.UploadStartedAt
			}
		}()
	}
	if opt.UploadFinishedAt != nil {
		updaters["upload_finished_at"] = *opt.UploadFinishedAt
		defer func() {
			if err != nil {
				model.UploadFinishedAt = *opt.UploadFinishedAt
			}
		}()
	}
	if opt.UploadFinishedReason != nil {
		updaters["upload_finished_reason"] = *opt.UploadFinishedReason
		defer func() {
			if err != nil {
				model.UploadFinishedReason = *opt.UploadFinishedReason
			}
		}()
	}

	// nolint: ineffassign,staticcheck
	db, ctx, df, err := startTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { df(err) }()

	if len(updaters) > 0 {
		err = db.Model(&models.Model{}).Where("id = ?", model.ID).Updates(updaters).Error
		if err != nil {
			return nil, err
		}
	}

	if opt.Labels != nil {
		modelRepository, err := ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
		if err != nil {
			return nil, err
		}
		org, err := OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
		if err != nil {
			return nil, err
		}
		user, err := GetCurrentUser(ctx)
		if err != nil {
			return nil, err
		}
		err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, *opt.Labels, user.ID, org.ID, model)
		if err != nil {
			return nil, err
		}
	}

	if opt.UploadStatus == nil || *opt.UploadStatus != modelschemas.ModelUploadStatusSuccess {
		return model, nil
	}

	return s.CreateImageBuilderJob(ctx, model)
}

func (s *modelService) CreateImageBuilderJob(ctx context.Context, model *models.Model) (*models.Model, error) {
	modelRepository, err := ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
	if err != nil {
		return nil, err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
	if err != nil {
		return nil, err
	}

	majorCluster, err := OrganizationService.GetMajorCluster(ctx, org)
	if err != nil {
		return nil, err
	}

	dockerFileCMKubeName := fmt.Sprintf("docker-file-%d", model.ID)
	dockerFileContent := `
FROM scratch

COPY . /model
`

	kubeName, err := s.GetImageBuilderKubeName(ctx, model)
	if err != nil {
		return nil, err
	}

	s3ObjectName, err := s.getS3ObjectName(ctx, model)
	if err != nil {
		return nil, err
	}

	imageName, err := s.GetImageName(ctx, model, true)
	if err != nil {
		return nil, err
	}

	s3BucketName, err := s.GetS3BucketName(ctx, model)
	if err != nil {
		return nil, err
	}

	kubeLabels, err := s.GetImageBuilderKubeLabels(ctx, model)
	if err != nil {
		return nil, err
	}

	err = ImageBuilderService.CreateImageBuilderJob(ctx, CreateImageBuilderJobOption{
		KubeName:             kubeName,
		ImageName:            imageName,
		S3ObjectName:         s3ObjectName,
		S3BucketName:         s3BucketName,
		Cluster:              majorCluster,
		DockerFileCMKubeName: &dockerFileCMKubeName,
		DockerFileContent:    &dockerFileContent,
		KubeLabels:           kubeLabels,
	})
	if err != nil {
		return nil, err
	}

	go func() {
		_, _ = s.SyncImageBuilderStatus(ctx, model)
	}()

	model.ImageBuildStatus = modelschemas.ImageBuildStatusBuilding

	return model, nil
}

func (s *modelService) GetImageBuilderKubeName(ctx context.Context, model *models.Model) (string, error) {
	modelRepository, err := ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
	if err != nil {
		return "", err
	}

	org, err := OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
	if err != nil {
		return "", err
	}

	guid := xid.New()
	return strings.ReplaceAll(strcase.ToKebab(fmt.Sprintf("yatai-model-image-builder-%s-%s-%s-%s", org.Name, modelRepository.Name, model.Version, guid.String())), ".", "-"), nil
}

func (s *modelService) Get(ctx context.Context, id uint) (*models.Model, error) {
	var model models.Model
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&model).Error
	if err != nil {
		return nil, err
	}
	if model.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &model, nil
}

func (s *modelService) GetByUid(ctx context.Context, uid string) (*models.Model, error) {
	var model models.Model
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&model).Error
	if err != nil {
		return nil, err
	}
	if model.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &model, nil
}

func (s *modelService) GetByVersion(ctx context.Context, modelRepositoryId uint, version string) (*models.Model, error) {
	var model models.Model
	err := getBaseQuery(ctx, s).Where("model_repository_id = ?", modelRepositoryId).Where("version = ?", version).First(&model).Error
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get model by model repository id %d and model %s", modelRepositoryId, version)
	}
	if model.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &model, nil
}

func (s *modelService) ListByUids(ctx context.Context, uids []string) ([]*models.Model, error) {
	models_ := make([]*models.Model, 0, len(uids))
	if len(uids) == 0 {
		return models_, nil
	}
	err := getBaseQuery(ctx, s).Where("uid in (?)", uids).Find(&models_).Error
	return models_, err
}

func (s *modelService) ListAllModules(ctx context.Context, organizationId uint) ([]string, error) {
	db := s.getBaseDB(ctx)
	query := db.Raw(`select distinct(model.manifest->>'module') from model join model_repository on model.model_repository_id = model_repository.id where model_repository.organization_id = ?`, organizationId)
	res := make([]string, 0)
	err := query.Find(&res).Error
	return res, err
}

func (s *modelService) List(ctx context.Context, opt ListModelOption) ([]*models.Model, uint, error) {
	query := getBaseQuery(ctx, s)
	query = query.Joins("LEFT JOIN model_repository ON model.model_repository_id = model_repository.id")
	if opt.BentoIds != nil {
		query = query.Joins("LEFT JOIN bento_model_rel ON bento_model_rel.model_id = model.id").Where("bento_model_rel.bento_id in (?)", *opt.BentoIds)
	}
	if opt.OrganizationId != nil {
		query = query.Where("model_repository.organization_id = ?", *opt.OrganizationId)
	}
	if opt.Ids != nil {
		query = query.Where("model.id in (?)", *opt.Ids)
	}
	if opt.Versions != nil {
		query = query.Where("model.version in (?)", *opt.Versions)
	}
	if opt.ModelRepositoryId != nil {
		query = query.Where("model.model_repository_id = ?", *opt.ModelRepositoryId)
	}
	if opt.CreatorId != nil {
		query = query.Where("model.creator_id = ?", *opt.CreatorId)
	}
	if opt.Names != nil {
		query = query.Where("model_repository.name in (?)", *opt.Names)
	}
	if opt.CreatorIds != nil {
		query = query.Where("model.creator_id in (?)", *opt.CreatorIds)
	}
	if opt.Modules != nil {
		query = query.Where("model.manifest->>'module' in (?)", *opt.Modules)
	}
	query = opt.BindQueryWithKeywords(query, "model_repository")
	query = opt.BindQueryWithLabels(query, modelschemas.ResourceTypeModel)
	query = query.Select("distinct(model.*)")
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	query = s.getBaseDB(ctx).Table("(?) as model", query)
	models_ := make([]*models.Model, 0)
	query = opt.BindQueryWithLimit(query)
	if opt.Order != nil {
		query = query.Order(*opt.Order)
	} else {
		query = query.Order("model.build_at DESC")
	}
	err = query.Find(&models_).Error
	if err != nil {
		return nil, 0, err
	}
	return models_, uint(total), nil
}

func (s *modelService) ListLatestByModelRepositoryIds(ctx context.Context, modelRepositoryIds []uint) ([]*models.Model, error) {
	db := mustGetSession(ctx)

	query := db.Raw(`select * from model where id in (
			select n.model_id from (
				select model_repository_id, max(id) as model_id from model
				where model_repository_id in (?) group by model_repository_id
			) as n)`, modelRepositoryIds)
	models_ := make([]*models.Model, 0, len(modelRepositoryIds))
	err := query.Find(&models_).Error
	if err != nil {
		return nil, err
	}
	return models_, err
}

func (s *modelService) ListImageBuilderPods(ctx context.Context, model *models.Model) ([]*models.KubePodWithStatus, error) {
	modelRepository, err := ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
	if err != nil {
		return nil, err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
	if err != nil {
		return nil, err
	}
	cluster, err := OrganizationService.GetMajorCluster(ctx, org)
	if err != nil {
		return nil, err
	}

	kubeLabels, err := s.GetImageBuilderKubeLabels(ctx, model)
	if err != nil {
		return nil, err
	}

	return ImageBuilderService.ListImageBuilderPods(ctx, cluster, kubeLabels)
}

func (s *modelService) GetImageBuilderKubeLabels(ctx context.Context, model *models.Model) (map[string]string, error) {
	modelRepository, err := ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		consts.KubeLabelYataiModelRepository: modelRepository.Name,
		consts.KubeLabelYataiModel:           model.Version,
	}, nil
}

func (s *modelService) CalculateImageBuildStatus(ctx context.Context, model *models.Model) (modelschemas.ImageBuildStatus, error) {
	defaultStatus := modelschemas.ImageBuildStatusPending
	pods, err := s.ListImageBuilderPods(ctx, model)
	if err != nil {
		return defaultStatus, err
	}
	return ImageBuilderService.CalculateImageBuildStatus(pods)
}

func (s *modelService) ListImageBuildStatusUnsynced(ctx context.Context) ([]*models.Model, error) {
	q := getBaseQuery(ctx, s)
	now := time.Now()
	t := now.Add(-time.Minute)
	q = q.Where("image_build_status != ? and (image_build_status_syncing_at is null or image_build_status_syncing_at < ? or image_build_status_updated_at is null or image_build_status_updated_at < ?)", modelschemas.ImageBuildStatusSuccess, t, t)
	models_ := make([]*models.Model, 0)
	err := q.Order("id DESC").Find(&models_).Error
	return models_, err
}

func (s *modelService) SyncImageBuilderStatus(ctx context.Context, model *models.Model) (modelschemas.ImageBuildStatus, error) {
	now := time.Now()
	nowPtr := &now
	_, err := s.Update(ctx, model, UpdateModelOption{
		ImageBuildStatusSyncingAt: &nowPtr,
	})
	if err != nil {
		return model.ImageBuildStatus, err
	}
	currentStatus, err := s.CalculateImageBuildStatus(ctx, model)
	if err != nil {
		return model.ImageBuildStatus, err
	}
	now = time.Now()
	nowPtr = &now
	_, err = s.Update(ctx, model, UpdateModelOption{
		ImageBuildStatus:          &currentStatus,
		ImageBuildStatusUpdatedAt: &nowPtr,
	})
	if err != nil {
		return currentStatus, err
	}
	if currentStatus == modelschemas.ImageBuildStatusSuccess {
		bentos, _, err := BentoService.List(ctx, ListBentoOption{
			ModelIds: &[]uint{model.ID},
		})
		if err != nil {
			return currentStatus, err
		}
		for _, bento := range bentos {
			bento := bento
			go func() {
				_, _ = BentoService.SyncImageBuilderStatus(ctx, bento)
			}()
		}
	}
	return currentStatus, nil
}

type IModelAssociated interface {
	GetAssociatedModelId() uint
	GetAssociatedModelCache() *models.Model
	SetAssociatedModelCache(model *models.Model)
}

func (s *modelService) GetAssociatedModel(ctx context.Context, associate IModelAssociated) (*models.Model, error) {
	cache := associate.GetAssociatedModelCache()
	if cache != nil {
		return cache, nil
	}
	model, err := s.Get(ctx, associate.GetAssociatedModelId())
	associate.SetAssociatedModelCache(model)
	return model, err
}

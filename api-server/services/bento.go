package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/huandu/xstrings"
	"github.com/iancoleman/strcase"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	commonconsts "github.com/bentoml/yatai-common/consts"
	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/utils"
)

type bentoService struct{}

var BentoService = bentoService{}

func (s *bentoService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Bento{})
}

type CreateBentoOption struct {
	CreatorId         uint
	BentoRepositoryId uint
	Version           string
	Description       string
	BuildAt           time.Time
	Manifest          *modelschemas.BentoManifestSchema
	Labels            modelschemas.LabelItemsSchema
}

type UpdateBentoOption struct {
	ImageBuildStatus          *modelschemas.ImageBuildStatus
	ImageBuildStatusSyncingAt **time.Time
	ImageBuildStatusUpdatedAt **time.Time
	UploadStatus              *modelschemas.BentoUploadStatus
	UploadStartedAt           **time.Time
	UploadFinishedAt          **time.Time
	UploadFinishedReason      *string
	Labels                    *modelschemas.LabelItemsSchema
	Manifest                  **modelschemas.BentoManifestSchema
}

type ListBentoOption struct {
	BaseListOption
	BaseListByLabelsOption
	OrganizationId    *uint
	BentoRepositoryId *uint
	Versions          *[]string
	ModelIds          *[]uint
	CreatorId         *uint
	CreatorIds        *[]uint
	Order             *string
	Names             *[]string
	Ids               *[]uint
}

func (s *bentoService) Create(ctx context.Context, opt CreateBentoOption) (bento *models.Bento, err error) {
	// nolint: ineffassign,staticcheck
	db, ctx, df, err := startTransaction(ctx)
	if err != nil {
		return
	}
	defer func() { df(err) }()
	bento = &models.Bento{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		BentoRepositoryAssociate: models.BentoRepositoryAssociate{
			BentoRepositoryId: opt.BentoRepositoryId,
		},
		Version:          opt.Version,
		Description:      opt.Description,
		ImageBuildStatus: modelschemas.ImageBuildStatusPending,
		UploadStatus:     modelschemas.BentoUploadStatusPending,
		BuildAt:          opt.BuildAt,
		Manifest:         opt.Manifest,
	}
	err = db.Create(bento).Error
	if err != nil {
		return
	}
	models_, err := s.ListModelsFromManifests(ctx, bento)
	if err != nil {
		return
	}
	for _, model := range models_ {
		rel := &models.BentoModelRel{
			BentoAssociate: models.BentoAssociate{
				BentoId: bento.ID,
			},
			ModelAssociate: models.ModelAssociate{
				ModelId: model.ID,
			},
		}
		err = db.Create(rel).Error
		if err != nil {
			return
		}
	}

	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return
	}
	user, err := GetCurrentUser(ctx)
	if err != nil {
		return
	}

	err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, opt.Labels, user.ID, org.ID, bento)

	return
}

func (s *bentoService) PreSignUploadUrl(ctx context.Context, bento *models.Bento) (url *url.URL, err error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
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

	bucketName, err := s.GetS3BucketName(ctx, bento)
	if err != nil {
		return
	}

	err = s3Config.MakeSureBucket(ctx, bucketName)
	if err != nil {
		return
	}

	objectName, err := s.getS3ObjectName(ctx, bento)
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

func (s *bentoService) StartMultipartUpload(ctx context.Context, bento *models.Bento) (uploadId string, err error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return
	}
	s3Config, err := OrganizationService.GetS3Config(ctx, org)
	if err != nil {
		return
	}
	minioCore, err := s3Config.GetMinioCore()
	if err != nil {
		err = errors.Wrap(err, "create s3 client")
		return
	}

	bucketName, err := s.GetS3BucketName(ctx, bento)
	if err != nil {
		return
	}

	err = s3Config.MakeSureBucket(ctx, bucketName)
	if err != nil {
		return
	}

	objectName, err := s.getS3ObjectName(ctx, bento)
	if err != nil {
		return
	}

	uploadId, err = minioCore.NewMultipartUpload(ctx, bucketName, objectName, minio.PutObjectOptions{})
	if err != nil {
		err = errors.Wrap(err, "new multipart upload")
		return
	}
	return
}

func (s *bentoService) PreSignMultipartUploadUrl(ctx context.Context, bento *models.Bento, uploadId string, partNumber int) (url_ *url.URL, err error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return
	}
	s3Config, err := OrganizationService.GetS3Config(ctx, org)
	if err != nil {
		return
	}
	minioCore, err := s3Config.GetMinioCore()
	if err != nil {
		err = errors.Wrap(err, "create s3 client")
		return
	}

	bucketName, err := s.GetS3BucketName(ctx, bento)
	if err != nil {
		return
	}

	err = s3Config.MakeSureBucket(ctx, bucketName)
	if err != nil {
		return
	}

	objectName, err := s.getS3ObjectName(ctx, bento)
	if err != nil {
		return
	}

	queryValues := make(url.Values)
	queryValues.Set("partNumber", strconv.Itoa(partNumber))
	queryValues.Set("uploadId", uploadId)

	url_, err = minioCore.Presign(ctx, http.MethodPut, bucketName, objectName, time.Hour, queryValues)
	if err != nil {
		err = errors.Wrap(err, "presigned put object")
		return
	}
	if s3Config.Endpoint != s3Config.EndpointInCluster {
		url_.Host = s3Config.Endpoint
	}
	return
}

func (s *bentoService) CompleteMultipartUpload(ctx context.Context, bento *models.Bento, uploadId string, parts []minio.CompletePart) (err error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return
	}
	s3Config, err := OrganizationService.GetS3Config(ctx, org)
	if err != nil {
		return
	}
	minioCore, err := s3Config.GetMinioCore()
	if err != nil {
		err = errors.Wrap(err, "create s3 client")
		return
	}

	bucketName, err := s.GetS3BucketName(ctx, bento)
	if err != nil {
		return
	}

	err = s3Config.MakeSureBucket(ctx, bucketName)
	if err != nil {
		return
	}

	objectName, err := s.getS3ObjectName(ctx, bento)
	if err != nil {
		return
	}

	_, err = minioCore.CompleteMultipartUpload(ctx, bucketName, objectName, uploadId, parts, minio.PutObjectOptions{})
	if err != nil {
		err = errors.Wrap(err, "new multipart upload")
		return
	}
	return
}

func (s *bentoService) Upload(ctx context.Context, bento *models.Bento, reader io.Reader, objectSize int64) (err error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
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

	bucketName, err := s.GetS3BucketName(ctx, bento)
	if err != nil {
		return
	}

	err = s3Config.MakeSureBucket(ctx, bucketName)
	if err != nil {
		return
	}

	objectName, err := s.getS3ObjectName(ctx, bento)
	if err != nil {
		return
	}

	logrus.Debugf("uploading to s3: %s/%s", bucketName, objectName)
	_, err = minioClient.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		err = errors.Wrap(err, "put object")
		return
	}

	logrus.Debugf("uploaded to s3: %s/%s", bucketName, objectName)
	return
}

func (s *bentoService) PreSignDownloadUrl(ctx context.Context, bento *models.Bento) (url *url.URL, err error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
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

	bucketName, err := s.GetS3BucketName(ctx, bento)
	if err != nil {
		return
	}

	err = s3Config.MakeSureBucket(ctx, bucketName)
	if err != nil {
		return
	}

	objectName, err := s.getS3ObjectName(ctx, bento)
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

func (s *bentoService) Download(ctx context.Context, bento *models.Bento, writer io.Writer) (err error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
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

	bucketName, err := s.GetS3BucketName(ctx, bento)
	if err != nil {
		return
	}

	err = s3Config.MakeSureBucket(ctx, bucketName)
	if err != nil {
		return
	}

	objectName, err := s.getS3ObjectName(ctx, bento)
	if err != nil {
		return
	}

	obj, err := minioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		err = errors.Wrap(err, "get object")
		return
	}

	_, err = io.Copy(writer, obj)
	if err != nil {
		err = errors.Wrap(err, "copy object")
	}

	return
}

func (s *bentoService) getS3ObjectName(ctx context.Context, bento *models.Bento) (string, error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return "", err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return "", err
	}
	objectName := fmt.Sprintf("bentos/%s/%s/%s.tar.gz", org.Name, bentoRepository.Name, bento.Version)
	return objectName, nil
}

func (s *bentoService) GetS3BucketName(ctx context.Context, bento *models.Bento) (string, error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return "", err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return "", err
	}

	s3Config, err := OrganizationService.GetS3Config(ctx, org)
	if err != nil {
		return "", err
	}

	s3BucketName := s3Config.BentosBucketName

	return s3BucketName, nil
}

func (s *bentoService) GetTag(ctx context.Context, bento *models.Bento) (modelschemas.Tag, error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return "", err
	}
	return modelschemas.Tag(fmt.Sprintf("%s:%s", bentoRepository.Name, bento.Version)), nil
}

func (s *bentoService) Update(ctx context.Context, bento *models.Bento, opt UpdateBentoOption) (*models.Bento, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.ImageBuildStatus != nil {
		updaters["image_build_status"] = *opt.ImageBuildStatus
		defer func() {
			if err == nil {
				bento.ImageBuildStatus = *opt.ImageBuildStatus
			}
		}()
	}
	if opt.ImageBuildStatusSyncingAt != nil {
		updaters["image_build_status_syncing_at"] = *opt.ImageBuildStatusSyncingAt
		defer func() {
			if err == nil {
				bento.ImageBuildStatusSyncingAt = *opt.ImageBuildStatusSyncingAt
			}
		}()
	}
	if opt.ImageBuildStatusUpdatedAt != nil {
		updaters["image_build_status_updated_at"] = *opt.ImageBuildStatusUpdatedAt
		defer func() {
			if err == nil {
				bento.ImageBuildStatusUpdatedAt = *opt.ImageBuildStatusUpdatedAt
			}
		}()
	}
	if opt.UploadStatus != nil {
		updaters["upload_status"] = *opt.UploadStatus
		defer func() {
			if err == nil {
				bento.UploadStatus = *opt.UploadStatus
			}
		}()
	}
	if opt.UploadStartedAt != nil {
		updaters["upload_started_at"] = *opt.UploadStartedAt
		defer func() {
			if err == nil {
				bento.UploadStartedAt = *opt.UploadStartedAt
			}
		}()
	}
	if opt.UploadFinishedAt != nil {
		updaters["upload_finished_at"] = *opt.UploadFinishedAt
		defer func() {
			if err == nil {
				bento.UploadFinishedAt = *opt.UploadFinishedAt
			}
		}()
	}
	if opt.UploadFinishedReason != nil {
		updaters["upload_finished_reason"] = *opt.UploadFinishedReason
		defer func() {
			if err == nil {
				bento.UploadFinishedReason = *opt.UploadFinishedReason
			}
		}()
	}
	if opt.Manifest != nil {
		updaters["manifest"] = *opt.Manifest
		defer func() {
			if err == nil {
				bento.Manifest = *opt.Manifest
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
		err = db.Model(&models.Bento{}).Where("id = ?", bento.ID).Updates(updaters).Error
		if err != nil {
			return nil, err
		}
	}

	if opt.Labels != nil {
		bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
		if err != nil {
			return nil, err
		}
		org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
		if err != nil {
			return nil, err
		}
		user, err := GetCurrentUser(ctx)
		if err != nil {
			return nil, err
		}
		err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, *opt.Labels, user.ID, org.ID, bento)
		if err != nil {
			return nil, err
		}
	}

	if opt.UploadStatus == nil || *opt.UploadStatus != modelschemas.BentoUploadStatusSuccess {
		return bento, err
	}

	return bento, nil
}

func (s *bentoService) GetImageBuilderKubeName(ctx context.Context, bento *models.Bento) (string, error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return "", err
	}

	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return "", err
	}

	guid := xid.New()
	return strings.ReplaceAll(strcase.ToKebab(fmt.Sprintf("yatai-bento-image-builder-%s-%s-%s-%s", org.Name, bentoRepository.Name, bento.Version, guid.String())), ".", "-"), nil
}

func (s *bentoService) Get(ctx context.Context, id uint) (*models.Bento, error) {
	var bento models.Bento
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&bento).Error
	if err != nil {
		return nil, err
	}
	if bento.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bento, nil
}

func (s *bentoService) GetByUid(ctx context.Context, uid string) (*models.Bento, error) {
	var bento models.Bento
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&bento).Error
	if err != nil {
		return nil, err
	}
	if bento.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bento, nil
}

func (s *bentoService) GetByVersion(ctx context.Context, bentoRepositoryId uint, version string) (*models.Bento, error) {
	var bento models.Bento
	err := getBaseQuery(ctx, s).Where("bento_repository_id = ?", bentoRepositoryId).Where("version = ?", version).First(&bento).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get bento %s", version)
	}
	if bento.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bento, nil
}

func (s *bentoService) ListByUids(ctx context.Context, uids []string) ([]*models.Bento, error) {
	bentos := make([]*models.Bento, 0, len(uids))
	if len(uids) == 0 {
		return bentos, nil
	}
	err := getBaseQuery(ctx, s).Where("uid in (?)", uids).Find(&bentos).Error
	return bentos, err
}

func (s *bentoService) ListModelsFromManifests(ctx context.Context, bento *models.Bento) (models_ []*models.Model, err error) {
	if bento.Manifest == nil {
		return
	}
	if bento.Manifest.Models == nil {
		return
	}
	modelNames := make([]string, 0, len(bento.Manifest.Models))
	modelNamesSeen := make(map[string]struct{}, len(bento.Manifest.Models))
	for _, modelTag := range bento.Manifest.Models {
		modelName, _, _ := xstrings.Partition(modelTag, ":")
		if _, ok := modelNamesSeen[modelName]; !ok {
			modelNames = append(modelNames, modelName)
		}
	}
	if len(modelNames) == 0 {
		return
	}
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return
	}
	modelRepositories, _, err := ModelRepositoryService.List(ctx, ListModelRepositoryOption{
		OrganizationId: utils.UintPtr(bentoRepository.OrganizationId),
		Names:          &modelNames,
	})
	if len(modelRepositories) == 0 {
		return
	}
	modelRepositoryNameMapping := make(map[string]*models.ModelRepository, len(modelRepositories))
	for _, modelRepository := range modelRepositories {
		modelRepositoryNameMapping[modelRepository.Name] = modelRepository
	}
	for _, modelTag := range bento.Manifest.Models {
		modelRepositoryName, _, version := xstrings.Partition(modelTag, ":")
		if modelRepository, ok := modelRepositoryNameMapping[modelRepositoryName]; ok {
			var model *models.Model
			model, err = ModelService.GetByVersion(ctx, modelRepository.ID, version)
			if err != nil {
				return
			}
			models_ = append(models_, model)
		}
	}
	return
}

func (s *bentoService) List(ctx context.Context, opt ListBentoOption) ([]*models.Bento, uint, error) {
	query := getBaseQuery(ctx, s)
	query = query.Joins("LEFT JOIN bento_repository ON bento.bento_repository_id = bento_repository.id")
	if opt.ModelIds != nil {
		query = query.Joins("LEFT JOIN bento_model_rel ON bento_model_rel.bento_id = bento.id").Where("bento_model_rel.model_id in (?)", *opt.ModelIds)
	}
	if opt.OrganizationId != nil {
		query = query.Where("bento_repository.organization_id = ?", *opt.OrganizationId)
	}
	if opt.BentoRepositoryId != nil {
		query = query.Where("bento.bento_repository_id = ?", *opt.BentoRepositoryId)
	}
	if opt.Ids != nil {
		query = query.Where("bento.id in (?)", *opt.Ids)
	}
	if opt.Versions != nil {
		query = query.Where("bento.version in (?)", *opt.Versions)
	}
	if opt.CreatorId != nil {
		query = query.Where("bento.creator_id = ?", *opt.CreatorId)
	}
	if opt.Names != nil {
		query = query.Where("bento_repository.name in (?)", *opt.Names)
	}
	if opt.CreatorIds != nil {
		query = query.Where("bento.creator_id in (?)", *opt.CreatorIds)
	}
	query = opt.BindQueryWithKeywords(query, "bento_repository")
	query = opt.BindQueryWithLabels(query, modelschemas.ResourceTypeBento)
	query = query.Select("distinct(bento.*)")
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	query = s.getBaseDB(ctx).Table("(?) as bento", query)
	bentos := make([]*models.Bento, 0)
	query = opt.BindQueryWithLimit(query)
	if opt.Order != nil {
		query = query.Order(*opt.Order)
	} else {
		query = query.Order("bento.build_at DESC")
	}
	err = query.Find(&bentos).Error
	if err != nil {
		return nil, 0, err
	}
	return bentos, uint(total), err
}

func (s *bentoService) GroupByBentoRepositoryIds(ctx context.Context, bentoRepositoryIds []uint, count uint) (map[uint][]*models.Bento, error) {
	db := mustGetSession(ctx)

	query := db.Raw(`select bento_repository_id, id from bento where bento_repository_id in (?) order by id desc`, bentoRepositoryIds)

	type Item struct {
		BentoRepositoryId uint `gorm:"column:bento_repository_id"`
		Id                uint `gorm:"column:id"`
	}

	items := make([]*Item, 0)
	err := query.Find(&items).Error
	if err != nil {
		return nil, err
	}

	ids := make([]uint, 0)
	idsMap := make(map[uint][]uint)
	for _, item := range items {
		ids_ := idsMap[item.BentoRepositoryId]
		if len(ids_) < int(count) {
			ids = append(ids, item.Id)
			ids_ = append(ids_, item.Id)
			idsMap[item.BentoRepositoryId] = ids_
		}
	}

	bentos, _, err := s.List(ctx, ListBentoOption{
		Ids: &ids,
	})
	if err != nil {
		return nil, err
	}
	bentoMap := make(map[uint]*models.Bento)
	for _, bento := range bentos {
		bentoMap[bento.ID] = bento
	}

	res := make(map[uint][]*models.Bento)

	for bentoRepositoryId, ids_ := range idsMap {
		bentos_ := make([]*models.Bento, 0)
		for _, id := range ids_ {
			bentos_ = append(bentos_, bentoMap[id])
		}
		res[bentoRepositoryId] = bentos_
	}

	return res, nil
}

func (s *bentoService) CountByBentoRepositoryIds(ctx context.Context, bentoRepositoryIds []uint) (map[uint]uint, error) {
	db := mustGetSession(ctx)

	query := db.Raw(`select bento_repository_id, count(1) as count from bento
						where bento_repository_id in (?) group by bento_repository_id
					`, bentoRepositoryIds)

	type Item struct {
		BentoRepositoryId uint `gorm:"column:bento_repository_id"`
		Count             uint `gorm:"column:count"`
	}

	items := make([]*Item, 0, len(bentoRepositoryIds))
	err := query.Find(&items).Error
	if err != nil {
		return nil, err
	}

	res := make(map[uint]uint, len(items))
	for _, item := range items {
		res[item.BentoRepositoryId] = item.Count
	}

	return res, err
}

func (s *bentoService) ListLatestByBentoRepositoryIds(ctx context.Context, bentoRepositoryIds []uint) ([]*models.Bento, error) {
	db := mustGetSession(ctx)

	query := db.Raw(`select * from bento where id in (
					select n.bento_id from (
						select bento_repository_id, max(id) as bento_id from bento
						where bento_repository_id in (?) group by bento_repository_id
					) as n)`, bentoRepositoryIds)

	bentos := make([]*models.Bento, 0, len(bentoRepositoryIds))
	err := query.Find(&bentos).Error
	if err != nil {
		return nil, err
	}

	return bentos, err
}

func (s *bentoService) GetImageBuilderKubeLabels(ctx context.Context, bento *models.Bento) (map[string]string, error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		commonconsts.KubeLabelIsBentoImageBuilder:  "true",
		commonconsts.KubeLabelYataiBentoRepository: bentoRepository.Name,
		commonconsts.KubeLabelYataiBento:           bento.Version,
	}, nil
}

func (s *bentoService) ListImageBuildStatusUnsynced(ctx context.Context) ([]*models.Bento, error) {
	q := getBaseQuery(ctx, s)
	now := time.Now()
	t := now.Add(-time.Minute)
	q = q.Where("image_build_status != ? and (image_build_status_syncing_at is null or image_build_status_syncing_at < ? or image_build_status_updated_at is null or image_build_status_updated_at < ?)", modelschemas.ImageBuildStatusSuccess, t, t)
	bentos := make([]*models.Bento, 0)
	err := q.Order("id DESC").Find(&bentos).Error
	return bentos, err
}

type IBentoAssociate interface {
	GetAssociatedBentoId() uint
	GetAssociatedBentoCache() *models.Bento
	SetAssociatedBentoCache(version *models.Bento)
}

func (s *bentoService) GetAssociatedBento(ctx context.Context, associate IBentoAssociate) (*models.Bento, error) {
	cache := associate.GetAssociatedBentoCache()
	if cache != nil {
		return cache, nil
	}
	bento, err := s.Get(ctx, associate.GetAssociatedBentoId())
	associate.SetAssociatedBentoCache(bento)
	return bento, err
}

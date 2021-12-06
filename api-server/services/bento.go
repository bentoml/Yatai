package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/huandu/xstrings"

	"github.com/bentoml/yatai/common/utils"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/iancoleman/strcase"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
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
	ImageBuildStatus          *modelschemas.BentoImageBuildStatus
	ImageBuildStatusSyncingAt **time.Time
	ImageBuildStatusUpdatedAt **time.Time
	UploadStatus              *modelschemas.BentoUploadStatus
	UploadStartedAt           **time.Time
	UploadFinishedAt          **time.Time
	UploadFinishedReason      *string
	Labels                    *modelschemas.LabelItemsSchema
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
		ImageBuildStatus: modelschemas.BentoImageBuildStatusPending,
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

func (s *bentoService) GetImageName(ctx context.Context, bento *models.Bento, inCluster bool) (string, error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return "", nil
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return "", nil
	}
	dockerRegistry, err := OrganizationService.GetDockerRegistry(ctx, org)
	if err != nil {
		return "", err
	}
	var imageName string
	if inCluster {
		imageName = fmt.Sprintf("%s:yatai.%s.%s.%s", dockerRegistry.BentosRepositoryURIInCluster, org.Name, bentoRepository.Name, bento.Version)
	} else {
		imageName = fmt.Sprintf("%s:yatai.%s.%s.%s", dockerRegistry.BentosRepositoryURI, org.Name, bentoRepository.Name, bento.Version)
	}
	return imageName, nil
}

func (s *bentoService) GetS3BucketName(ctx context.Context, bento *models.Bento) (string, error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return "", nil
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return "", nil
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

	if len(updaters) == 0 {
		return bento, nil
	}

	// nolint: ineffassign,staticcheck
	db, ctx, df, err := startTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { df(err) }()

	err = db.Model(&models.Bento{}).Where("id = ?", bento.ID).Updates(updaters).Error
	if err != nil {
		return nil, err
	}

	if opt.Labels != nil {
		bento, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
		if err != nil {
			return nil, err
		}
		org, err := OrganizationService.GetAssociatedOrganization(ctx, bento)
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

	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return nil, err
	}

	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return nil, err
	}

	majorCluster, err := OrganizationService.GetMajorCluster(ctx, org)
	if err != nil {
		return nil, err
	}

	kubeCli, _, err := ClusterService.GetKubeCliSet(ctx, majorCluster)
	if err != nil {
		return nil, err
	}

	kubeNamespace := consts.KubeNamespaceYataiBentoImageBuilder

	_, err = KubeNamespaceService.MakeSureNamespace(ctx, majorCluster, kubeNamespace)
	if err != nil {
		return nil, err
	}

	s3Config, err := OrganizationService.GetS3Config(ctx, org)
	if err != nil {
		return nil, err
	}

	dockerRegistry, err := OrganizationService.GetDockerRegistry(ctx, org)
	if err != nil {
		return nil, err
	}

	dockerCMKubeName := "docker-config"
	cmObj := struct {
		Auths map[string]struct {
			Auth string `json:"auth"`
		} `json:"auths,omitempty"`
	}{}

	if dockerRegistry.Username != "" {
		cmObj.Auths = map[string]struct {
			Auth string `json:"auth"`
		}{
			dockerRegistry.Server: {
				Auth: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", dockerRegistry.Username, dockerRegistry.Password))),
			},
		}
	}

	dockerCMContent, err := json.Marshal(cmObj)
	if err != nil {
		return nil, err
	}
	cmsCli := kubeCli.CoreV1().ConfigMaps(kubeNamespace)
	oldCm, err := cmsCli.Get(ctx, dockerCMKubeName, metav1.GetOptions{})
	// nolint: gocritic
	if apierrors.IsNotFound(err) {
		_, err = cmsCli.Create(ctx, &apiv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: dockerCMKubeName},
			Data: map[string]string{
				"config.json": string(dockerCMContent),
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		oldCm.Data["config.json"] = string(dockerCMContent)
		_, err = cmsCli.Update(ctx, oldCm, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
	}

	volumeMounts := []apiv1.VolumeMount{
		{
			Name:      dockerCMKubeName,
			MountPath: "/kaniko/.docker/",
		},
	}

	// nolint: goconst
	s3ForcePath := "true"
	if s3Config.Endpoint == consts.AmazonS3Endpoint {
		// nolint: goconst
		s3ForcePath = "false"
	}

	envs := []apiv1.EnvVar{
		{
			Name:  "AWS_ACCESS_KEY_ID",
			Value: s3Config.AccessKey,
		},
		{
			Name:  "AWS_SECRET_ACCESS_KEY",
			Value: s3Config.SecretKey,
		},
		{
			Name:  "AWS_REGION",
			Value: s3Config.Region,
		},
		{
			Name:  "S3_ENDPOINT",
			Value: s3Config.EndpointWithSchemeInCluster,
		},
		{
			Name:  "S3_FORCE_PATH_STYLE",
			Value: s3ForcePath,
		},
	}

	podsCli := kubeCli.CoreV1().Pods(kubeNamespace)

	kubeName, err := s.GetImageBuilderKubeName(ctx, bento)
	if err != nil {
		return nil, err
	}

	s3ObjectName, err := s.getS3ObjectName(ctx, bento)
	if err != nil {
		return nil, err
	}

	imageName, err := s.GetImageName(ctx, bento, true)
	if err != nil {
		return nil, err
	}

	s3BucketName, err := s.GetS3BucketName(ctx, bento)
	if err != nil {
		return nil, err
	}

	err = s3Config.MakeSureBucket(ctx, s3BucketName)
	if err != nil {
		return nil, err
	}

	args := []string{
		"--dockerfile=./env/docker/Dockerfile",
		fmt.Sprintf("--context=s3://%s/%s", s3BucketName, s3ObjectName),
		fmt.Sprintf("--destination=%s", imageName),
	}

	if !dockerRegistry.Secure {
		args = append(args, "--insecure")
	}

	_, err = podsCli.Get(ctx, kubeName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = podsCli.Create(ctx, &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: kubeName,
				Labels: map[string]string{
					consts.KubeLabelYataiBentoRepository: bentoRepository.Name,
					consts.KubeLabelYataiBento:           bento.Version,
				},
			},
			Spec: apiv1.PodSpec{
				RestartPolicy: apiv1.RestartPolicyNever,
				Volumes: []apiv1.Volume{
					{
						Name: dockerCMKubeName,
						VolumeSource: apiv1.VolumeSource{
							ConfigMap: &apiv1.ConfigMapVolumeSource{
								LocalObjectReference: apiv1.LocalObjectReference{
									Name: dockerCMKubeName,
								},
							},
						},
					},
				},
				Containers: []apiv1.Container{
					{
						Name:         "builder",
						Image:        "gcr.io/kaniko-project/executor:latest",
						Args:         args,
						VolumeMounts: volumeMounts,
						Env:          envs,
						TTY:          true,
						Stdin:        true,
					},
				},
			},
		}, metav1.CreateOptions{})

		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	go func() {
		_, _ = s.SyncImageBuilderStatus(ctx, bento)
	}()

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

	return strings.ReplaceAll(strcase.ToKebab(fmt.Sprintf("yatai-bento-image-builder-%s-%s-%s", org.Name, bentoRepository.Name, bento.Version)), ".", "-"), nil
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

func (s *bentoService) ListImageBuilderPods(ctx context.Context, bento *models.Bento) ([]*models.KubePodWithStatus, error) {
	bentoRepository, err := BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
	if err != nil {
		return nil, err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
	if err != nil {
		return nil, err
	}
	cluster, err := OrganizationService.GetMajorCluster(ctx, org)
	if err != nil {
		return nil, err
	}
	_, podLister, err := GetPodInformer(ctx, cluster, consts.KubeNamespaceYataiBentoImageBuilder)
	if err != nil {
		return nil, err
	}

	selector, err := labels.Parse(fmt.Sprintf("%s = %s, %s = %s", consts.KubeLabelYataiBentoRepository, bentoRepository.Name, consts.KubeLabelYataiBento, bento.Version))
	if err != nil {
		return nil, err
	}

	pods, err := podLister.List(selector)
	if err != nil {
		return nil, err
	}

	_, eventLister, err := GetEventInformer(ctx, cluster, consts.KubeNamespaceYataiBentoImageBuilder)
	if err != nil {
		return nil, err
	}

	events, err := eventLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	pods_ := make([]apiv1.Pod, 0, len(pods))
	for _, p := range pods {
		pods_ = append(pods_, *p)
	}

	events_ := make([]apiv1.Event, 0, len(pods))
	for _, e := range events {
		events_ = append(events_, *e)
	}

	return KubePodService.MapKubePodsToKubePodWithStatuses(ctx, pods_, events_), nil
}

func (s *bentoService) CalculateImageBuildStatus(ctx context.Context, bento *models.Bento) (modelschemas.BentoImageBuildStatus, error) {
	defaultStatus := modelschemas.BentoImageBuildStatusPending
	pods, err := s.ListImageBuilderPods(ctx, bento)
	if err != nil {
		return defaultStatus, err
	}

	if len(pods) == 0 {
		return defaultStatus, nil
	}

	for _, p := range pods {
		if p.Status.Status == modelschemas.KubePodActualStatusRunning || p.Status.Status == modelschemas.KubePodActualStatusPending {
			return modelschemas.BentoImageBuildStatusBuilding, nil
		}
		if p.Status.Status == modelschemas.KubePodActualStatusTerminating || p.Status.Status == modelschemas.KubePodActualStatusUnknown || p.Status.Status == modelschemas.KubePodActualStatusFailed {
			return modelschemas.BentoImageBuildStatusFailed, nil
		}
	}

	hasPending := false
	hasFailed := false
	hasBuilding := false

	models_, err := s.ListModelsFromManifests(ctx, bento)
	if err != nil {
		return defaultStatus, err
	}

	for _, model := range models_ {
		if model.ImageBuildStatus == modelschemas.ModelImageBuildStatusPending {
			hasPending = true
		}
		if model.ImageBuildStatus == modelschemas.ModelImageBuildStatusFailed {
			hasFailed = true
		}
		if model.ImageBuildStatus == modelschemas.ModelImageBuildStatusBuilding {
			hasBuilding = true
		}
	}

	if hasFailed {
		return modelschemas.BentoImageBuildStatusFailed, nil
	}

	if hasBuilding || hasPending {
		return modelschemas.BentoImageBuildStatusBuilding, nil
	}

	return modelschemas.BentoImageBuildStatusSuccess, nil
}

func (s *bentoService) ListImageBuildStatusUnsynced(ctx context.Context) ([]*models.Bento, error) {
	q := getBaseQuery(ctx, s)
	now := time.Now()
	t := now.Add(-time.Minute)
	q = q.Where("image_build_status != ? and (image_build_status_syncing_at is null or image_build_status_syncing_at < ? or image_build_status_updated_at is null or image_build_status_updated_at < ?)", modelschemas.BentoImageBuildStatusSuccess, t, t)
	bentos := make([]*models.Bento, 0)
	err := q.Order("id DESC").Find(&bentos).Error
	return bentos, err
}

func (s *bentoService) SyncImageBuilderStatus(ctx context.Context, bento *models.Bento) (modelschemas.BentoImageBuildStatus, error) {
	now := time.Now()
	nowPtr := &now
	_, err := s.Update(ctx, bento, UpdateBentoOption{
		ImageBuildStatusSyncingAt: &nowPtr,
	})
	if err != nil {
		return bento.ImageBuildStatus, err
	}
	currentStatus, err := s.CalculateImageBuildStatus(ctx, bento)
	if err != nil {
		return bento.ImageBuildStatus, err
	}
	now = time.Now()
	nowPtr = &now
	_, err = s.Update(ctx, bento, UpdateBentoOption{
		ImageBuildStatus:          &currentStatus,
		ImageBuildStatusUpdatedAt: &nowPtr,
	})
	if err != nil {
		return currentStatus, err
	}
	return currentStatus, nil
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

package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/schemas/modelschemas"
)

type modelVersionService struct{}

var ModelVersionService = &modelVersionService{}

func (s *modelVersionService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.ModelVersion{})
}

type CreateModelVersionOption struct {
	CreatorId   uint
	ModelId     uint
	Version     string
	Description string
	BuildAt     time.Time
	Manifest    *modelschemas.ModelVersionManifestSchema
	Labels      modelschemas.LabelItemsSchema
}

type UpdateModelVersionOption struct {
	ImageBuildStatus          *modelschemas.ModelVersionImageBuildStatus
	ImageBuildStatusSyncingAt **time.Time
	ImageBuildStatusUpdatedAt **time.Time
	UploadStatus              *modelschemas.ModelVersionUploadStatus
	UploadStartedAt           **time.Time
	UploadFinishedAt          **time.Time
	UploadFinishedReason      *string
	Labels                    *modelschemas.LabelItemsSchema
}

type ListModelVersionOption struct {
	BaseListOption
	BaseListByLabelsOption
	ModelId         *uint
	Versions        *[]string
	BentoVersionIds *[]uint
	OrganizationId  *uint
	CreatorId       *uint
	CreatorIds      *[]uint
	Order           *string
	Names           *[]string
}

func (s *modelVersionService) Create(ctx context.Context, opt CreateModelVersionOption) (modelVersion *models.ModelVersion, err error) {
	// nolint: ineffassign, staticcheck
	db, ctx, df, err := startTransaction(ctx)
	if err != nil {
		return
	}
	defer func() { df(err) }()
	modelVersion = &models.ModelVersion{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		ModelAssociate: models.ModelAssociate{
			ModelId: opt.ModelId,
		},
		Version:          opt.Version,
		Description:      opt.Description,
		ImageBuildStatus: modelschemas.ModelVersionImageBuildStatusPending,
		UploadStatus:     modelschemas.ModelVersionUploadStatusPending,
		BuildAt:          opt.BuildAt,
		Manifest:         opt.Manifest,
	}

	err = db.Create(modelVersion).Error
	if err != nil {
		return
	}

	var model *models.Model
	model, err = ModelService.GetAssociatedModel(ctx, modelVersion)
	if err != nil {
		return
	}
	var org *models.Organization
	org, err = OrganizationService.GetAssociatedOrganization(ctx, model)
	if err != nil {
		return
	}
	var user *models.User
	user, err = GetCurrentUser(ctx)
	if err != nil {
		return
	}
	err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, opt.Labels, user.ID, org.ID, modelVersion)
	return
}

func (s *modelVersionService) PreSignS3UploadUrl(ctx context.Context, modelVersion *models.ModelVersion) (url *url.URL, err error) {
	model, err := ModelService.GetAssociatedModel(ctx, modelVersion)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, model)
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

	bucketName, err := s.GetS3BucketName(ctx, modelVersion)
	if err != nil {
		return
	}

	err = s3Config.MakeSureBucket(ctx, bucketName)
	if err != nil {
		return
	}

	objectName, err := s.getS3ObjectName(ctx, modelVersion)
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

func (s *modelVersionService) PreSignS3DownloadUrl(ctx context.Context, modelVersion *models.ModelVersion) (url *url.URL, err error) {
	model, err := ModelService.GetAssociatedModel(ctx, modelVersion)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, model)
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

	bucketName, err := s.GetS3BucketName(ctx, modelVersion)
	if err != nil {
		return
	}

	err = s3Config.MakeSureBucket(ctx, bucketName)
	if err != nil {
		return
	}

	objectName, err := s.getS3ObjectName(ctx, modelVersion)
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

func (s *modelVersionService) getS3ObjectName(ctx context.Context, modelVersion *models.ModelVersion) (string, error) {
	model, err := ModelService.GetAssociatedModel(ctx, modelVersion)
	if err != nil {
		return "", err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, model)
	if err != nil {
		return "", err
	}
	objectName := fmt.Sprintf("models/%s/%s/%s.tar.gz", org.Name, model.Name, modelVersion.Version)
	return objectName, nil
}

func (s *modelVersionService) GetImageName(ctx context.Context, modelVersion *models.ModelVersion, inCluster bool) (string, error) {
	model, err := ModelService.GetAssociatedModel(ctx, modelVersion)
	if err != nil {
		return "", nil
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, model)
	if err != nil {
		return "", nil
	}
	dockerRegistry, err := OrganizationService.GetDockerRegistry(ctx, org)
	if err != nil {
		return "", err
	}
	var imageName string
	if inCluster {
		imageName = fmt.Sprintf("%s:yatai.%s.%s.%s", dockerRegistry.ModelsRepositoryURIInCluster, org.Name, model.Name, modelVersion.Version)
	} else {
		imageName = fmt.Sprintf("%s:yatai.%s.%s.%s", dockerRegistry.ModelsRepositoryURI, org.Name, model.Name, modelVersion.Version)
	}
	return imageName, nil
}

func (s *modelVersionService) GetS3BucketName(ctx context.Context, modelVersion *models.ModelVersion) (string, error) {
	model, err := ModelService.GetAssociatedModel(ctx, modelVersion)
	if err != nil {
		return "", nil
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, model)
	if err != nil {
		return "", nil
	}

	s3Config, err := OrganizationService.GetS3Config(ctx, org)
	if err != nil {
		return "", err
	}

	s3BucketName := s3Config.ModelsBucketName

	return s3BucketName, nil
}

func (s *modelVersionService) Update(ctx context.Context, modelVersion *models.ModelVersion, opt UpdateModelVersionOption) (*models.ModelVersion, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.ImageBuildStatus != nil {
		updaters["image_build_status"] = *opt.ImageBuildStatus
		defer func() {
			if err != nil {
				modelVersion.ImageBuildStatus = *opt.ImageBuildStatus
			}
		}()
	}
	if opt.ImageBuildStatusSyncingAt != nil {
		updaters["image_build_status_syncing_at"] = *opt.ImageBuildStatusSyncingAt
		defer func() {
			if err != nil {
				modelVersion.ImageBuildStatusSyncingAt = *opt.ImageBuildStatusSyncingAt
			}
		}()
	}
	if opt.ImageBuildStatusUpdatedAt != nil {
		updaters["image_build_status_updated_at"] = *opt.ImageBuildStatusUpdatedAt
		defer func() {
			if err != nil {
				modelVersion.ImageBuildStatusUpdatedAt = *opt.ImageBuildStatusUpdatedAt
			}
		}()
	}
	if opt.UploadStatus != nil {
		updaters["upload_status"] = *opt.UploadStatus
		defer func() {
			if err != nil {
				modelVersion.UploadStatus = *opt.UploadStatus
			}
		}()
	}
	if opt.UploadStartedAt != nil {
		updaters["upload_started_at"] = *opt.UploadStartedAt
		defer func() {
			if err != nil {
				modelVersion.UploadStartedAt = *opt.UploadStartedAt
			}
		}()
	}
	if opt.UploadFinishedAt != nil {
		updaters["upload_finished_at"] = *opt.UploadFinishedAt
		defer func() {
			if err != nil {
				modelVersion.UploadFinishedAt = *opt.UploadFinishedAt
			}
		}()
	}
	if opt.UploadFinishedReason != nil {
		updaters["upload_finished_reason"] = *opt.UploadFinishedReason
		defer func() {
			if err != nil {
				modelVersion.UploadFinishedReason = *opt.UploadFinishedReason
			}
		}()
	}
	if len(updaters) == 0 {
		return modelVersion, nil
	}

	// nolint: ineffassign,staticcheck
	db, ctx, df, err := startTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { df(err) }()

	err = db.Model(&models.ModelVersion{}).Where("id = ?", modelVersion.ID).Updates(updaters).Error
	if err != nil {
		return nil, err
	}

	if opt.Labels != nil {
		model, err := ModelService.GetAssociatedModel(ctx, modelVersion)
		if err != nil {
			return nil, err
		}
		org, err := OrganizationService.GetAssociatedOrganization(ctx, model)
		if err != nil {
			return nil, err
		}
		user, err := GetCurrentUser(ctx)
		if err != nil {
			return nil, err
		}
		err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, *opt.Labels, user.ID, org.ID, modelVersion)
		if err != nil {
			return nil, err
		}
	}

	if opt.UploadStatus == nil || *opt.UploadStatus != modelschemas.ModelVersionUploadStatusSuccess {
		return modelVersion, nil
	}

	model, err := ModelService.GetAssociatedModel(ctx, modelVersion)
	if err != nil {
		return nil, err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, model)
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

	kubeNamespace := consts.KubeNamespaceYataiModelVersionImageBuilder

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

	dockerConfigObj := struct {
		Auths map[string]struct {
			Auth string `json:"auth"`
		} `json:"auths,omitempty"`
	}{}

	if dockerRegistry.Username != "" {
		dockerConfigObj.Auths = map[string]struct {
			Auth string `json:"auth"`
		}{
			dockerRegistry.Server: {
				Auth: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", dockerRegistry.Username, dockerRegistry.Password))),
			},
		}
	}

	dockerConfigContent, err := json.Marshal(dockerConfigObj)
	if err != nil {
		return nil, err
	}
	cmsCli := kubeCli.CoreV1().ConfigMaps(kubeNamespace)
	dockerConfigCMKubeName := "docker-config"
	oldDockerConfigCM, err := cmsCli.Get(ctx, dockerConfigCMKubeName, metav1.GetOptions{})
	// nolint: gocritic
	if apierrors.IsNotFound(err) {
		_, err = cmsCli.Create(ctx, &apiv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: dockerConfigCMKubeName},
			Data: map[string]string{
				"config.json": string(dockerConfigContent),
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		oldDockerConfigCM.Data["config.json"] = string(dockerConfigContent)
		_, err = cmsCli.Update(ctx, oldDockerConfigCM, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
	}

	dockerFileCMKubeName := fmt.Sprintf("docker-file-%d", modelVersion.ID)
	dockerFileContent := `
FROM scratch

COPY . /model
`
	oldDockerFileCM, err := cmsCli.Get(ctx, dockerFileCMKubeName, metav1.GetOptions{})
	// nolint: gocritic
	if apierrors.IsNotFound(err) {
		_, err = cmsCli.Create(ctx, &apiv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: dockerFileCMKubeName},
			Data: map[string]string{
				"Dockerfile": dockerFileContent,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		oldDockerFileCM.Data["Dockerfile"] = dockerFileContent
		_, err = cmsCli.Update(ctx, oldDockerFileCM, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
	}

	volumes := []apiv1.Volume{
		{
			Name: dockerConfigCMKubeName,
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: dockerConfigCMKubeName,
					},
				},
			},
		},
		{
			Name: dockerFileCMKubeName,
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: dockerFileCMKubeName,
					},
				},
			},
		},
	}

	volumeMounts := []apiv1.VolumeMount{
		{
			Name:      dockerConfigCMKubeName,
			MountPath: "/kaniko/.docker/",
		},
		{
			Name:      dockerFileCMKubeName,
			MountPath: "/docker/",
		},
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
			Value: "true",
		},
	}

	podsCli := kubeCli.CoreV1().Pods(kubeNamespace)

	kubeName, err := s.GetImageBuilderKubeName(ctx, modelVersion)
	if err != nil {
		return nil, err
	}

	s3ObjectName, err := s.getS3ObjectName(ctx, modelVersion)
	if err != nil {
		return nil, err
	}

	imageName, err := s.GetImageName(ctx, modelVersion, true)
	if err != nil {
		return nil, err
	}

	s3BucketName, err := s.GetS3BucketName(ctx, modelVersion)
	if err != nil {
		return nil, err
	}

	err = s3Config.MakeSureBucket(ctx, s3BucketName)
	if err != nil {
		return nil, err
	}

	args := []string{
		"--dockerfile=/docker/Dockerfile",
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
					consts.KubeLabelYataiModel:        model.Name,
					consts.KubeLabelYataiModelVersion: modelVersion.Version,
				},
			},
			Spec: apiv1.PodSpec{
				RestartPolicy: apiv1.RestartPolicyNever,
				Volumes:       volumes,
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
		_, _ = s.SyncImageBuilderStatus(ctx, modelVersion)
	}()

	return modelVersion, nil
}

func (s *modelVersionService) GetImageBuilderKubeName(ctx context.Context, modelVersion *models.ModelVersion) (string, error) {
	model, err := ModelService.GetAssociatedModel(ctx, modelVersion)
	if err != nil {
		return "", err
	}

	org, err := OrganizationService.GetAssociatedOrganization(ctx, model)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(strcase.ToKebab(fmt.Sprintf("yatai-model-image-builder-%s-%s-%s", org.Name, model.Name, modelVersion.Version)), ".", "-"), nil
}

func (s *modelVersionService) Get(ctx context.Context, id uint) (*models.ModelVersion, error) {
	var modelVersion models.ModelVersion
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&modelVersion).Error
	if err != nil {
		return nil, err
	}
	if modelVersion.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &modelVersion, nil
}

func (s *modelVersionService) GetByUid(ctx context.Context, uid string) (*models.ModelVersion, error) {
	var modelVersion models.ModelVersion
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&modelVersion).Error
	if err != nil {
		return nil, err
	}
	if modelVersion.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &modelVersion, nil
}

func (s *modelVersionService) GetByVersion(ctx context.Context, modelId uint, version string) (*models.ModelVersion, error) {
	var modelVersion models.ModelVersion
	err := getBaseQuery(ctx, s).Where("model_id = ?", modelId).Where("version = ?", version).First(&modelVersion).Error
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get model version by model id %d and version %s", modelId, version)
	}
	if modelVersion.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &modelVersion, nil
}

func (s *modelVersionService) ListByUids(ctx context.Context, uids []uint) ([]*models.ModelVersion, error) {
	modelVersions := make([]*models.ModelVersion, 0, len(uids))
	if len(uids) == 0 {
		return modelVersions, nil
	}
	err := getBaseQuery(ctx, s).Where("id in (?)", uids).Find(&modelVersions).Error
	return modelVersions, err
}

func (s *modelVersionService) List(ctx context.Context, opt ListModelVersionOption) ([]*models.ModelVersion, uint, error) {
	query := getBaseQuery(ctx, s)
	query = query.Joins("LEFT JOIN model ON model_version.model_id = model.id")
	if opt.BentoVersionIds != nil {
		query = query.Joins("LEFT JOIN bento_version_model_version_rel ON bento_version_model_version_rel.model_version_id = model_version.id").Where("bento_version_model_version_rel.bento_version_id in (?)", *opt.BentoVersionIds)
	}
	if opt.OrganizationId != nil {
		query = query.Where("model.organization_id = ?", *opt.OrganizationId)
	}
	if opt.Versions != nil {
		query = query.Where("model_version.version in (?)", *opt.Versions)
	}
	if opt.ModelId != nil {
		query = query.Where("model_version.model_id = ?", *opt.ModelId)
	}
	if opt.CreatorId != nil {
		query = query.Where("model_version.creator_id = ?", *opt.CreatorId)
	}
	if opt.Names != nil {
		query = query.Where("model.name in (?)", *opt.Names)
	}
	if opt.CreatorIds != nil {
		query = query.Where("model_version.creator_id in (?)", *opt.CreatorIds)
	}
	query = opt.BindQueryWithKeywords(query, "model")
	query = opt.BindQueryWithLabels(query, modelschemas.ResourceTypeModelVersion)
	query = query.Select("distinct(model_version.*)")
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	modelVersions := make([]*models.ModelVersion, 0)
	query = opt.BindQueryWithLimit(query)
	if opt.Order != nil {
		query = query.Order(*opt.Order)
	} else {
		query = query.Order("model_version.build_at DESC")
	}
	err = query.Find(&modelVersions).Error
	if err != nil {
		return nil, 0, err
	}
	return modelVersions, uint(total), nil
}

func (s *modelVersionService) ListLatestByModelIds(ctx context.Context, modelIds []uint) ([]*models.ModelVersion, error) {
	db := mustGetSession(ctx)

	query := db.Raw(`select * from model_version where id in (
			select n.version_id from (
				select model_id, max(id) as version_id from model_version
				where model_id in (?) group by model_id
			) as n)`, modelIds)
	versions := make([]*models.ModelVersion, 0, len(modelIds))
	err := query.Find(&versions).Error
	if err != nil {
		return nil, err
	}
	return versions, err
}

func (s *modelVersionService) ListImageBuilderPods(ctx context.Context, modelVersion *models.ModelVersion) ([]*models.KubePodWithStatus, error) {
	model, err := ModelService.GetAssociatedModel(ctx, modelVersion)
	if err != nil {
		return nil, err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, model)
	if err != nil {
		return nil, err
	}
	cluster, err := OrganizationService.GetMajorCluster(ctx, org)
	if err != nil {
		return nil, err
	}
	_, podLister, err := GetPodInformer(ctx, cluster, consts.KubeNamespaceYataiModelVersionImageBuilder)
	if err != nil {
		return nil, err
	}

	selector, err := labels.Parse(fmt.Sprintf("%s = %s, %s = %s", consts.KubeLabelYataiModel, model.Name, consts.KubeLabelYataiModelVersion, modelVersion.Version))
	if err != nil {
		return nil, err
	}
	pods, err := podLister.List(selector)
	if err != nil {
		return nil, err
	}
	_, eventLister, err := GetEventInformer(ctx, cluster, consts.KubeNamespaceYataiModelVersionImageBuilder)
	if err != nil {
		return nil, err
	}

	events, err := eventLister.List(selector)
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

func (s *modelVersionService) CalculateImageBuildStatus(ctx context.Context, modelVersion *models.ModelVersion) (modelschemas.ModelVersionImageBuildStatus, error) {
	defaultStatus := modelschemas.ModelVersionImageBuildStatusPending
	pods, err := s.ListImageBuilderPods(ctx, modelVersion)
	if err != nil {
		return defaultStatus, err
	}

	if len(pods) == 0 {
		return defaultStatus, nil
	}

	for _, p := range pods {
		if p.Status.Status == modelschemas.KubePodActualStatusRunning || p.Status.Status == modelschemas.KubePodActualStatusPending {
			return modelschemas.ModelVersionImageBuildStatusBuilding, nil
		}
		if p.Status.Status == modelschemas.KubePodActualStatusTerminating || p.Status.Status == modelschemas.KubePodActualStatusUnknown || p.Status.Status == modelschemas.KubePodActualStatusFailed {
			return modelschemas.ModelVersionImageBuildStatusFailed, nil
		}
	}

	return modelschemas.ModelVersionImageBuildStatusSuccess, nil
}

func (s *modelVersionService) ListImageBuildStatusUnsynced(ctx context.Context) ([]*models.ModelVersion, error) {
	q := getBaseQuery(ctx, s)
	now := time.Now()
	t := now.Add(-time.Minute)
	q = q.Where("image_build_status != ? and (image_build_status_syncing_at is null or image_build_status_syncing_at < ? or image_build_status_updated_at is null or image_build_status_updated_at < ?)", modelschemas.ModelVersionImageBuildStatusSuccess, t, t)
	modelVersions := make([]*models.ModelVersion, 0)
	err := q.Order("id DESC").Find(&modelVersions).Error
	return modelVersions, err
}

func (s *modelVersionService) SyncImageBuilderStatus(ctx context.Context, modelVersion *models.ModelVersion) (modelschemas.ModelVersionImageBuildStatus, error) {
	now := time.Now()
	nowPtr := &now
	_, err := s.Update(ctx, modelVersion, UpdateModelVersionOption{
		ImageBuildStatusSyncingAt: &nowPtr,
	})
	if err != nil {
		return modelVersion.ImageBuildStatus, err
	}
	currentStatus, err := s.CalculateImageBuildStatus(ctx, modelVersion)
	if err != nil {
		return modelVersion.ImageBuildStatus, err
	}
	now = time.Now()
	nowPtr = &now
	_, err = s.Update(ctx, modelVersion, UpdateModelVersionOption{
		ImageBuildStatus:          &currentStatus,
		ImageBuildStatusUpdatedAt: &nowPtr,
	})
	if err != nil {
		return currentStatus, err
	}
	if currentStatus == modelschemas.ModelVersionImageBuildStatusSuccess {
		bentoVersions, _, err := BentoVersionService.List(ctx, ListBentoVersionOption{
			ModelVersionIds: &[]uint{modelVersion.ID},
		})
		if err != nil {
			return currentStatus, err
		}
		for _, bv := range bentoVersions {
			bv := bv
			go func() {
				_, _ = BentoVersionService.SyncImageBuilderStatus(ctx, bv)
			}()
		}
	}
	return currentStatus, nil
}

type IModelVersionAssociated interface {
	GetAssociatedModelVersionId() uint
	GetAssociatedModelVersionCache() *models.ModelVersion
	SetAssociatedModelVersionCache(version *models.ModelVersion)
}

func (s *modelVersionService) GetAssociatedModelVersion(ctx context.Context, associate IModelVersionAssociated) (*models.ModelVersion, error) {
	cache := associate.GetAssociatedModelVersionCache()
	if cache != nil {
		return cache, nil
	}
	version, err := s.Get(ctx, associate.GetAssociatedModelVersionId())
	associate.SetAssociatedModelVersionCache(version)
	return version, err
}

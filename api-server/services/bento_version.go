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

// nolint:gosec
var awsSecretTemplate = `
[default]
aws_access_key_id = {{.AccessKeyId}}
aws_secret_access_key = {{.SecretAccessKey}}
`

type bentoVersionService struct{}

var BentoVersionService = bentoVersionService{}

func (s *bentoVersionService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.BentoVersion{})
}

type CreateBentoVersionOption struct {
	CreatorId   uint
	BentoId     uint
	Version     string
	Description string
	BuildAt     time.Time
	Manifest    *modelschemas.BentoVersionManifestSchema
}

type UpdateBentoVersionOption struct {
	ImageBuildStatus          *modelschemas.BentoVersionImageBuildStatus
	ImageBuildStatusSyncingAt **time.Time
	ImageBuildStatusUpdatedAt **time.Time
	UploadStatus              *modelschemas.BentoVersionUploadStatus
	UploadStartedAt           **time.Time
	UploadFinishedAt          **time.Time
	UploadFinishedReason      *string
}

type ListBentoVersionOption struct {
	BaseListOption
	OrganizationId  *uint
	BentoId         *uint
	Versions        *[]string
	ModelVersionIds *[]uint
	CreatorId       *uint
	CreatorIds      *[]uint
	Order           *string
	Names           *[]string
}

func (s *bentoVersionService) Create(ctx context.Context, opt CreateBentoVersionOption) (bentoVersion *models.BentoVersion, err error) {
	// nolint: ineffassign,staticcheck
	db, ctx, df, err := startTransaction(ctx)
	if err != nil {
		return
	}
	defer func() { df(err) }()
	bentoVersion = &models.BentoVersion{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		BentoAssociate: models.BentoAssociate{
			BentoId: opt.BentoId,
		},
		Version:          opt.Version,
		Description:      opt.Description,
		ImageBuildStatus: modelschemas.BentoVersionImageBuildStatusPending,
		UploadStatus:     modelschemas.BentoVersionUploadStatusPending,
		BuildAt:          opt.BuildAt,
		Manifest:         opt.Manifest,
	}
	err = db.Create(bentoVersion).Error
	if err != nil {
		return
	}
	modelVersions, err := s.ListModelVersionsFromManifests(ctx, bentoVersion)
	if err != nil {
		return
	}
	for _, modelVersion := range modelVersions {
		rel := &models.BentoVersionModelVersionRel{
			BentoVersionAssociate: models.BentoVersionAssociate{
				BentoVersionId: bentoVersion.ID,
			},
			ModelVersionAssociate: models.ModelVersionAssociate{
				ModelVersionId: modelVersion.ID,
			},
		}
		err = db.Create(rel).Error
		if err != nil {
			return
		}
	}
	return
}

func (s *bentoVersionService) PreSignS3UploadUrl(ctx context.Context, bentoVersion *models.BentoVersion) (url *url.URL, err error) {
	bento, err := BentoService.GetAssociatedBento(ctx, bentoVersion)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bento)
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

	bucketName, err := s.GetS3BucketName(ctx, bentoVersion)
	if err != nil {
		return
	}

	err = s3Config.MakeSureBucket(ctx, bucketName)
	if err != nil {
		return
	}

	objectName, err := s.getS3ObjectName(ctx, bentoVersion)
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

func (s *bentoVersionService) PreSignS3DownloadUrl(ctx context.Context, bentoVersion *models.BentoVersion) (url *url.URL, err error) {
	bento, err := BentoService.GetAssociatedBento(ctx, bentoVersion)
	if err != nil {
		return
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bento)
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

	bucketName, err := s.GetS3BucketName(ctx, bentoVersion)
	if err != nil {
		return
	}

	err = s3Config.MakeSureBucket(ctx, bucketName)
	if err != nil {
		return
	}

	objectName, err := s.getS3ObjectName(ctx, bentoVersion)
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

func (s *bentoVersionService) getS3ObjectName(ctx context.Context, bentoVersion *models.BentoVersion) (string, error) {
	bento, err := BentoService.GetAssociatedBento(ctx, bentoVersion)
	if err != nil {
		return "", err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bento)
	if err != nil {
		return "", err
	}
	objectName := fmt.Sprintf("bentos/%s/%s/%s.tar.gz", org.Name, bento.Name, bentoVersion.Version)
	return objectName, nil
}

func (s *bentoVersionService) GetImageName(ctx context.Context, bentoVersion *models.BentoVersion, inCluster bool) (string, error) {
	bento, err := BentoService.GetAssociatedBento(ctx, bentoVersion)
	if err != nil {
		return "", nil
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bento)
	if err != nil {
		return "", nil
	}
	dockerRegistry, err := OrganizationService.GetDockerRegistry(ctx, org)
	if err != nil {
		return "", err
	}
	var imageName string
	if inCluster {
		imageName = fmt.Sprintf("%s:yatai.%s.%s.%s", dockerRegistry.BentosRepositoryURIInCluster, org.Name, bento.Name, bentoVersion.Version)
	} else {
		imageName = fmt.Sprintf("%s:yatai.%s.%s.%s", dockerRegistry.BentosRepositoryURI, org.Name, bento.Name, bentoVersion.Version)
	}
	return imageName, nil
}

func (s *bentoVersionService) GetS3BucketName(ctx context.Context, bentoVersion *models.BentoVersion) (string, error) {
	bento, err := BentoService.GetAssociatedBento(ctx, bentoVersion)
	if err != nil {
		return "", nil
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bento)
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

func (s *bentoVersionService) Update(ctx context.Context, bentoVersion *models.BentoVersion, opt UpdateBentoVersionOption) (*models.BentoVersion, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.ImageBuildStatus != nil {
		updaters["image_build_status"] = *opt.ImageBuildStatus
		defer func() {
			if err == nil {
				bentoVersion.ImageBuildStatus = *opt.ImageBuildStatus
			}
		}()
	}
	if opt.ImageBuildStatusSyncingAt != nil {
		updaters["image_build_status_syncing_at"] = *opt.ImageBuildStatusSyncingAt
		defer func() {
			if err == nil {
				bentoVersion.ImageBuildStatusSyncingAt = *opt.ImageBuildStatusSyncingAt
			}
		}()
	}
	if opt.ImageBuildStatusUpdatedAt != nil {
		updaters["image_build_status_updated_at"] = *opt.ImageBuildStatusUpdatedAt
		defer func() {
			if err == nil {
				bentoVersion.ImageBuildStatusUpdatedAt = *opt.ImageBuildStatusUpdatedAt
			}
		}()
	}
	if opt.UploadStatus != nil {
		updaters["upload_status"] = *opt.UploadStatus
		defer func() {
			if err == nil {
				bentoVersion.UploadStatus = *opt.UploadStatus
			}
		}()
	}
	if opt.UploadStartedAt != nil {
		updaters["upload_started_at"] = *opt.UploadStartedAt
		defer func() {
			if err == nil {
				bentoVersion.UploadStartedAt = *opt.UploadStartedAt
			}
		}()
	}
	if opt.UploadFinishedAt != nil {
		updaters["upload_finished_at"] = *opt.UploadFinishedAt
		defer func() {
			if err == nil {
				bentoVersion.UploadFinishedAt = *opt.UploadFinishedAt
			}
		}()
	}
	if opt.UploadFinishedReason != nil {
		updaters["upload_finished_reason"] = *opt.UploadFinishedReason
		defer func() {
			if err == nil {
				bentoVersion.UploadFinishedReason = *opt.UploadFinishedReason
			}
		}()
	}

	if len(updaters) == 0 {
		return bentoVersion, nil
	}

	// nolint: ineffassign,staticcheck
	db, ctx, df, err := startTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { df(err) }()

	err = db.Model(&models.BentoVersion{}).Where("id = ?", bentoVersion.ID).Updates(updaters).Error
	if err != nil {
		return nil, err
	}

	if opt.UploadStatus == nil || *opt.UploadStatus != modelschemas.BentoVersionUploadStatusSuccess {
		return bentoVersion, err
	}

	bento, err := BentoService.GetAssociatedBento(ctx, bentoVersion)
	if err != nil {
		return nil, err
	}

	org, err := OrganizationService.GetAssociatedOrganization(ctx, bento)
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

	kubeNamespace := consts.KubeNamespaceYataiBentoVersionImageBuilder

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

	kubeName, err := s.GetImageBuilderKubeName(ctx, bentoVersion)
	if err != nil {
		return nil, err
	}

	s3ObjectName, err := s.getS3ObjectName(ctx, bentoVersion)
	if err != nil {
		return nil, err
	}

	imageName, err := s.GetImageName(ctx, bentoVersion, true)
	if err != nil {
		return nil, err
	}

	s3BucketName, err := s.GetS3BucketName(ctx, bentoVersion)
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
					consts.KubeLabelYataiBento:        bento.Name,
					consts.KubeLabelYataiBentoVersion: bentoVersion.Version,
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
		_, _ = s.SyncImageBuilderStatus(ctx, bentoVersion)
	}()

	return bentoVersion, nil
}

func (s *bentoVersionService) GetImageBuilderKubeName(ctx context.Context, bentoVersion *models.BentoVersion) (string, error) {
	bento, err := BentoService.GetAssociatedBento(ctx, bentoVersion)
	if err != nil {
		return "", err
	}

	org, err := OrganizationService.GetAssociatedOrganization(ctx, bento)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(strcase.ToKebab(fmt.Sprintf("yatai-bento-image-builder-%s-%s-%s", org.Name, bento.Name, bentoVersion.Version)), ".", "-"), nil
}

func (s *bentoVersionService) Get(ctx context.Context, id uint) (*models.BentoVersion, error) {
	var bentoVersion models.BentoVersion
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&bentoVersion).Error
	if err != nil {
		return nil, err
	}
	if bentoVersion.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bentoVersion, nil
}

func (s *bentoVersionService) GetByUid(ctx context.Context, uid string) (*models.BentoVersion, error) {
	var bentoVersion models.BentoVersion
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&bentoVersion).Error
	if err != nil {
		return nil, err
	}
	if bentoVersion.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bentoVersion, nil
}

func (s *bentoVersionService) GetByVersion(ctx context.Context, bentoId uint, version string) (*models.BentoVersion, error) {
	var bentoVersion models.BentoVersion
	err := getBaseQuery(ctx, s).Where("bento_id = ?", bentoId).Where("version = ?", version).First(&bentoVersion).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get bento version %s", version)
	}
	if bentoVersion.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bentoVersion, nil
}

func (s *bentoVersionService) ListByUids(ctx context.Context, uids []string) ([]*models.BentoVersion, error) {
	bentoVersions := make([]*models.BentoVersion, 0, len(uids))
	if len(uids) == 0 {
		return bentoVersions, nil
	}
	err := getBaseQuery(ctx, s).Where("uid in (?)", uids).Find(&bentoVersions).Error
	return bentoVersions, err
}

func (s *bentoVersionService) ListModelVersionsFromManifests(ctx context.Context, bentoVersion *models.BentoVersion) (modelVersions []*models.ModelVersion, err error) {
	if bentoVersion.Manifest == nil {
		return
	}
	if bentoVersion.Manifest.Models == nil {
		return
	}
	modelNames := make([]string, 0, len(bentoVersion.Manifest.Models))
	modelNamesSeen := make(map[string]struct{}, len(bentoVersion.Manifest.Models))
	for _, modelTag := range bentoVersion.Manifest.Models {
		modelName, _, _ := xstrings.Partition(modelTag, ":")
		if _, ok := modelNamesSeen[modelName]; !ok {
			modelNames = append(modelNames, modelName)
		}
	}
	if len(modelNames) == 0 {
		return
	}
	bento, err := BentoService.GetAssociatedBento(ctx, bentoVersion)
	if err != nil {
		return
	}
	models_, _, err := ModelService.List(ctx, ListModelOption{
		OrganizationId: utils.UintPtr(bento.OrganizationId),
		Names:          &modelNames,
	})
	if len(models_) == 0 {
		return
	}
	modelNameMapping := make(map[string]*models.Model, len(models_))
	for _, model := range models_ {
		modelNameMapping[model.Name] = model
	}
	for _, modelTag := range bentoVersion.Manifest.Models {
		modelName, _, version := xstrings.Partition(modelTag, ":")
		if model, ok := modelNameMapping[modelName]; ok {
			var modelVersion *models.ModelVersion
			modelVersion, err = ModelVersionService.GetByVersion(ctx, model.ID, version)
			if err != nil {
				return
			}
			modelVersions = append(modelVersions, modelVersion)
		}
	}
	return
}

func (s *bentoVersionService) List(ctx context.Context, opt ListBentoVersionOption) ([]*models.BentoVersion, uint, error) {
	query := getBaseQuery(ctx, s)
	query = query.Joins("LEFT JOIN bento ON bento_version.bento_id = bento.id")
	if opt.ModelVersionIds != nil {
		query = query.Joins("LEFT JOIN bento_version_model_version_rel ON bento_version_model_version_rel.bento_version_id = bento_version.id").Where("bento_version_model_version_rel.model_version_id in (?)", *opt.ModelVersionIds)
	}
	if opt.OrganizationId != nil {
		query = query.Where("bento.organization_id = ?", *opt.OrganizationId)
	}
	if opt.BentoId != nil {
		query = query.Where("bento_version.bento_id = ?", *opt.BentoId)
	}
	if opt.Versions != nil {
		query = query.Where("bento_version.version in (?)", *opt.Versions)
	}
	if opt.CreatorId != nil {
		query = query.Where("bento_version.creator_id = ?", *opt.CreatorId)
	}
	if opt.Names != nil {
		query = query.Where("bento.name in (?)", *opt.Names)
	}
	if opt.CreatorIds != nil {
		query = query.Where("bento_version.creator_id in (?)", *opt.CreatorIds)
	}
	query = opt.BindQueryWithKeywords(query, "bento")
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	bentoVersions := make([]*models.BentoVersion, 0)
	query = opt.BindQueryWithLimit(query)
	if opt.Order != nil {
		query = query.Order(*opt.Order)
	} else {
		query = query.Order("bento_version.build_at DESC")
	}
	err = query.Select("bento_version.*").Find(&bentoVersions).Error
	if err != nil {
		return nil, 0, err
	}
	return bentoVersions, uint(total), err
}

func (s *bentoVersionService) ListLatestByBentoIds(ctx context.Context, bentoIds []uint) ([]*models.BentoVersion, error) {
	db := mustGetSession(ctx)

	query := db.Raw(`select * from bento_version where id in (
					select n.version_id from (
						select bento_id, max(id) as version_id from bento_version
						where bento_id in (?) group by bento_id
					) as n)`, bentoIds)

	versions := make([]*models.BentoVersion, 0, len(bentoIds))
	err := query.Find(&versions).Error
	if err != nil {
		return nil, err
	}

	return versions, err
}

func (s *bentoVersionService) ListImageBuilderPods(ctx context.Context, bentoVersion *models.BentoVersion) ([]*models.KubePodWithStatus, error) {
	bento, err := BentoService.GetAssociatedBento(ctx, bentoVersion)
	if err != nil {
		return nil, err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, bento)
	if err != nil {
		return nil, err
	}
	cluster, err := OrganizationService.GetMajorCluster(ctx, org)
	if err != nil {
		return nil, err
	}
	_, podLister, err := GetPodInformer(ctx, cluster, consts.KubeNamespaceYataiBentoVersionImageBuilder)
	if err != nil {
		return nil, err
	}

	selector, err := labels.Parse(fmt.Sprintf("%s = %s, %s = %s", consts.KubeLabelYataiBento, bento.Name, consts.KubeLabelYataiBentoVersion, bentoVersion.Version))
	if err != nil {
		return nil, err
	}

	pods, err := podLister.List(selector)
	if err != nil {
		return nil, err
	}

	_, eventLister, err := GetEventInformer(ctx, cluster, consts.KubeNamespaceYataiBentoVersionImageBuilder)
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

func (s *bentoVersionService) CalculateImageBuildStatus(ctx context.Context, bentoVersion *models.BentoVersion) (modelschemas.BentoVersionImageBuildStatus, error) {
	defaultStatus := modelschemas.BentoVersionImageBuildStatusPending
	pods, err := s.ListImageBuilderPods(ctx, bentoVersion)
	if err != nil {
		return defaultStatus, err
	}

	if len(pods) == 0 {
		return defaultStatus, nil
	}

	for _, p := range pods {
		if p.Status.Status == modelschemas.KubePodActualStatusRunning || p.Status.Status == modelschemas.KubePodActualStatusPending {
			return modelschemas.BentoVersionImageBuildStatusBuilding, nil
		}
		if p.Status.Status == modelschemas.KubePodActualStatusTerminating || p.Status.Status == modelschemas.KubePodActualStatusUnknown || p.Status.Status == modelschemas.KubePodActualStatusFailed {
			return modelschemas.BentoVersionImageBuildStatusFailed, nil
		}
	}

	hasPending := false
	hasFailed := false
	hasBuilding := false

	modelVersions, err := s.ListModelVersionsFromManifests(ctx, bentoVersion)
	if err != nil {
		return defaultStatus, err
	}

	for _, modelVersion := range modelVersions {
		if modelVersion.ImageBuildStatus == modelschemas.ModelVersionImageBuildStatusPending {
			hasPending = true
		}
		if modelVersion.ImageBuildStatus == modelschemas.ModelVersionImageBuildStatusFailed {
			hasFailed = true
		}
		if modelVersion.ImageBuildStatus == modelschemas.ModelVersionImageBuildStatusBuilding {
			hasBuilding = true
		}
	}

	if hasFailed {
		return modelschemas.BentoVersionImageBuildStatusFailed, nil
	}

	if hasBuilding || hasPending {
		return modelschemas.BentoVersionImageBuildStatusBuilding, nil
	}

	return modelschemas.BentoVersionImageBuildStatusSuccess, nil
}

func (s *bentoVersionService) ListImageBuildStatusUnsynced(ctx context.Context) ([]*models.BentoVersion, error) {
	q := getBaseQuery(ctx, s)
	now := time.Now()
	t := now.Add(-time.Minute)
	q = q.Where("image_build_status != ? and (image_build_status_syncing_at is null or image_build_status_syncing_at < ? or image_build_status_updated_at is null or image_build_status_updated_at < ?)", modelschemas.BentoVersionImageBuildStatusSuccess, t, t)
	bentoVersions := make([]*models.BentoVersion, 0)
	err := q.Order("id DESC").Find(&bentoVersions).Error
	return bentoVersions, err
}

func (s *bentoVersionService) SyncImageBuilderStatus(ctx context.Context, bentoVersion *models.BentoVersion) (modelschemas.BentoVersionImageBuildStatus, error) {
	now := time.Now()
	nowPtr := &now
	_, err := s.Update(ctx, bentoVersion, UpdateBentoVersionOption{
		ImageBuildStatusSyncingAt: &nowPtr,
	})
	if err != nil {
		return bentoVersion.ImageBuildStatus, err
	}
	currentStatus, err := s.CalculateImageBuildStatus(ctx, bentoVersion)
	if err != nil {
		return bentoVersion.ImageBuildStatus, err
	}
	now = time.Now()
	nowPtr = &now
	_, err = s.Update(ctx, bentoVersion, UpdateBentoVersionOption{
		ImageBuildStatus:          &currentStatus,
		ImageBuildStatusUpdatedAt: &nowPtr,
	})
	if err != nil {
		return currentStatus, err
	}
	return currentStatus, nil
}

type IBentoVersionAssociate interface {
	GetAssociatedBentoVersionId() uint
	GetAssociatedBentoVersionCache() *models.BentoVersion
	SetAssociatedBentoVersionCache(version *models.BentoVersion)
}

func (s *bentoVersionService) GetAssociatedBentoVersion(ctx context.Context, associate IBentoVersionAssociate) (*models.BentoVersion, error) {
	cache := associate.GetAssociatedBentoVersionCache()
	if cache != nil {
		return cache, nil
	}
	version, err := s.Get(ctx, associate.GetAssociatedBentoVersionId())
	associate.SetAssociatedBentoVersionCache(version)
	return version, err
}

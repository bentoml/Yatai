package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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
}

type UpdateModelVersionOption struct {
	ImageBuildStatus          *modelschemas.ModelVersionImageBuildStatus
	ImageBuildStatusSyncingAt **time.Time
	ImageBuildStatusUpdatedAt **time.Time
	UploadStatus              *modelschemas.ModelVersionUploadStatus
	UploadStartedAt           **time.Time
	UploadFinishedAt          **time.Time
	UploadFinishedReason      *string
}

type ListModelVersionOption struct {
	BaseListOption
	ModelId *uint
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
	if org.Config == nil {
		err = errors.New("This organization does not have configuration")
		return
	}
	if org.Config.AWS == nil || org.Config.AWS.S3 == nil {
		err = errors.New("This organization does not have aws s3 storage set up")
		return
	}
	minioConf := org.Config.AWS.S3
	minioClient, err := minio.New("s3.amazonaws.com", &minio.Options{
		Creds:  credentials.NewStaticV4(org.Config.AWS.AccessKeyId, org.Config.AWS.SecretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		err = errors.Wrap(err, "create s3 client")
		return
	}

	bucketName := minioConf.BucketName

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: minioConf.Region})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists != nil || !exists {
			err = errors.Wrapf(err, "create bucket %s", bucketName)
			return
		}
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

func (s *modelVersionService) GetImageName(ctx context.Context, modelVersion *models.ModelVersion) (string, error) {
	model, err := ModelService.GetAssociatedModel(ctx, modelVersion)
	if err != nil {
		return "", err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, model)
	if err != nil {
		return "", err
	}
	if org.Config == nil || org.Config.AWS == nil || org.Config.AWS.ECR == nil {
		return "", errors.Errorf("Organization %s does not have ECR configuration", org.Name)
	}

	imageName := fmt.Sprintf("%s:yatai.%s.%s.%s", org.Config.AWS.ECR.RepositoryURI, org.Name, model.Name, modelVersion.Version)
	return imageName, nil
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

	if opt.UploadStatus != nil || *opt.UploadStatus != modelschemas.ModelVersionUploadStatusSuccess {
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
	if org.Config == nil || org.Config.AWS == nil {
		return nil, errors.Errorf("Organization %s does not have AWS configuration", org.Name)
	}

	awsSecretKubeName := "aws-secret"
	var awsSecretBuffer bytes.Buffer
	t := template.Must(template.New(awsSecretKubeName).Parse(awsSecretTemplate))
	if err := t.Execute(&awsSecretBuffer, map[string]string{
		"AccessKeyId":     org.Config.AWS.AccessKeyId,
		"SecretAccessKey": org.Config.AWS.SecretAccessKey,
	}); err != nil {
		return nil, err
	}

	secretsCli := kubeCli.CoreV1().Secrets(kubeNamespace)
	_, err = secretsCli.Get(ctx, awsSecretKubeName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = secretsCli.Create(ctx, &apiv1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: awsSecretKubeName},
			StringData: map[string]string{
				"credentials": awsSecretBuffer.String(),
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	dockerCMKubeName := "docker-config"
	dockerCMContent, err := json.Marshal(struct {
		CredsStore string `json:"creds_store"`
	}{
		CredsStore: "ecr-login",
	})
	if err != nil {
		return nil, err
	}
	cmsCli := kubeCli.CoreV1().ConfigMaps(kubeNamespace)
	_, err = cmsCli.Get(ctx, dockerCMKubeName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = cmsCli.Create(ctx, &apiv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: dockerCMKubeName},
			Data: map[string]string{
				"config.json": string(dockerCMContent),
			},
		}, metav1.CreateOptions{})
	} else if err != nil {
		return nil, err
	}

	podsCli := kubeCli.CoreV1().Pods(kubeNamespace)

	kubeName, err := s.GetImageBuilderKubeName(ctx, modelVersion)
	if err != nil {
		return nil, err
	}
	if org.Config == nil || org.Config.AWS == nil || org.Config.AWS.S3 == nil {
		return nil, errors.Errorf("Organization %s does not have AWS S3 configuration", org.Name)
	}

	s3ObjectName, err := s.getS3ObjectName(ctx, modelVersion)
	if err != nil {
		return nil, err
	}

	imageName, err := s.GetImageName(ctx, modelVersion)
	if err != nil {
		return nil, err
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
					{
						Name: awsSecretKubeName,
						VolumeSource: apiv1.VolumeSource{
							Secret: &apiv1.SecretVolumeSource{
								SecretName: awsSecretKubeName,
							},
						},
					},
				},
				Containers: []apiv1.Container{
					{
						Name:  "image-builder",
						Image: "gcr.io/kaniko-project/executor:latest",
						Args: []string{
							"--dockerfile=./Dockerfile",
							fmt.Sprintf("--context=s3://%s/%s", org.Config.AWS.S3.BucketName, s3ObjectName),
							fmt.Sprintf("--destination=%s", imageName),
						},
						VolumeMounts: []apiv1.VolumeMount{
							{
								Name:      dockerCMKubeName,
								MountPath: "/kaniko/.docker/",
							},
							{
								Name:      awsSecretKubeName,
								MountPath: "/root/.aws/",
							},
						},
						Env: []apiv1.EnvVar{
							{
								Name:  "AWS_REGION",
								Value: org.Config.AWS.S3.Region,
							},
						},
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

	return strings.ReplaceAll(strcase.ToKebab(fmt.Sprintf("yatai-image-builder-%s-%s-%s", org.Name, model.Name, modelVersion.Version)), ".", "-"), nil
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
	if opt.ModelId != nil {
		query = query.Where("model_id = ?", *opt.ModelId)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	modelVersions := make([]*models.ModelVersion, 0)
	query = opt.BindQueryWithLimit(query).Order("build_at DESC")
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
			return modelschemas.ModelversionImageBuildStatusFailed, nil
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

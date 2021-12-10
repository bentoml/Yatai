package services

import (
	"context"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/schemas/modelschemas"
)

type imageBuilderService struct{}

var ImageBuilderService = &imageBuilderService{}

type CreateImageBuilderJobOption struct {
	KubeName             string
	ImageName            string
	S3ObjectName         string
	S3BucketName         string
	Cluster              *models.Cluster
	DockerFileCMKubeName *string
	DockerFileContent    *string
	DockerFilePath       *string
	KubeLabels           map[string]string
}

func (s *imageBuilderService) CreateImageBuilderJob(ctx context.Context, opt CreateImageBuilderJobOption) (err error) {
	kubeCli, _, err := ClusterService.GetKubeCliSet(ctx, opt.Cluster)
	if err != nil {
		return
	}

	kubeNamespace := consts.KubeNamespaceYataiModelImageBuilder

	_, err = KubeNamespaceService.MakeSureNamespace(ctx, opt.Cluster, kubeNamespace)
	if err != nil {
		return
	}

	dockerConfigCM, err := ClusterService.MakeSureDockerConfigCM(ctx, opt.Cluster, kubeNamespace)
	if err != nil {
		return
	}
	dockerConfigCMKubeName := dockerConfigCM.Name

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
	}

	volumeMounts := []apiv1.VolumeMount{
		{
			Name:      dockerConfigCMKubeName,
			MountPath: "/kaniko/.docker/",
		},
	}

	dockerFilePath := ""
	if opt.DockerFilePath != nil {
		dockerFilePath = *opt.DockerFilePath
	}

	cmsCli := kubeCli.CoreV1().ConfigMaps(kubeNamespace)
	if opt.DockerFileCMKubeName != nil && opt.DockerFileContent != nil {
		var oldDockerFileCM *apiv1.ConfigMap
		oldDockerFileCM, err = cmsCli.Get(ctx, *opt.DockerFileCMKubeName, metav1.GetOptions{})
		// nolint: gocritic
		if apierrors.IsNotFound(err) {
			_, err = cmsCli.Create(ctx, &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: *opt.DockerFileCMKubeName},
				Data: map[string]string{
					"Dockerfile": *opt.DockerFileContent,
				},
			}, metav1.CreateOptions{})
			if err != nil {
				return
			}
		} else if err != nil {
			return
		} else {
			oldDockerFileCM.Data["Dockerfile"] = *opt.DockerFileContent
			_, err = cmsCli.Update(ctx, oldDockerFileCM, metav1.UpdateOptions{})
			if err != nil {
				return
			}
		}
		volumes = append(volumes,
			apiv1.Volume{
				Name: *opt.DockerFileCMKubeName,
				VolumeSource: apiv1.VolumeSource{
					ConfigMap: &apiv1.ConfigMapVolumeSource{
						LocalObjectReference: apiv1.LocalObjectReference{
							Name: *opt.DockerFileCMKubeName,
						},
					},
				},
			},
		)
		volumeMounts = append(volumeMounts,
			apiv1.VolumeMount{
				Name:      *opt.DockerFileCMKubeName,
				MountPath: "/docker/",
			},
		)
		dockerFilePath = "/docker/Dockerfile"
	}

	org, err := OrganizationService.GetAssociatedOrganization(ctx, opt.Cluster)
	if err != nil {
		return
	}

	s3Config, err := OrganizationService.GetS3Config(ctx, org)
	if err != nil {
		return
	}

	dockerRegistry, err := OrganizationService.GetDockerRegistry(ctx, org)
	if err != nil {
		return
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

	kubeName := opt.KubeName
	s3ObjectName := opt.S3ObjectName
	imageName := opt.ImageName
	s3BucketName := opt.S3BucketName

	err = s3Config.MakeSureBucket(ctx, s3BucketName)
	if err != nil {
		return
	}

	args := []string{
		fmt.Sprintf("--dockerfile=%s", dockerFilePath),
		fmt.Sprintf("--context=s3://%s/%s", s3BucketName, s3ObjectName),
		fmt.Sprintf("--destination=%s", imageName),
	}

	if !dockerRegistry.Secure {
		args = append(args, "--insecure")
	}

	podsCli := kubeCli.CoreV1().Pods(kubeNamespace)

	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   kubeName,
			Labels: opt.KubeLabels,
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
	}

	selectorPieces := make([]string, 0, len(opt.KubeLabels))
	for k, v := range opt.KubeLabels {
		selectorPieces = append(selectorPieces, fmt.Sprintf("%s = %s", k, v))
	}

	if len(selectorPieces) > 0 {
		var pods *apiv1.PodList
		pods, err = podsCli.List(ctx, metav1.ListOptions{
			LabelSelector: strings.Join(selectorPieces, ", "),
		})
		if err != nil {
			return
		}
		for _, pod_ := range pods.Items {
			err = podsCli.Delete(ctx, pod_.Name, metav1.DeleteOptions{})
			if err != nil {
				return
			}
		}
	}

	oldPod, err := podsCli.Get(ctx, kubeName, metav1.GetOptions{})
	isNotFound := apierrors.IsNotFound(err)
	if !isNotFound && err != nil {
		return
	}
	if isNotFound {
		_, err = podsCli.Create(ctx, pod, metav1.CreateOptions{})
		if err != nil {
			return
		}
	} else {
		oldPod.Spec = pod.Spec
		_, err = podsCli.Update(ctx, oldPod, metav1.UpdateOptions{})
		if err != nil {
			return
		}
	}

	return nil
}

func (s *imageBuilderService) ListImageBuilderPods(ctx context.Context, cluster *models.Cluster, kubeLabels map[string]string) ([]*models.KubePodWithStatus, error) {
	_, podLister, err := GetPodInformer(ctx, cluster, consts.KubeNamespaceYataiModelImageBuilder)
	if err != nil {
		return nil, err
	}

	selectorPieces := make([]string, 0, len(kubeLabels))
	for k, v := range kubeLabels {
		selectorPieces = append(selectorPieces, fmt.Sprintf("%s = %s", k, v))
	}

	selector, err := labels.Parse(strings.Join(selectorPieces, ", "))
	if err != nil {
		return nil, err
	}
	pods, err := podLister.List(selector)
	if err != nil {
		return nil, err
	}
	_, eventLister, err := GetEventInformer(ctx, cluster, consts.KubeNamespaceYataiModelImageBuilder)
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

	pods__ := KubePodService.MapKubePodsToKubePodWithStatuses(ctx, pods_, events_)

	res := make([]*models.KubePodWithStatus, 0)
	for _, pod := range pods__ {
		if pod.Status.Status == modelschemas.KubePodActualStatusTerminating {
			continue
		}
		res = append(res, pod)
	}

	return res, nil
}

func (s *imageBuilderService) CalculateImageBuildStatus(pods []*models.KubePodWithStatus) (modelschemas.ImageBuildStatus, error) {
	defaultStatus := modelschemas.ImageBuildStatusPending

	if len(pods) == 0 {
		return defaultStatus, nil
	}

	for _, p := range pods {
		if p.Status.Status == modelschemas.KubePodActualStatusRunning || p.Status.Status == modelschemas.KubePodActualStatusPending {
			return modelschemas.ImageBuildStatusBuilding, nil
		}
		if p.Status.Status == modelschemas.KubePodActualStatusUnknown || p.Status.Status == modelschemas.KubePodActualStatusFailed {
			return modelschemas.ImageBuildStatusFailed, nil
		}
	}

	return modelschemas.ImageBuildStatusSuccess, nil
}

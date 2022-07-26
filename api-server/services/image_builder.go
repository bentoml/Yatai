package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/config"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/utils"
)

type imageBuilderService struct{}

var ImageBuilderService = &imageBuilderService{}

type CreateImageBuilderJobOption struct {
	KubeName          string
	ImageName         string
	S3ObjectName      string
	S3BucketName      string
	Cluster           *models.Cluster
	DockerFileContent *string
	DockerFilePath    *string
	KubeLabels        map[string]string
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

	dockerConfigSecret, err := ClusterService.MakeSureDockerConfigSecret(ctx, opt.Cluster, kubeNamespace)
	if err != nil {
		return
	}
	dockerConfigSecretKubeName := dockerConfigSecret.Name

	volumes := []apiv1.Volume{
		{
			Name: "yatai",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "workspace",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: dockerConfigSecretKubeName,
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName: dockerConfigSecretKubeName,
				},
			},
		},
	}

	volumeMounts := []apiv1.VolumeMount{
		{
			Name:      "yatai",
			MountPath: "/yatai",
		},
		{
			Name:      "workspace",
			MountPath: "/workspace",
		},
		{
			Name:      dockerConfigSecretKubeName,
			MountPath: "/kaniko/.docker/",
		},
	}

	org, err := OrganizationService.GetAssociatedOrganization(ctx, opt.Cluster)
	if err != nil {
		return
	}

	dockerRegistry, err := OrganizationService.GetDockerRegistry(ctx, org)
	if err != nil {
		return
	}

	s3Config, err := OrganizationService.GetS3Config(ctx, org)
	if err != nil {
		return
	}

	// nolint: goconst
	s3ForcePath := "true"
	if s3Config.Endpoint == consts.AmazonS3Endpoint {
		// nolint: goconst
		s3ForcePath = "false"
	}

	kubeName := opt.KubeName
	s3ObjectName := opt.S3ObjectName
	imageName := opt.ImageName
	s3BucketName := opt.S3BucketName

	err = s3Config.MakeSureBucket(ctx, s3BucketName)
	if err != nil {
		return
	}

	privileged := false
	if config.YataiConfig.DockerImageBuilder != nil && config.YataiConfig.DockerImageBuilder.Privileged {
		privileged = true
	}

	securityContext := &apiv1.SecurityContext{
		RunAsUser:  utils.Int64Ptr(1000),
		RunAsGroup: utils.Int64Ptr(1000),
	}

	if privileged {
		securityContext = nil
	}

	s3DownloaderCommand := "s3-downloader && rm /workspace/buildcontext/context.tar.gz"
	if !privileged {
		s3DownloaderCommand += " && chown -R 1000:1000 /workspace"
	}

	initContainers := []apiv1.Container{
		{
			Name:  "s3-downloader",
			Image: "quay.io/bentoml/s3-downloader:0.0.1",
			Command: []string{
				"sh",
				"-c",
				s3DownloaderCommand,
			},
			VolumeMounts: volumeMounts,
			Env: []apiv1.EnvVar{
				{
					Name:  "KANIKO_DIR",
					Value: "/workspace",
				},
				{
					Name:  "BUILD_CONTEXT",
					Value: fmt.Sprintf("s3://%s/%s", s3BucketName, s3ObjectName),
				},
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
			},
		},
	}

	dockerFilePath := ""
	if opt.DockerFilePath != nil {
		dockerFilePath = filepath.Join("/workspace/buildcontext", *opt.DockerFilePath)
	}
	if opt.DockerFileContent != nil {
		dockerFilePath = "/yatai/Dockerfile"
		initContainers = append(initContainers, apiv1.Container{
			Name:  "init-dockerfile",
			Image: "quay.io/bentoml/busybox:1.33",
			Command: []string{
				"sh",
				"-c",
				fmt.Sprintf("echo \"%s\" > %s", *opt.DockerFileContent, dockerFilePath),
			},
			VolumeMounts:    volumeMounts,
			SecurityContext: securityContext,
		})
	}

	envs := []apiv1.EnvVar{
		{
			Name:  "DOCKER_CONFIG",
			Value: "/kaniko/.docker/",
		},
	}

	if !privileged {
		envs = append(envs, apiv1.EnvVar{
			Name:  "BUILDKITD_FLAGS",
			Value: "--oci-worker-no-process-sandbox",
		})
	}

	args := []string{
		"build",
		"--frontend",
		"dockerfile.v0",
		"--local",
		"context=/workspace/buildcontext",
		"--local",
		fmt.Sprintf("dockerfile=%s", filepath.Dir(dockerFilePath)),
		"--output",
		fmt.Sprintf("type=image,name=%s,push=true,registry.insecure=%v", imageName, !dockerRegistry.Secure),
	}

	// dockerRegistry, err := OrganizationService.GetDockerRegistry(ctx, org)
	// if err != nil {
	// 	return
	// }

	// if !dockerRegistry.Secure {
	// 	args = append(args, "--insecure")
	// }

	annotations := make(map[string]string, 1)
	if !privileged {
		annotations["container.apparmor.security.beta.kubernetes.io/builder"] = "unconfined"
	}

	image := "quay.io/bentoml/buildkit:master-rootless"
	if privileged {
		image = "quay.io/bentoml/buildkit:master"
	}

	securityContext_ := &apiv1.SecurityContext{
		SeccompProfile: &apiv1.SeccompProfile{
			Type: apiv1.SeccompProfileTypeUnconfined,
		},
		RunAsUser:  utils.Int64Ptr(1000),
		RunAsGroup: utils.Int64Ptr(1000),
	}
	if privileged {
		securityContext_ = &apiv1.SecurityContext{
			Privileged: utils.BoolPtr(true),
		}
	}

	podsCli := kubeCli.CoreV1().Pods(kubeNamespace)

	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        kubeName,
			Labels:      opt.KubeLabels,
			Annotations: annotations,
		},
		Spec: apiv1.PodSpec{
			RestartPolicy:  apiv1.RestartPolicyNever,
			Volumes:        volumes,
			InitContainers: initContainers,
			Containers: []apiv1.Container{
				{
					Name:            "builder",
					Image:           image,
					ImagePullPolicy: apiv1.PullAlways,
					Command:         []string{"buildctl-daemonless.sh"},
					Args:            args,
					VolumeMounts:    volumeMounts,
					Env:             envs,
					TTY:             true,
					Stdin:           true,
					SecurityContext: securityContext_,
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

package consts

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	KubeIngressBaseHost = "mcd.megvii-inc.com"

	KubeLabelYataiSelector        = "yatai.io/selector"
	KubeLabelMcdInfraCli          = "mcd-infra-cli"
	KubeLabelMcdKubectl           = "mcd-kubectl"
	KubeLabelMcdUser              = "mcd-user"
	KubeLabelMcdApp               = "mcd-app"
	KubeLabelMcdAppPool           = "mcd-app-pool"
	KubeLabelYataiDeploymentId    = "yatai.io/deployment-id"
	KubeLabelMcdDeploymentFixedId = "mcd-deployment-fixed-id"
	KubeLabelMcdDeploymentType    = "mcd-deployment-type"
	KubeLabelMcdDeploymentEnvType = "mcd-deployment-env-type"
	KubeLabelCreator              = "creator"
	// nolint: gosec
	KubeLabelMcdDeployToken = "mcd-deploy-token"

	KubeLabelMcdAppCompType = "mcd-app-comp-type"
	KubeLabelMcdAppCompName = "mcd-app-comp-name"

	KubeLabelMcdOwnerReference = "mcd-owner-reference"

	KubeLabelGPUAccelerator = "gpu-accelerator"

	KubeLabelHostName = "kubernetes.io/hostname"
	KubeLabelArch     = "kubernetes.io/arch"

	KubeLabelMcdNodePool       = "mcd.io/node-pool"
	KubeLabelAlibabaEdgeWorker = "alibabacloud.com/is-edge-worker"
	KubeLabelMcdEdgeWorker     = "mcd.io/is-edge-worker"
	KubeLabelFalse             = "false"
	KubeLabelTrue              = "true"

	KubeLabelManagedBy    = "app.kubernetes.io/managed-by"
	KubeLabelHelmHeritage = "heritage"
	KubeLabelHelmRelease  = "release"

	KubeAnnotationMcdCommitId     = "mcd-commit-id"
	KubeAnnotationMcdDeploymentId = "mcd-deployment-id"
	KubeAnnotationHelmReleaseName = "meta.helm.sh/release-name"

	KubeAnnotationPrometheusScrape = "prometheus.io/scrape"
	KubeAnnotationPrometheusPort   = "prometheus.io/port"
	KubeAnnotationPrometheusPath   = "prometheus.io/path"

	KubeAnnotationARMSAutoEnable = "armsPilotAutoEnable"
	KubeAnnotationARMSAppName    = "armsPilotCreateAppName"

	KubeCreator = "mcd"

	KubeVolumeNamePermdir                            = "permdir"
	KubeVolumeNameFastPermdir                        = "fast-permdir"
	KubeVolumeNameHostTimezone                       = "host-timezone"
	KubeVolumeNameMcdTracingAgentDir                 = "mcd-tracing"
	KubeVolumeNameMcdJmxAgentDir                     = "mcd-jmx"
	KubeVolumeMountPathPermdir                       = "/permdir"
	KubeVolumeMountPathFastPermdir                   = "/fast_permdir"
	KubeVolumeNameDockerSock                         = "mcd-docker-sock"
	KubeVolumeMountPathDockerSock                    = "/var/run/docker.sock"
	KubeVolumeNameDockerGraphStorage                 = "mcd-docker-graph-storage"
	KubeVolumeMountPathDockerGraphStorage            = "/var/lib/docker"
	KubeVolumeNameVarRun                             = "mcd-var-run"
	KubeVolumeMountPathVarRun                        = "/var/run"
	KubePersistentVolumeClaimNamePermdir             = "mcd-app-permdir"
	KubePersistentVolumeClaimNameFastPermdir         = "mcd-app-fast-permdir"
	KubePersistentVolumeClaimPermdirStorageClass     = "mcd-nfs"
	KubePersistentVolumeClaimFastPermdirStorageClass = "mcd-fast-nfs"
	KubeAliCouldStorageClassProvisioner              = "nasplugin.csi.alibabacloud.com"

	KubeIngressCanaryHeader      = "mcd-canary"
	KubeIngressCanaryHeaderValue = "always"

	KubeNameMcdDns                      = "mcd-dns"
	KubeNamespaceYataiDeploymentDefault = "yatai-deployment"

	KubeStorageClassNameMcd       = "mcd"
	KubeStorageClassNameLocalPath = "local-path"

	KubeResourceGPUNvidia = "nvidia.com/gpu"

	KubeEventResourceKindPod        = "Pod"
	KubeEventResourceKindHPA        = "HorizontalPodAutoscaler"
	KubeEventResourceKindReplicaSet = "ReplicaSet"

	KubeTaintKeyDedicatedNodeGroup = "mcd.io/dedicated-node-group"
	KubeLabelDedicatedNodeGroup    = "mcd.io/dedicated-node-group"

	KubeLabelMcdESEnable  = "mcd-es-enable"
	KubeLabelMcdESSaveDay = "mcd-es-save-day"

	KubeCSIDriverImage = "image.csi.k8s.io"

	KubeDefaultMcdResourceQuotaName = "mcd"

	KubeLabelNodeResourceResizeCPU    = "mcd.io/resize-node-cpu"
	KubeLabelNodeResourceResizePods   = "mcd.io/resize-node-pods"
	KubeLabelNodeResourceResizeMemory = "mcd.io/resize-node-memory"
)

var KubeListEverything = metav1.ListOptions{
	LabelSelector: labels.Everything().String(),
	FieldSelector: fields.Everything().String(),
}

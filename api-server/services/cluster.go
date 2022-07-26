package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"

	"github.com/bentoml/grafana-operator/api/integreatly/v1alpha1"

	"github.com/bentoml/yatai-common/system"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/config"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/helmchart"
	"github.com/bentoml/yatai/common/utils"
)

const (
	// High enough QPS to fit all expected use cases.
	defaultQPS = 1e6
	// High enough Burst to fit all expected use cases.
	defaultBurst = 1e6
)

type clusterService struct{}

var ClusterService = clusterService{}

func (*clusterService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Cluster{})
}

type CreateClusterOption struct {
	CreatorId      uint
	OrganizationId uint
	Name           string
	Description    string
	KubeConfig     string
	Config         *modelschemas.ClusterConfigSchema
}

type UpdateClusterOption struct {
	Description *string
	Config      **modelschemas.ClusterConfigSchema
	KubeConfig  *string
}

type ListClusterOption struct {
	BaseListOption
	VisitorId      *uint
	OrganizationId *uint
	Ids            *[]uint
	Names          *[]string
	CreatorIds     *[]uint
	Order          *string
}

func (s *clusterService) Create(ctx context.Context, opt CreateClusterOption) (*models.Cluster, error) {
	errs := validation.IsDNS1035Label(opt.Name)
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ";"))
	}

	// nolint: ineffassign,staticcheck
	db, ctx, df, err := startTransaction(ctx)
	if err != nil {
		return nil, err
	}

	defer func() { df(err) }()
	cluster := models.Cluster{
		ResourceMixin: models.ResourceMixin{
			Name: opt.Name,
		},
		Description: opt.Description,
		KubeConfig:  opt.KubeConfig,
		Config:      opt.Config,
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		OrganizationAssociate: models.OrganizationAssociate{
			OrganizationId: opt.OrganizationId,
		},
	}
	err = db.Create(&cluster).Error
	if err != nil {
		return nil, err
	}

	return &cluster, err
}

func (s *clusterService) Update(ctx context.Context, c *models.Cluster, opt UpdateClusterOption) (*models.Cluster, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.Config != nil {
		updaters["config"] = *opt.Config
		defer func() {
			if err == nil {
				c.Config = *opt.Config
			}
		}()
	}
	if opt.Description != nil {
		updaters["description"] = *opt.Description
		defer func() {
			if err == nil {
				c.Description = *opt.Description
			}
		}()
	}
	if opt.KubeConfig != nil {
		updaters["kube_config"] = *opt.KubeConfig
		defer func() {
			if err == nil {
				c.KubeConfig = *opt.KubeConfig
			}
		}()
	}

	if len(updaters) == 0 {
		return c, nil
	}

	err = s.getBaseDB(ctx).Where("id = ?", c.ID).Updates(updaters).Error
	if err != nil {
		return nil, err
	}

	return c, err
}

func (s *clusterService) Get(ctx context.Context, id uint) (*models.Cluster, error) {
	var cluster models.Cluster
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	if cluster.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &cluster, nil
}

func (s *clusterService) GetByUid(ctx context.Context, uid string) (*models.Cluster, error) {
	var cluster models.Cluster
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	if cluster.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &cluster, nil
}

func (s *clusterService) GetByName(ctx context.Context, organizationId uint, name string) (*models.Cluster, error) {
	var cluster models.Cluster
	err := getBaseQuery(ctx, s).Where("organization_id = ?", organizationId).Where("name = ?", name).First(&cluster).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get cluster %s", name)
	}
	if cluster.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &cluster, nil
}

func (s *clusterService) GetIdByName(ctx context.Context, organizationId uint, name string) (uint, error) {
	var cluster models.Cluster
	err := mustGetSession(ctx).Select("id").Where("organization_id = ?", organizationId).Where("name = ?", name).First(&cluster).Error
	return cluster.ID, err
}

func (s *clusterService) List(ctx context.Context, opt ListClusterOption) ([]*models.Cluster, uint, error) {
	clusters := make([]*models.Cluster, 0)
	query := getBaseQuery(ctx, s)
	if opt.VisitorId != nil {
		userID := opt.VisitorId
		user, err := UserService.Get(ctx, *userID)
		if err != nil {
			return nil, 0, errors.Wrapf(err, "get user %d", userID)
		}
		if !UserService.IsAdmin(ctx, user, nil) {
			clusterMembers, err := ClusterMemberService.List(ctx, ListClusterMemberOption{UserId: userID})
			if err != nil {
				return nil, 0, err
			}
			clusterIds := make([]uint, 0, len(clusterMembers))
			for _, member := range clusterMembers {
				clusterIds = append(clusterIds, member.ClusterId)
			}
			clusterIds = append(clusterIds, 0) // Add a fill value of 0 because it cannot be empty
			query = query.Where("(id in (?) OR creator_id = ?)", clusterIds, userID)
		}
	}
	if opt.OrganizationId != nil {
		query = query.Where("organization_id = ?", *opt.OrganizationId)
	}
	if opt.Ids != nil {
		if len(*opt.Ids) == 0 {
			return clusters, 0, nil
		}
		query = query.Where("id in (?)", *opt.Ids)
	}
	if opt.Names != nil {
		if len(*opt.Names) == 0 {
			return clusters, 0, nil
		}
		query = query.Where("name in (?)", *opt.Names)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	query = opt.BindQueryWithLimit(query)
	if opt.Ids == nil {
		if opt.Order == nil {
			query = query.Order("id DESC")
		} else {
			query = query.Order(*opt.Order)
		}
	}
	err = query.Find(&clusters).Error
	if err != nil {
		return nil, 0, err
	}
	return clusters, uint(total), err
}

func (s *clusterService) GetRESTClientGetter(ctx context.Context, c *models.Cluster, namespace string) (genericclioptions.RESTClientGetter, error) {
	_, restConfig, err := s.GetKubeCliSet(ctx, c)
	if err != nil {
		return nil, errors.Wrap(err, "get kube cli set")
	}
	return helmchart.NewRESTClientGetter(namespace, nil, &restConfig), nil
}

func (s *clusterService) GetDeploymentKubeNamespace(c *models.Cluster) string {
	defaultKubeNamespace := consts.KubeNamespaceYataiDeployment
	if c.Config == nil {
		return defaultKubeNamespace
	}
	kubeNamespace := strings.TrimSpace(c.Config.DefaultDeploymentKubeNamespace)
	if kubeNamespace == "" {
		return defaultKubeNamespace
	}
	return kubeNamespace
}

func (s *clusterService) GetDefault(ctx context.Context, orgId uint) (defaultCluster *models.Cluster, err error) {
	clusters, total, err := s.List(ctx, ListClusterOption{
		BaseListOption: BaseListOption{
			Start: utils.UintPtr(0),
			Count: utils.UintPtr(1),
		},
		OrganizationId: utils.UintPtr(orgId),
		Order:          utils.StringPtr("id ASC"),
	})
	if err != nil {
		err = errors.Wrapf(err, "list clusters")
		return
	}

	adminUser, err := UserService.GetDefaultAdmin(ctx)
	if err != nil {
		err = errors.Wrapf(err, "get default admin user")
		return
	}

	if total == 0 {
		defaultCluster, err = s.Create(ctx, CreateClusterOption{
			CreatorId:      adminUser.ID,
			OrganizationId: orgId,
			Name:           "default",
		})
		if err != nil {
			err = errors.Wrapf(err, "create default cluster")
			return
		}
		_, err = ClusterMemberService.Create(ctx, adminUser.ID, CreateClusterMemberOption{
			CreatorId: adminUser.ID,
			UserId:    adminUser.ID,
			ClusterId: defaultCluster.ID,
			Role:      modelschemas.MemberRoleAdmin,
		})
		if err != nil {
			err = errors.Wrapf(err, "create default cluster member")
			return
		}
	} else {
		defaultCluster = clusters[0]
	}

	return
}

func (s *clusterService) GetKubeCliSet(ctx context.Context, c *models.Cluster) (clientSet *kubernetes.Clientset, restConfig *rest.Config, err error) {
	if !config.YataiConfig.IsSass && c.KubeConfig == "" {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			kubeConfig :=
				clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
			restConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
			if err != nil {
				return nil, nil, errors.Wrap(err, "get in-cluster rest config")
			}
		}
	} else {
		configV1 := clientcmdapiv1.Config{}
		var jsonBytes []byte
		jsonBytes, err = yaml.YAMLToJSON([]byte(c.KubeConfig))
		if err != nil {
			return nil, nil, errors.Wrap(err, "k8s cluster config yaml to json")
		}
		err = json.Unmarshal(jsonBytes, &configV1)
		if err != nil {
			return nil, nil, errors.Wrap(err, "yaml unmarshal k8s cluster config")
		}

		var configObject runtime.Object
		configObject, err = clientcmdlatest.Scheme.ConvertToVersion(&configV1, clientcmdapi.SchemeGroupVersion)
		if err != nil {
			return nil, nil, errors.Wrap(err, "scheme convert to version")
		}
		configInternal := configObject.(*clientcmdapi.Config)

		restConfig, err = clientcmd.NewDefaultClientConfig(*configInternal, &clientcmd.ConfigOverrides{
			ClusterDefaults: clientcmdapi.Cluster{Server: ""},
		}).ClientConfig()

		if err != nil {
			return nil, nil, errors.Wrap(err, "new default k8s client config")
		}
	}

	restConfig.QPS = defaultQPS
	restConfig.Burst = defaultBurst

	clientSet, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, errors.Wrap(err, "new for k8s config")
	}

	return clientSet, restConfig, nil
}

func (s *clusterService) GetDockerRegistryRef(ctx context.Context, cluster *models.Cluster) (ref *modelschemas.DockerRegistryRefSchema, err error) {
	namespace := consts.KubeNamespaceYataiOperators
	name := "yatai-docker-registry-config"
	key := "config"
	ref = &modelschemas.DockerRegistryRefSchema{
		Namespace: namespace,
		Name:      name,
		Key:       key,
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, cluster)
	if err != nil {
		return ref, errors.Wrap(err, "get associated organization")
	}
	dockerRegistry, err := OrganizationService.GetDockerRegistry(ctx, org)
	if err != nil {
		return ref, errors.Wrap(err, "get docker registry")
	}
	content, err := json.Marshal(dockerRegistry)
	if err != nil {
		return ref, errors.Wrap(err, "marshal docker registry")
	}
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			key: content,
		},
	}
	clientSet, _, err := s.GetKubeCliSet(ctx, cluster)
	if err != nil {
		return ref, errors.Wrap(err, "get kube cli set")
	}
	secretCli := clientSet.CoreV1().Secrets(namespace)
	oldSecret, err := secretCli.Get(ctx, name, metav1.GetOptions{})
	isNotFound := apierrors.IsNotFound(err)
	if err != nil && !isNotFound {
		return ref, errors.Wrap(err, "get secret")
	}
	if !isNotFound {
		if bytes.Equal(content, oldSecret.Data[key]) {
			return ref, nil
		}
		oldSecret.Data[key] = content
		_, err = secretCli.Update(ctx, oldSecret, metav1.UpdateOptions{})
		if err != nil {
			return ref, errors.Wrap(err, "update secret")
		}
	} else {
		_, err = secretCli.Create(ctx, &secret, metav1.CreateOptions{})
		if err != nil {
			return ref, errors.Wrap(err, "create secret")
		}
	}
	return ref, nil
}

func (s *clusterService) GenerateGrafanaHostname(ctx context.Context, cluster *models.Cluster) (string, error) {
	clientset, _, err := s.GetKubeCliSet(ctx, cluster)
	if err != nil {
		return "", errors.Wrap(err, "get kube cli set")
	}
	domainSuffix, err := system.GetDomainSuffix(ctx, clientset)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("grafana-yatai-infra-external.%s", domainSuffix), nil
}

func (s *clusterService) GetGrafanaRootPath(ctx context.Context, cluster *models.Cluster) (string, error) {
	return fmt.Sprintf("/api/v1/clusters/%s/grafana/", cluster.Name), nil
}

func (s *clusterService) GetGrafana(ctx context.Context, cluster *models.Cluster) (*v1alpha1.Grafana, error) {
	_, ingLister, err := GetIngressInformer(ctx, cluster, consts.KubeNamespaceYataiComponents)
	if err != nil {
		return nil, err
	}

	ing, err := ingLister.Get("yatai-grafana")
	if err != nil {
		return nil, err
	}

	_, secretLister, err := GetSecretInformer(ctx, cluster, consts.KubeNamespaceYataiComponents)
	if err != nil {
		return nil, err
	}

	secret, err := secretLister.Get("yatai-grafana")
	if err != nil {
		return nil, err
	}

	password := secret.Data["admin-password"]
	user := secret.Data["admin-user"]

	grafanaConfig := v1alpha1.GrafanaConfig{
		Security: &v1alpha1.GrafanaConfigSecurity{
			AdminPassword: string(password),
			AdminUser:     string(user),
		},
	}

	return &v1alpha1.Grafana{
		Spec: v1alpha1.GrafanaSpec{
			Config: grafanaConfig,
			Ingress: &v1alpha1.GrafanaIngress{
				Hostname: ing.Spec.Rules[0].Host,
			},
		},
	}, err
}

func (s *clusterService) MakeSureDockerConfigSecret(ctx context.Context, cluster *models.Cluster, namespace string) (dockerConfigSecret *corev1.Secret, err error) {
	org, err := OrganizationService.GetAssociatedOrganization(ctx, cluster)
	if err != nil {
		return nil, err
	}

	dockerRegistry, err := OrganizationService.GetDockerRegistry(ctx, org)
	if err != nil {
		return nil, err
	}

	dockerConfigCMKubeName := "docker-config"
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

	kubeCli, _, err := s.GetKubeCliSet(ctx, cluster)
	if err != nil {
		return nil, err
	}

	secretsCli := kubeCli.CoreV1().Secrets(namespace)

	dockerConfigSecret, err = secretsCli.Get(ctx, dockerConfigCMKubeName, metav1.GetOptions{})
	dockerConfigIsNotFound := apierrors.IsNotFound(err)
	// nolint: gocritic
	if err != nil && !dockerConfigIsNotFound {
		return nil, err
	}
	err = nil
	if dockerConfigIsNotFound {
		dockerConfigSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: dockerConfigCMKubeName},
			StringData: map[string]string{
				"config.json": string(dockerConfigContent),
			},
		}
		_, err_ := secretsCli.Create(ctx, dockerConfigSecret, metav1.CreateOptions{})
		if err_ != nil {
			dockerConfigSecret, err = secretsCli.Get(ctx, dockerConfigCMKubeName, metav1.GetOptions{})
			dockerConfigIsNotFound = apierrors.IsNotFound(err)
			if err != nil && !dockerConfigIsNotFound {
				return nil, err
			}
			if dockerConfigIsNotFound {
				return nil, err_
			}
			if err != nil {
				err = nil
			}
		}
	} else {
		dockerConfigSecret.Data["config.json"] = dockerConfigContent
		_, err = secretsCli.Update(ctx, dockerConfigSecret, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
	}

	return
}

func (s *clusterService) MakeSureDockerRegcred(ctx context.Context, cluster *models.Cluster, namespace string) (secret *corev1.Secret, err error) {
	org, err := OrganizationService.GetAssociatedOrganization(ctx, cluster)
	if err != nil {
		return
	}

	dockerRegistry, err := OrganizationService.GetDockerRegistry(ctx, org)
	if err != nil {
		return
	}

	if dockerRegistry.Username != "" {
		var kubeCli *kubernetes.Clientset
		kubeCli, _, err = ClusterService.GetKubeCliSet(ctx, cluster)
		if err != nil {
			return
		}
		secretsCli := kubeCli.CoreV1().Secrets(namespace)
		secret, err = secretsCli.Get(ctx, consts.KubeSecretNameRegcred, metav1.GetOptions{})
		isNotFound := apierrors.IsNotFound(err)
		if err != nil && !isNotFound {
			return
		}
		dockerConfig := struct {
			Auths map[string]struct {
				Auth string `json:"auth"`
			} `json:"auths"`
		}{
			Auths: map[string]struct {
				Auth string `json:"auth"`
			}{
				dockerRegistry.Server: {
					Auth: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", dockerRegistry.Username, dockerRegistry.Password))),
				},
			},
		}
		var dockerConfigContent []byte
		dockerConfigContent, err = json.Marshal(&dockerConfig)
		if err != nil {
			return
		}
		if isNotFound {
			secret = &corev1.Secret{
				Type: corev1.SecretTypeDockerConfigJson,
				ObjectMeta: metav1.ObjectMeta{
					Name:      consts.KubeSecretNameRegcred,
					Namespace: namespace,
				},
				Data: map[string][]byte{
					".dockerconfigjson": dockerConfigContent,
				},
			}
			_, err = secretsCli.Create(ctx, secret, metav1.CreateOptions{})
			if err != nil {
				return
			}
		} else {
			secret.Data[".dockerconfigjson"] = dockerConfigContent
			_, err = secretsCli.Update(ctx, secret, metav1.UpdateOptions{})
			if err != nil {
				return
			}
		}
	}
	return
}

type IClusterAssociate interface {
	GetAssociatedClusterId() uint
	GetAssociatedClusterCache() *models.Cluster
	SetAssociatedClusterCache(user *models.Cluster)
}

func (s *clusterService) GetAssociatedCluster(ctx context.Context, associate IClusterAssociate) (*models.Cluster, error) {
	cache := associate.GetAssociatedClusterCache()
	if cache != nil {
		return cache, nil
	}
	cluster, err := s.Get(ctx, associate.GetAssociatedClusterId())
	associate.SetAssociatedClusterCache(cluster)
	return cluster, err
}

type INullableClusterAssociate interface {
	GetAssociatedClusterId() *uint
	GetAssociatedClusterCache() *models.Cluster
	SetAssociatedClusterCache(cluster *models.Cluster)
}

func (s *clusterService) GetAssociatedNullableCluster(ctx context.Context, associate INullableClusterAssociate) (*models.Cluster, error) {
	cache := associate.GetAssociatedClusterCache()
	if cache != nil {
		return cache, nil
	}
	clusterId := associate.GetAssociatedClusterId()
	if clusterId == nil {
		return nil, nil
	}
	cluster, err := s.Get(ctx, *clusterId)
	associate.SetAssociatedClusterCache(cluster)
	return cluster, err
}

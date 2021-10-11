package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
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
	Order          *string
}

func (s *clusterService) Create(ctx context.Context, opt CreateClusterOption) (*models.Cluster, error) {
	errs := validation.IsDNS1035Label(opt.Name)
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ";"))
	}

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
	err := mustGetSession(ctx).Create(&cluster).Error
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
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	query = opt.BindQuery(query)
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

func (s *clusterService) GetKubeCliSet(ctx context.Context, c *models.Cluster) (*kubernetes.Clientset, *rest.Config, error) {
	configV1 := clientcmdapiv1.Config{}
	jsonBytes, err := yaml.YAMLToJSON([]byte(c.KubeConfig))
	if err != nil {
		return nil, nil, errors.Wrap(err, "k8s cluster config yaml to json")
	}
	err = json.Unmarshal(jsonBytes, &configV1)
	if err != nil {
		return nil, nil, errors.Wrap(err, "yaml unmarshal k8s cluster config")
	}
	configObject, err := clientcmdlatest.Scheme.ConvertToVersion(&configV1, clientcmdapi.SchemeGroupVersion)
	if err != nil {
		return nil, nil, errors.Wrap(err, "scheme convert to version")
	}
	configInternal := configObject.(*clientcmdapi.Config)

	clientConfig, err := clientcmd.NewDefaultClientConfig(*configInternal, &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmdapi.Cluster{Server: ""},
	}).ClientConfig()

	if err != nil {
		return nil, nil, errors.Wrap(err, "new default k8s client config")
	}

	clientConfig.QPS = defaultQPS
	clientConfig.Burst = defaultBurst

	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, nil, errors.Wrap(err, "new for k8s config")
	}

	return clientSet, clientConfig, nil
}

func (s *clusterService) GetIngressIp(ctx context.Context, cluster *models.Cluster) (string, error) {
	ip := cluster.Config.IngressIp
	if ip == "" {
		return "", errors.Errorf("please specify the ingress ip or hostname in cluster %s", cluster.Name)
	}
	if net.ParseIP(ip) == nil {
		addr, err := net.LookupIP(ip)
		if err != nil {
			return "", errors.Wrapf(err, "lookup ip from ingress hostname %s in cluster %s", ip, cluster.Name)
		}
		if len(addr) == 0 {
			return "", errors.Errorf("cannot lookup ip from ingress hostname %s in cluster %s", ip, cluster.Name)
		}
		ip = addr[0].String()
	}
	return ip, nil
}

func (s *clusterService) GetGrafanaHostname(ctx context.Context, cluster *models.Cluster) (string, error) {
	ip, err := ClusterService.GetIngressIp(ctx, cluster)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("grafana.yatai-infra.%s.sslip.io", ip), nil
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

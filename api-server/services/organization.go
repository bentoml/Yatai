// nolint: goconst
package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/config"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/utils"
)

type organizationService struct{}

var OrganizationService = organizationService{}

func (*organizationService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Organization{})
}

type CreateOrganizationOption struct {
	CreatorId   uint
	Name        string
	Description string
	Config      *modelschemas.OrganizationConfigSchema
}

type UpdateOrganizationOption struct {
	Description *string
	Config      **modelschemas.OrganizationConfigSchema
}

type ListOrganizationOption struct {
	BaseListOption
	VisitorId *uint
	Ids       *[]uint
	Order     *string
}

func (s *organizationService) Create(ctx context.Context, opt CreateOrganizationOption) (*models.Organization, error) {
	errs := validation.IsDNS1035Label(opt.Name)
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ";"))
	}

	org := models.Organization{
		ResourceMixin: models.ResourceMixin{
			Name: opt.Name,
		},
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		Description: opt.Description,
		Config:      opt.Config,
	}
	err := mustGetSession(ctx).Create(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (s *organizationService) Update(ctx context.Context, o *models.Organization, opt UpdateOrganizationOption) (*models.Organization, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.Description != nil {
		updaters["description"] = *opt.Description
		defer func() {
			if err == nil {
				o.Description = *opt.Description
			}
		}()
	}
	if opt.Config != nil {
		updaters["config"] = *opt.Config
		defer func() {
			if err == nil {
				o.Config = *opt.Config
			}
		}()
	}
	if len(updaters) == 0 {
		return o, nil
	}
	err = s.getBaseDB(ctx).Where("id = ?", o.ID).Updates(updaters).Error
	return o, err
}

func (s *organizationService) Get(ctx context.Context, id uint) (*models.Organization, error) {
	var org models.Organization
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&org).Error
	if err != nil {
		return nil, err
	}
	if org.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &org, nil
}

func (s *organizationService) GetByUid(ctx context.Context, uid string) (*models.Organization, error) {
	var org models.Organization
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&org).Error
	if err != nil {
		return nil, err
	}
	if org.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &org, nil
}

func (s *organizationService) GetByName(ctx context.Context, name string) (*models.Organization, error) {
	var org models.Organization
	err := getBaseQuery(ctx, s).Where("name = ?", name).First(&org).Error
	if err != nil {
		return nil, err
	}
	if org.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &org, nil
}

func (s *organizationService) GetDefault(ctx context.Context) (*models.Organization, error) {
	var defaultOrg *models.Organization
	orgs, total, err := s.List(ctx, ListOrganizationOption{
		BaseListOption: BaseListOption{
			Start: utils.UintPtr(0),
			Count: utils.UintPtr(1),
		},
		Order: utils.StringPtr("id ASC"),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "list organizations")
	}

	adminUser, err := UserService.GetDefaultAdmin(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "get default admin user")
	}

	if total == 0 {
		defaultOrg, err = s.Create(ctx, CreateOrganizationOption{
			CreatorId: adminUser.ID,
			Name:      "default",
		})
		if err != nil {
			return nil, errors.Wrapf(err, "create default organization")
		}
		_, err = OrganizationMemberService.Create(ctx, adminUser.ID, CreateOrganizationMemberOption{
			CreatorId:      adminUser.ID,
			UserId:         adminUser.ID,
			OrganizationId: defaultOrg.ID,
			Role:           modelschemas.MemberRoleAdmin,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "create default organization member")
		}
	} else {
		defaultOrg = orgs[0]
	}
	return defaultOrg, nil
}

func (s *organizationService) List(ctx context.Context, opt ListOrganizationOption) ([]*models.Organization, uint, error) {
	orgs := make([]*models.Organization, 0)
	query := getBaseQuery(ctx, s)
	if opt.VisitorId != nil {
		orgIds, err := OrganizationMemberService.ListOrganizationIds(ctx, *opt.VisitorId)
		if err != nil {
			return nil, 0, errors.Wrap(err, "list organization ids")
		}
		// postgresql `in` clause cannot be empty, so push 0 to avoid it empty
		orgIds = append(orgIds, 0)
		query = query.Where("(creator_id = ? or id in (?))", *opt.VisitorId, orgIds)
	}
	if opt.Ids != nil {
		if len(*opt.Ids) == 0 {
			return orgs, 0, nil
		}
		query = query.Where("id in (?)", *opt.Ids)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	if opt.Ids == nil {
		if opt.Order != nil {
			query = query.Order(*opt.Order)
		} else {
			query = query.Order("id DESC")
		}
	}
	err = opt.BindQueryWithLimit(query).Find(&orgs).Error
	return orgs, uint(total), err
}

func (s *organizationService) GetUserOrganization(ctx context.Context, userId uint) (*models.Organization, error) {
	orgs, _, err := s.List(ctx, ListOrganizationOption{
		BaseListOption: BaseListOption{
			Start: utils.UintPtr(0),
			Count: utils.UintPtr(1),
		},
		VisitorId: utils.UintPtr(userId),
	})
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		return nil, errors.Wrap(consts.ErrNotFound, "cannot found organization")
	}
	return orgs[0], nil
}

type IOrganizationAssociate interface {
	GetAssociatedOrganizationId() uint
	GetAssociatedOrganizationCache() *models.Organization
	SetAssociatedOrganizationCache(organization *models.Organization)
}

func (s *organizationService) GetAssociatedOrganization(ctx context.Context, associate IOrganizationAssociate) (*models.Organization, error) {
	cache := associate.GetAssociatedOrganizationCache()
	if cache != nil {
		return cache, nil
	}
	organization, err := s.Get(ctx, associate.GetAssociatedOrganizationId())
	associate.SetAssociatedOrganizationCache(organization)
	return organization, err
}

type INullableOrganizationAssociate interface {
	GetAssociatedOrganizationId() *uint
	GetAssociatedOrganizationCache() *models.Organization
	SetAssociatedOrganizationCache(cluster *models.Organization)
}

func (s *organizationService) GetAssociatedNullableOrganization(ctx context.Context, associate INullableOrganizationAssociate) (*models.Organization, error) {
	cache := associate.GetAssociatedOrganizationCache()
	if cache != nil {
		return cache, nil
	}
	organizationId := associate.GetAssociatedOrganizationId()
	if organizationId == nil {
		return nil, nil
	}
	organization, err := s.Get(ctx, *organizationId)
	associate.SetAssociatedOrganizationCache(organization)
	return organization, err
}

func (s *organizationService) GetMajorCluster(ctx context.Context, org *models.Organization) (*models.Cluster, error) {
	if org.Config == nil || org.Config.MajorClusterUid == "" {
		clusters, _, err := ClusterService.List(ctx, ListClusterOption{
			BaseListOption: BaseListOption{
				Start: utils.UintPtr(0),
				Count: utils.UintPtr(1),
			},
			VisitorId:      nil,
			OrganizationId: nil,
			Ids:            nil,
			Order:          utils.StringPtr("id ASC"),
		})
		if err != nil {
			return nil, err
		}
		if len(clusters) == 0 {
			return nil, errors.Errorf("please add a cluster in organization %s", org.Name)
		}
		return clusters[0], nil
	}
	return ClusterService.GetByUid(ctx, org.Config.MajorClusterUid)
}

type S3Config struct {
	Endpoint                    string
	EndpointInCluster           string
	EndpointWithScheme          string
	EndpointWithSchemeInCluster string
	AccessKey                   string
	SecretKey                   string
	Secure                      bool
	Region                      string
	BentosBucketName            string
	ModelsBucketName            string
}

func (c *S3Config) GetMinioClient() (*minio.Client, error) {
	endpoint := c.Endpoint
	if config.YataiConfig.InCluster && !config.YataiConfig.IsSass {
		endpoint = c.EndpointInCluster
	}
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV2(c.AccessKey, c.SecretKey, ""),
		Secure: c.Secure,
	})
}

func (c *S3Config) MakeSureBucket(ctx context.Context, bucketName string) error {
	minioClient, err := c.GetMinioClient()
	if err != nil {
		return err
	}
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return errors.Wrapf(err, "get bucket %s exist", bucketName)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: c.Region})
		if err != nil {
			exists_, err_ := minioClient.BucketExists(ctx, bucketName)
			if err_ != nil {
				return errors.Wrapf(err_, "get bucket %s exist", bucketName)
			}
			if exists_ {
				return nil
			}
			return errors.Wrapf(err, "make bucket %s", bucketName)
		}
	}
	return nil
}

func (s *organizationService) GetS3Config(ctx context.Context, org *models.Organization) (conf *S3Config, err error) {
	if config.YataiConfig.S3 != nil {
		scheme := "http"
		if config.YataiConfig.S3.Secure {
			scheme = "https"
		}
		bentosBucketName := "yatai"
		if config.YataiConfig.S3.BucketName != "" {
			bentosBucketName = config.YataiConfig.S3.BucketName
		}
		modelsBucketName := "yatai"
		if config.YataiConfig.S3.BucketName != "" {
			modelsBucketName = config.YataiConfig.S3.BucketName
		}
		conf = &S3Config{
			Endpoint:                    config.YataiConfig.S3.Endpoint,
			EndpointInCluster:           config.YataiConfig.S3.Endpoint,
			EndpointWithScheme:          fmt.Sprintf("%s://%s", scheme, config.YataiConfig.S3.Endpoint),
			EndpointWithSchemeInCluster: fmt.Sprintf("%s://%s", scheme, config.YataiConfig.S3.Endpoint),
			AccessKey:                   config.YataiConfig.S3.AccessKey,
			SecretKey:                   config.YataiConfig.S3.SecretKey,
			Secure:                      config.YataiConfig.S3.Secure,
			Region:                      config.YataiConfig.S3.Region,
			BentosBucketName:            bentosBucketName,
			ModelsBucketName:            modelsBucketName,
		}
		return
	}
	if org.Config != nil && org.Config.S3 != nil && org.Config.S3.Endpoint != "" {
		s3Config := org.Config.S3
		endpoint := s3Config.Endpoint
		scheme := "http"
		if s3Config.Secure {
			scheme = "https"
		}
		bentosBucketName := "bentos"
		if s3Config.BentosBucketName != "" {
			bentosBucketName = s3Config.BentosBucketName
		}
		modelsBucketName := "models"
		if s3Config.ModelsBucketName != "" {
			modelsBucketName = s3Config.ModelsBucketName
		}
		conf = &S3Config{
			Endpoint:                    endpoint,
			EndpointInCluster:           endpoint,
			EndpointWithScheme:          fmt.Sprintf("%s://%s", scheme, endpoint),
			EndpointWithSchemeInCluster: fmt.Sprintf("%s://%s", scheme, endpoint),
			AccessKey:                   s3Config.AccessKey,
			SecretKey:                   s3Config.SecretKey,
			Secure:                      s3Config.Secure,
			Region:                      s3Config.Region,
			BentosBucketName:            bentosBucketName,
			ModelsBucketName:            modelsBucketName,
		}
		return
	}
	if org.Config != nil && org.Config.AWS != nil && org.Config.AWS.S3 != nil {
		awsS3Conf := org.Config.AWS.S3
		conf = &S3Config{
			Endpoint:                    consts.AmazonS3Endpoint,
			EndpointInCluster:           consts.AmazonS3Endpoint,
			EndpointWithScheme:          fmt.Sprintf("https://%s", consts.AmazonS3Endpoint),
			EndpointWithSchemeInCluster: fmt.Sprintf("https://%s", consts.AmazonS3Endpoint),
			AccessKey:                   org.Config.AWS.AccessKeyId,
			SecretKey:                   org.Config.AWS.SecretAccessKey,
			Secure:                      true,
			Region:                      awsS3Conf.Region,
			BentosBucketName:            awsS3Conf.BentosBucketName,
			ModelsBucketName:            awsS3Conf.ModelsBucketName,
		}
		return
	}
	err = errors.New("no s3 config")
	return
}

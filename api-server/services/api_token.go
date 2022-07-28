package services

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/bentoml/yatai-common/config"
	"github.com/bentoml/yatai-common/consts"
	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/utils"
)

type apiTokenService struct{}

var ApiTokenService = apiTokenService{}

func (*apiTokenService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.ApiToken{})
}

type CreateApiTokenOption struct {
	UserId         uint
	OrganizationId uint
	Name           string
	Description    string
	Scopes         *modelschemas.ApiTokenScopes
	ExpiredAt      *time.Time
}

type UpdateApiTokenOption struct {
	Description *string
	Scopes      **modelschemas.ApiTokenScopes
	ExpiredAt   **time.Time
	LastUsedAt  **time.Time
}

type ListApiTokenOption struct {
	BaseListOption
	VisitorId      *uint
	OrganizationId *uint
	Ids            *[]uint
	Order          *string
}

func (s *apiTokenService) Create(ctx context.Context, opt CreateApiTokenOption) (*models.ApiToken, error) {
	errs := validation.IsDNS1035Label(opt.Name)
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ";"))
	}

	guid := xid.New()
	token := guid.String()

	apiToken := models.ApiToken{
		ResourceMixin: models.ResourceMixin{
			Name: opt.Name,
		},
		Description: opt.Description,
		UserAssociate: models.UserAssociate{
			UserId: opt.UserId,
		},
		OrganizationAssociate: models.OrganizationAssociate{
			OrganizationId: opt.OrganizationId,
		},
		Token:     token,
		Scopes:    opt.Scopes,
		ExpiredAt: opt.ExpiredAt,
	}
	err := mustGetSession(ctx).Create(&apiToken).Error
	if err != nil {
		return nil, err
	}
	return &apiToken, err
}

func (s *apiTokenService) Update(ctx context.Context, c *models.ApiToken, opt UpdateApiTokenOption) (*models.ApiToken, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.Description != nil {
		updaters["description"] = *opt.Description
		defer func() {
			if err == nil {
				c.Description = *opt.Description
			}
		}()
	}
	if opt.Scopes != nil {
		updaters["scopes"] = *opt.Scopes
		defer func() {
			if err == nil {
				c.Scopes = *opt.Scopes
			}
		}()
	}
	if opt.ExpiredAt != nil {
		updaters["expired_at"] = *opt.ExpiredAt
		defer func() {
			if err == nil {
				c.ExpiredAt = *opt.ExpiredAt
			}
		}()
	}
	if opt.LastUsedAt != nil {
		updaters["last_used_at"] = *opt.LastUsedAt
		defer func() {
			if err == nil {
				c.LastUsedAt = *opt.LastUsedAt
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

func (s *apiTokenService) Get(ctx context.Context, id uint) (*models.ApiToken, error) {
	var apiToken models.ApiToken
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&apiToken).Error
	if err != nil {
		return nil, err
	}
	if apiToken.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &apiToken, nil
}

func (s *apiTokenService) Delete(ctx context.Context, m *models.ApiToken) (*models.ApiToken, error) {
	err := s.getBaseDB(ctx).Unscoped().Delete(m).Error
	return m, err
}

func (s *apiTokenService) GetByUid(ctx context.Context, uid string) (*models.ApiToken, error) {
	var apiToken models.ApiToken
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&apiToken).Error
	if err != nil {
		return nil, err
	}
	if apiToken.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &apiToken, nil
}

func (s *apiTokenService) GetByToken(ctx context.Context, token string) (*models.ApiToken, error) {
	pieces := strings.Split(token, ":")
	if len(pieces) == 3 && pieces[0] == consts.YataiApiTokenPrefixYataiDeploymentOperator {
		clusterName := pieces[1]
		token_ := pieces[2]
		defaultOrg, err := OrganizationService.GetDefault(ctx)
		if err != nil {
			err = errors.Wrap(err, "failed to get default organization")
			return nil, err
		}
		cluster, err := ClusterService.GetByName(ctx, defaultOrg.ID, clusterName)
		if err != nil {
			err = errors.Wrapf(err, "failed to get cluster %s", clusterName)
			return nil, err
		}
		cliset, _, err := ClusterService.GetKubeCliSet(ctx, cluster)
		if err != nil {
			err = errors.Wrap(err, "failed to get kube cli set")
			return nil, err
		}
		yataiConf, err := config.GetYataiConfig(ctx, cliset, consts.KubeNamespaceYataiDeploymentComponent, false)
		if err != nil {
			err = errors.Wrap(err, "failed to get yatai config")
			return nil, err
		}
		if token_ != yataiConf.ApiToken {
			err = errors.New("the api token is not valid")
			return nil, err
		}
		adminUser, err := UserService.GetDefaultAdmin(ctx)
		if err != nil {
			err = errors.Wrap(err, "failed to get default admin user")
			return nil, err
		}
		var apiToken *models.ApiToken
		apiToken, err = ApiTokenService.GetByName(ctx, defaultOrg.ID, adminUser.ID, consts.YataiK8sBotApiTokenName)
		apiTokenIsNotFound := utils.IsNotFound(err)
		if err != nil && !apiTokenIsNotFound {
			err = errors.Wrapf(err, "get api token")
			return nil, err
		}
		if apiTokenIsNotFound {
			scopes := modelschemas.ApiTokenScopes{
				modelschemas.ApiTokenScopeApi,
			}
			apiToken, err = ApiTokenService.Create(ctx, CreateApiTokenOption{
				Name:           consts.YataiK8sBotApiTokenName,
				OrganizationId: defaultOrg.ID,
				UserId:         adminUser.ID,
				Description:    "yatai k8s bot api token",
				Scopes:         &scopes,
			})
			if err != nil {
				err = errors.Wrapf(err, "create api token %s", consts.YataiK8sBotApiTokenName)
				return nil, err
			}
		}
		return apiToken, nil
	}
	var apiToken models.ApiToken
	err := getBaseQuery(ctx, s).Where("token = ?", token).First(&apiToken).Error
	if err != nil {
		return nil, err
	}
	if apiToken.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &apiToken, nil
}

func (s *apiTokenService) GetByName(ctx context.Context, organizationId, userId uint, name string) (*models.ApiToken, error) {
	var apiToken models.ApiToken
	err := getBaseQuery(ctx, s).Where("organization_id = ?", organizationId).Where("user_id = ?", userId).Where("name = ?", name).First(&apiToken).Error
	if err != nil {
		return nil, err
	}
	if apiToken.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &apiToken, nil
}

func (s *apiTokenService) List(ctx context.Context, opt ListApiTokenOption) ([]*models.ApiToken, uint, error) {
	apiTokens := make([]*models.ApiToken, 0)
	query := getBaseQuery(ctx, s)
	if opt.VisitorId != nil {
		query = query.Where("user_id = ?", opt.VisitorId)
	}
	if opt.OrganizationId != nil {
		query = query.Where("organization_id = ?", *opt.OrganizationId)
	}
	if opt.Ids != nil {
		if len(*opt.Ids) == 0 {
			return apiTokens, 0, nil
		}
		query = query.Where("id in (?)", *opt.Ids)
	} else {
		query = query.Where("name != ?", consts.YataiK8sBotApiTokenName)
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
	err = query.Find(&apiTokens).Error
	if err != nil {
		return nil, 0, err
	}
	return apiTokens, uint(total), err
}

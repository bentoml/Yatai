package services

import (
	"context"
	"fmt"

	jujuerrors "github.com/juju/errors"

	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type userService struct{}

var UserService = userService{}

func (*userService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.User{})
}

const LoginUserKey = "loginUser"

type CreateUserOption struct {
	Name           string
	FirstName      string
	LastName       string
	GithubUsername *string
	Email          *string
	Password       string
	Perm           *modelschemas.UserPerm
}

type UpdateUserOption struct {
	Config         **models.UserConfig
	GithubUsername **string
}

type ListUserOption struct {
	BaseListOption
	Perm  *modelschemas.UserPerm
	Order *string
}

func (s *userService) Create(ctx context.Context, opt CreateUserOption) (*models.User, error) {
	hashedPassword, err := generateHashedPassword(opt.Password)
	if err != nil {
		return nil, err
	}
	user := models.User{
		ResourceMixin: models.ResourceMixin{
			Name: opt.Name,
		},
		FirstName:      opt.FirstName,
		LastName:       opt.LastName,
		GithubUsername: opt.GithubUsername,
		Email:          opt.Email,
		Password:       string(hashedPassword),
		Perm:           modelschemas.UserPermDefault,
	}
	if opt.Perm != nil {
		user.Perm = *opt.Perm
	} else {
		_, total, err := s.List(ctx, ListUserOption{
			BaseListOption: BaseListOption{
				Start: utils.UintPtr(0),
				Count: utils.UintPtr(0),
			},
			Perm: modelschemas.UserPermPtr(modelschemas.UserPermAdmin),
		})
		if err != nil {
			return nil, errors.Wrap(err, "get user total count")
		}
		if total == 0 {
			user.Perm = modelschemas.UserPermAdmin
		}
	}
	err = mustGetSession(ctx).Create(&user).Error
	return &user, err
}

func (s *userService) Update(ctx context.Context, u *models.User, opt UpdateUserOption) (*models.User, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.Config != nil {
		updaters["config"] = *opt.Config
		defer func() {
			if err == nil {
				u.Config = *opt.Config
			}
		}()
	}
	if opt.GithubUsername != nil {
		updaters["github_username"] = *opt.GithubUsername
		defer func() {
			if err == nil {
				u.GithubUsername = *opt.GithubUsername
			}
		}()
	}
	if len(updaters) == 0 {
		return u, nil
	}
	err = s.getBaseDB(ctx).Where("id = ?", u.ID).Updates(updaters).Error
	return u, err
}

func (s *userService) UpdatePassword(ctx context.Context, u *models.User, currentPassword, newPassword string) (*models.User, error) {
	err := s.CheckPassword(ctx, u, currentPassword)
	if err != nil {
		return nil, err
	}
	hashedPassword, err := generateHashedPassword(newPassword)
	if err != nil {
		return nil, err
	}
	err = s.getBaseDB(ctx).Where("id = ?", u.ID).Updates(map[string]interface{}{
		"password": hashedPassword,
	}).Error
	return u, err
}

func (s *userService) CheckPassword(ctx context.Context, u *models.User, password string) error {
	if len(password) == 0 {
		return jujuerrors.Forbiddenf("password cannot be empty")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return jujuerrors.Forbiddenf("incorrect password")
	}
	return nil
}

func (s *userService) GenerateApiToken(ctx context.Context, u *models.User) (*models.User, error) {
	guid := xid.New()
	token := guid.String()
	err := s.getBaseDB(ctx).Where("id = ?", u.ID).Updates(map[string]interface{}{
		"api_token": token,
	}).Error
	if err == nil {
		u.ApiToken = token
	}
	return u, err
}

func (s *userService) DeleteApiToken(ctx context.Context, u *models.User) (*models.User, error) {
	err := s.getBaseDB(ctx).Where("id = ?", u.ID).Updates(map[string]interface{}{"api_token": ""}).Error
	if err == nil {
		u.ApiToken = ""
	}
	return u, err
}

func (s *userService) GetByApiToken(ctx context.Context, token string) (*models.User, error) {
	var u = models.User{}
	err := s.getBaseDB(ctx).Where("api_token = ?", token).First(&u).Error
	return &u, err
}

func generateHashedPassword(rawPassword string) ([]byte, error) {
	if len(rawPassword) == 0 {
		return []byte(""), nil
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), 8)
	if err != nil {
		return nil, errors.New("generate hashed password")
	}
	return hashedPassword, nil
}

func (*userService) GetUserDisplayName(user *models.User) string {
	if user.FirstName == "" && user.LastName == "" {
		return user.Name
	}
	return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
}

func (*userService) Get(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := mustGetSession(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &user, nil
}

func (*userService) GetByName(ctx context.Context, name string) (*models.User, error) {
	var user models.User
	err := mustGetSession(ctx).Where("name = ?", name).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (*userService) GetByGithubUsername(ctx context.Context, githubUsername string) (*models.User, error) {
	var user models.User
	err := mustGetSession(ctx).Where("github_username = ?", githubUsername).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (*userService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := mustGetSession(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &user, nil
}

func (s *userService) List(ctx context.Context, opt ListUserOption) ([]*models.User, uint, error) {
	query := getBaseQuery(ctx, s)
	if opt.Search != nil && *opt.Search != "" {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", *opt.Search))
	}
	if opt.Perm != nil {
		query = query.Where("perm = ?", *opt.Perm)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	users := make([]*models.User, 0)
	if opt.Order != nil {
		query = query.Order(*opt.Order)
	} else {
		query = query.Order("id DESC")
	}
	err = opt.BindQuery(query).Find(&users).Error
	return users, uint(total), err
}

func (*userService) ListByIds(ctx context.Context, ids []uint) ([]*models.User, error) {
	users := make([]*models.User, 0, len(ids))
	if len(ids) == 0 {
		return users, nil
	}
	err := mustGetSession(ctx).Where("id in (?)", ids).Find(&users).Error
	return users, err
}

func (*userService) ListByNames(ctx context.Context, names []string) ([]*models.User, error) {
	if len(names) == 0 {
		return nil, nil
	}
	users := make([]*models.User, 0, len(names))
	err := mustGetSession(ctx).Where("name in (?)", names).Find(&users).Error
	return users, err
}

func (*userService) IsAdmin(ctx context.Context, user *models.User, organization *models.Organization) bool {
	if user == nil {
		return false
	}
	if organization == nil {
		return user.IsSuperAdmin()
	}
	err := MemberService.CanOperate(ctx, &OrganizationMemberService, user.ID, organization.ID)
	return err == nil
}

type IUserAssociate interface {
	GetAssociatedUserId() uint
	GetAssociatedUserCache() *models.User
	SetAssociatedUserCache(user *models.User)
}

func (s *userService) GetAssociatedUser(ctx context.Context, associate IUserAssociate) (*models.User, error) {
	cache := associate.GetAssociatedUserCache()
	if cache != nil {
		return cache, nil
	}
	user, err := s.Get(ctx, associate.GetAssociatedUserId())
	associate.SetAssociatedUserCache(user)
	return user, err
}

type ICreatorAssociate interface {
	GetAssociatedCreatorId() uint
	GetAssociatedCreatorCache() *models.User
	SetAssociatedCreatorCache(user *models.User)
}

func (s *userService) GetAssociatedCreator(ctx context.Context, associate ICreatorAssociate) (*models.User, error) {
	cache := associate.GetAssociatedCreatorCache()
	if cache != nil {
		return cache, nil
	}
	user, err := s.Get(ctx, associate.GetAssociatedCreatorId())
	associate.SetAssociatedCreatorCache(user)
	return user, err
}

func SetLoginUser(ctx *gin.Context, user *models.User) {
	if user == nil {
		return
	}
	ctx.Set(LoginUserKey, user)
}

func GetCurrentUser(ctx context.Context) (*models.User, error) {
	user_ := ctx.Value(LoginUserKey)
	if user_ == nil {
		return nil, errors.Wrap(consts.ErrNotFound, "cannot find current user")
	}
	user, ok := user_.(*models.User)
	if !ok {
		return nil, errors.New("get login user err")
	}
	return user, nil
}

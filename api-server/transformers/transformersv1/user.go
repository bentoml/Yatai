package transformersv1

import (
	"context"
	// nolint: gosec
	"crypto/md5"
	"encoding/hex"

	"github.com/bentoml/yatai/common/utils"

	"github.com/bentoml/yatai/api-server/services"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

const gravatarMirrorUrl = "https://en.gravatar.com/avatar/"

func getAvatarUrl(user *models.User) (string, error) {
	if user.Email == nil {
		return "", nil
	}
	// nolint: gosec
	hasher := md5.New()
	_, err := hasher.Write([]byte(*user.Email))
	if err != nil {
		return "", err
	}
	md5Str := hex.EncodeToString(hasher.Sum(nil))
	return utils.UrlJoin(gravatarMirrorUrl, md5Str, map[string]string{
		"d": "robohash",
		"s": "300",
	}), nil
}

func ToUserSchema(ctx context.Context, user *models.User) (*schemasv1.UserSchema, error) {
	if user == nil {
		return nil, nil
	}
	ss, err := ToUserSchemas(ctx, []*models.User{user})
	if err != nil {
		return nil, errors.Wrap(err, "ToUserSchemas")
	}
	return ss[0], nil
}

func ToUserSchemas(ctx context.Context, users []*models.User) ([]*schemasv1.UserSchema, error) {
	resourceSchemasMap, err := ToResourceSchemasMap(ctx, users)
	if err != nil {
		return nil, errors.Wrap(err, "ToResourceSchemasMap")
	}
	res := make([]*schemasv1.UserSchema, 0, len(users))
	for _, u := range users {
		avatarUrl, err := getAvatarUrl(u)
		if err != nil {
			return nil, errors.Wrap(err, "get avatar url")
		}
		email := ""
		if u.Email != nil {
			email = *u.Email
		}
		resourceSchema, ok := resourceSchemasMap[u.GetUid()]
		if !ok {
			return nil, errors.Errorf("resource schema not found for user %s", u.GetUid())
		}
		res = append(res, &schemasv1.UserSchema{
			ResourceSchema: resourceSchema,
			FirstName:      u.FirstName,
			LastName:       u.LastName,
			Email:          email,
			AvatarUrl:      avatarUrl,
		})
	}
	return res, nil
}

type ICreatorAssociate interface {
	services.ICreatorAssociate
	models.IResource
}

func GetAssociatedCreatorSchema(ctx context.Context, associate ICreatorAssociate) (*schemasv1.UserSchema, error) {
	user, err := services.UserService.GetAssociatedCreator(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated creator", associate.GetResourceType(), associate.GetName())
	}
	userSchema, err := ToUserSchema(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "ToUserSchema")
	}
	return userSchema, nil
}

type IUserAssociate interface {
	services.IUserAssociate
	models.IResource
}

func GetAssociatedUserSchema(ctx context.Context, associate IUserAssociate) (*schemasv1.UserSchema, error) {
	user, err := services.UserService.GetAssociatedUser(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated user", associate.GetResourceType(), associate.GetName())
	}
	userSchema, err := ToUserSchema(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "ToUserSchema")
	}
	return userSchema, nil
}

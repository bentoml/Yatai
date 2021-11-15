package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToApiTokenSchema(ctx context.Context, apiToken *models.ApiToken) (*schemasv1.ApiTokenSchema, error) {
	if apiToken == nil {
		return nil, nil
	}
	ss, err := ToApiTokenSchemas(ctx, []*models.ApiToken{apiToken})
	if err != nil {
		return nil, errors.Wrap(err, "ToApiTokenSchemas")
	}
	return ss[0], nil
}

func ToApiTokenSchemas(ctx context.Context, apiTokens []*models.ApiToken) ([]*schemasv1.ApiTokenSchema, error) {
	res := make([]*schemasv1.ApiTokenSchema, 0, len(apiTokens))
	for _, apiToken := range apiTokens {
		userSchema, err := GetAssociatedUserSchema(ctx, apiToken)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedUserSchema")
		}
		organizationSchema, err := GetAssociatedOrganizationSchema(ctx, apiToken)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedOrganizationSchema")
		}
		res = append(res, &schemasv1.ApiTokenSchema{
			ResourceSchema: ToResourceSchema(apiToken),
			Description:    apiToken.Description,
			User:           userSchema,
			Organization:   organizationSchema,
			Scopes:         apiToken.Scopes,
			ExpiredAt:      apiToken.ExpiredAt,
			LastUsedAt:     apiToken.LastUsedAt,
			IsExpired:      apiToken.IsExpired(),
		})
	}
	return res, nil
}

func ToApiTokenFullSchema(ctx context.Context, apiToken *models.ApiToken) (*schemasv1.ApiTokenFullSchema, error) {
	if apiToken == nil {
		return nil, nil
	}
	s, err := ToApiTokenSchema(ctx, apiToken)
	if err != nil {
		return nil, errors.Wrap(err, "ToApiTokenSchema")
	}
	return &schemasv1.ApiTokenFullSchema{
		ApiTokenSchema: *s,
		Token:          apiToken.Token,
	}, nil
}

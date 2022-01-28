package services

import (
	"context"
	"fmt"
	"strings"

	jujuerrors "github.com/juju/errors"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/modelschemas"
)

type IMemberManager interface {
	CheckRoles(ctx context.Context, userId, resourceId uint, roles []modelschemas.MemberRole) (bool, error)
	GetOrganization(ctx context.Context, resourceId uint) (*models.Organization, error)
	GetResourceType() modelschemas.ResourceType
}

type memberService struct{}

var MemberService = memberService{}

func (s *memberService) checkApiToken(m IMemberManager, user *models.User, ops []modelschemas.ApiTokenScopeOp) error {
	if user.ApiToken == nil {
		return nil
	}
	if user.ApiToken.Scopes.Contains(modelschemas.ApiTokenScopeApi) {
		return nil
	}
	resourceType := m.GetResourceType()
	scopeStrs := make([]string, 0, len(ops))
	for _, op := range ops {
		scopeStr := fmt.Sprintf("%s_%s", op, resourceType)
		scopeStrs = append(scopeStrs, scopeStr)
	}
	for _, scopeStr := range scopeStrs {
		scope := modelschemas.ApiTokenScope(scopeStr)
		if user.ApiToken.Scopes.Contains(scope) {
			return nil
		}
	}
	return errors.Errorf("the api_token need the scopes: %s", strings.Join(scopeStrs, " or "))
}

func (s *memberService) CanView(ctx context.Context, m IMemberManager, user *models.User, resourceId uint) error {
	if err := s.checkApiToken(m, user, []modelschemas.ApiTokenScopeOp{modelschemas.ApiTokenScopeOpRead, modelschemas.ApiTokenScopeOpWrite, modelschemas.ApiTokenScopeOpOperate}); err != nil {
		return err
	}
	userId := user.ID
	resourceType := m.GetResourceType()
	resource, err := ResourceService.Get(ctx, resourceType, resourceId)
	if err != nil {
		return errors.Wrap(err, "check can view")
	}
	organization, err := m.GetOrganization(ctx, resourceId)
	if err != nil {
		return err
	}
	if organization.CreatorId == userId {
		return nil
	}

	if UserService.IsAdmin(ctx, user, organization) {
		return nil
	}
	can, err := m.CheckRoles(ctx, userId, resourceId, []modelschemas.MemberRole{
		modelschemas.MemberRoleGuest,
		modelschemas.MemberRoleDeveloper,
		modelschemas.MemberRoleAdmin,
	})
	if err != nil {
		return errors.Wrapf(err, "check can view")
	}
	if !can {
		return jujuerrors.Unauthorizedf("user %s cannot view this %s: %s", user.Name, resource.GetResourceType(), resource.GetName())
	}
	return nil
}

func (s *memberService) CanUpdate(ctx context.Context, m IMemberManager, user *models.User, resourceId uint) error {
	if err := s.checkApiToken(m, user, []modelschemas.ApiTokenScopeOp{modelschemas.ApiTokenScopeOpWrite, modelschemas.ApiTokenScopeOpOperate}); err != nil {
		return err
	}
	userId := user.ID
	resourceType := m.GetResourceType()
	resource, err := ResourceService.Get(ctx, resourceType, resourceId)
	if err != nil {
		return errors.Wrap(err, "check can update")
	}
	organization, err := m.GetOrganization(ctx, resourceId)
	if err != nil {
		return err
	}
	if organization.CreatorId == userId {
		return nil
	}

	if UserService.IsAdmin(ctx, user, organization) {
		return nil
	}
	can, err := m.CheckRoles(ctx, userId, resourceId, []modelschemas.MemberRole{
		modelschemas.MemberRoleDeveloper,
		modelschemas.MemberRoleAdmin,
	})
	if err != nil {
		return errors.Wrapf(err, "check can update")
	}
	if !can {
		return jujuerrors.Unauthorizedf("user %s cannot update this %s: %s", user.Name, resource.GetResourceType(), resource.GetName())
	}
	return nil
}

func (s *memberService) CanOperate(ctx context.Context, m IMemberManager, user *models.User, resourceId uint) error {
	if err := s.checkApiToken(m, user, []modelschemas.ApiTokenScopeOp{modelschemas.ApiTokenScopeOpOperate}); err != nil {
		return err
	}
	userId := user.ID
	resourceType := m.GetResourceType()
	resource, err := ResourceService.Get(ctx, resourceType, resourceId)
	if err != nil {
		return errors.Wrap(err, "check can operate")
	}
	organization, err := m.GetOrganization(ctx, resourceId)
	if err != nil {
		return err
	}
	if organization.CreatorId == userId {
		return nil
	}

	if user.IsSuperAdmin() {
		return nil
	}
	can, err := m.CheckRoles(ctx, userId, resourceId, []modelschemas.MemberRole{
		modelschemas.MemberRoleAdmin,
	})
	if err != nil {
		return errors.Wrapf(err, "check can operate")
	}
	if !can {
		return jujuerrors.Unauthorizedf("user %s cannot operate this %s: %s", user.Name, resource.GetResourceType(), resource.GetName())
	}
	return nil
}

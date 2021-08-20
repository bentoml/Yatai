package services

import (
	"context"

	"github.com/bentoml/yatai/schemas/modelschemas"

	jujuerrors "github.com/juju/errors"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
)

type IMemberManager interface {
	CheckRoles(ctx context.Context, userId, resourceId uint, roles []modelschemas.MemberRole) (bool, error)
	GetOrganization(ctx context.Context, resourceId uint) (*models.Organization, error)
	GetResourceType() modelschemas.ResourceType
}

type memberService struct{}

var MemberService = memberService{}

func (s *memberService) CanView(ctx context.Context, m IMemberManager, userId, resourceId uint) error {
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

	user, err := UserService.Get(ctx, userId)
	if err != nil {
		return errors.Wrapf(err, "check can view")
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

func (s *memberService) CanUpdate(ctx context.Context, m IMemberManager, userId, resourceId uint) error {
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

	user, err := UserService.Get(ctx, userId)
	if err != nil {
		return errors.Wrapf(err, "check can update")
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

func (s *memberService) CanOperate(ctx context.Context, m IMemberManager, userId, resourceId uint) error {
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

	user, err := UserService.Get(ctx, userId)
	if err != nil {
		return errors.Wrapf(err, "check can operate")
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

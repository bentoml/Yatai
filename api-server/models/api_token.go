package models

import (
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type ApiToken struct {
	ResourceMixin
	OrganizationAssociate
	UserAssociate
	Description string                       `json:"description"`
	Token       string                       `json:"token"`
	Scopes      *modelschemas.ApiTokenScopes `json:"scopes"`
	ExpiredAt   *time.Time                   `json:"expired_at"`
	LastUsedAt  *time.Time                   `json:"last_used_at"`
}

func (a *ApiToken) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeApiToken
}

func (a *ApiToken) IsExpired() bool {
	if a.ExpiredAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiredAt)
}

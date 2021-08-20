package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type User struct {
	ResourceMixin
	Perm            modelschemas.UserPerm `json:"perm"`
	FirstName       string                `json:"first_name"`
	LastName        string                `json:"last_name"`
	Email           string                `json:"email"`
	Password        string                `json:"password"`
	ApiToken        string                `json:"api_token"`
	IsEmailVerified bool                  `json:"is_email_verified"`
	GithubUsername  string                `json:"github_username"`
	Config          *UserConfig           `json:"config"`
}

type UserConfig struct {
	Theme string `json:"theme,omitempty"`
}

func (c *UserConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal([]byte(value.(string)), c)
}

func (c *UserConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

func (u *User) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeUser
}

func (u *User) IsSuperAdmin() bool {
	return u.Perm == modelschemas.UserPermAdmin
}

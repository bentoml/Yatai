package models

import (
	"time"

	"gorm.io/gorm"
)

type IBaseModel interface {
	GetId() uint
	GetUid() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() gorm.DeletedAt
}

type BaseModel struct {
	gorm.Model
	Uid string `json:"uid" gorm:"default:generate_object_id()"`
}

func (b *BaseModel) GetId() uint {
	return b.ID
}

func (b *BaseModel) GetUid() string {
	return b.Uid
}

func (b *BaseModel) GetCreatedAt() time.Time {
	return b.CreatedAt
}

func (b *BaseModel) GetUpdatedAt() time.Time {
	return b.UpdatedAt
}

func (b *BaseModel) GetDeletedAt() gorm.DeletedAt {
	return b.DeletedAt
}

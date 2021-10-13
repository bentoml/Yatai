package models

import (
	"gorm.io/gorm"
)

type Cache struct {
	gorm.Model
	Key   string `json:"key"`
	Value string `json:"value"`
}

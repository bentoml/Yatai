package models

import "gorm.io/gorm"

type BentoModelRel struct {
	gorm.Model
	BentoAssociate
	ModelAssociate
}

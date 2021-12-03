package models

import "gorm.io/gorm"

type BentoVersionModelVersionRel struct {
	gorm.Model
	BentoVersionAssociate
	ModelVersionAssociate
}

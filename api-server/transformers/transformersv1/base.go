package transformersv1

import (
	"time"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToBaseSchema(base models.IBaseModel) schemasv1.BaseSchema {
	var deletedAt *time.Time
	deletedAt_ := base.GetDeletedAt()
	if deletedAt_.Valid {
		deletedAt = &deletedAt_.Time
	}
	return schemasv1.BaseSchema{
		Uid:       base.GetUid(),
		CreatedAt: base.GetCreatedAt(),
		UpdatedAt: base.GetUpdatedAt(),
		DeletedAt: deletedAt,
	}
}

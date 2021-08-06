package utils

import (
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/common/consts"
)

func IsNotFound(err error) bool {
	return errors.Is(err, consts.ErrNotFound)
}

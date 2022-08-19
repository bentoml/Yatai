package web

import (
	"io/ioutil"
	"net/http"
	"path"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/bentoml/yatai/api-server/config"
	"github.com/bentoml/yatai/common/scookie"
)

var (
	indexContent  []byte
	indexLoadOnce sync.Once
)

func Index(ctx *gin.Context) {
	indexLoadOnce.Do(func() {
		var err error
		indexContent, err = ioutil.ReadFile(path.Join(config.GetUIDistDir(), "index.html"))
		if err != nil {
			logrus.Panicf("failed to read index.html:%s", err.Error())
		}
	})
	ctx.Data(200, "text/html; charset=utf-8", indexContent)
}

func Logout(ctx *gin.Context) {
	_ = scookie.DeleteUsernameFromCookie(ctx)
	ctx.Redirect(http.StatusFound, "/login")
}

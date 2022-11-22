package controllersv1

import (
	"time"

	"github.com/gin-gonic/gin"
	version2 "github.com/hashicorp/go-version"

	"github.com/bentoml/yatai/api-server/config"
	"github.com/bentoml/yatai/api-server/version"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/reqcli"
)

type newsController struct {
	// nolint: unused
	baseController
}

var NewsController = newsController{}

type NewsItem struct {
	Level             string `json:"level"`
	Title             string `json:"title"`
	Link              string `json:"link"`
	Cover             string `json:"cover"`
	StartedAt         string `json:"started_at"`
	EndedAt           string `json:"ended_at"`
	VersionConstraint string `json:"version_constraint"`
}

func (item NewsItem) Valid() (bool, error) {
	if item.VersionConstraint != "" {
		v, err := version2.NewVersion(version.Version)
		if err != nil {
			return false, err
		}
		constraints, err := version2.NewConstraint(item.VersionConstraint)
		if err != nil {
			return false, err
		}
		if !constraints.Check(v) {
			return false, nil
		}
	}
	now := time.Now()
	if item.StartedAt != "" {
		startedAt, err := time.Parse(time.RFC3339, item.StartedAt)
		if err != nil {
			return false, err
		}
		if startedAt.After(now) {
			return false, nil
		}
	}
	if item.EndedAt != "" {
		endedAt, err := time.Parse(time.RFC3339, item.EndedAt)
		if err != nil {
			return false, err
		}
		if endedAt.Before(now) {
			return false, nil
		}
	}
	return true, nil
}

type NewsContent struct {
	Notifications  []NewsItem `json:"notifications"`
	Documentations []NewsItem `json:"documentations"`
	ReleaseNotes   []NewsItem `json:"release_notes"`
	BlogPosts      []NewsItem `json:"blog_posts"`
}

func (c *newsController) GetNews(ctx *gin.Context) (news *NewsContent, err error) {
	newsUrl := consts.DefaultNewsURL
	if config.YataiConfig.NewsURL != "" {
		newsUrl = config.YataiConfig.NewsURL
	}
	var rawNews NewsContent
	_, err = reqcli.NewJsonRequestBuilder().Method("GET").Url(newsUrl).Result(&rawNews).Do(ctx)
	if err != nil {
		return nil, err
	}
	news = &NewsContent{}
	for _, item := range rawNews.Notifications {
		if ok, err := item.Valid(); err != nil {
			return nil, err
		} else if !ok {
			continue
		}
		news.Notifications = append(news.Notifications, item)
	}
	for _, item := range rawNews.Documentations {
		if ok, err := item.Valid(); err != nil {
			return nil, err
		} else if !ok {
			continue
		}
		news.Documentations = append(news.Documentations, item)
	}
	for _, item := range rawNews.ReleaseNotes {
		if ok, err := item.Valid(); err != nil {
			return nil, err
		} else if !ok {
			continue
		}
		news.ReleaseNotes = append(news.ReleaseNotes, item)
	}
	for _, item := range rawNews.BlogPosts {
		if ok, err := item.Valid(); err != nil {
			return nil, err
		} else if !ok {
			continue
		}
		news.BlogPosts = append(news.BlogPosts, item)
	}
	return
}

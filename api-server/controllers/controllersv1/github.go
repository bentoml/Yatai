package controllersv1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/bentoml/yatai/api-server/config"

	"github.com/pkg/errors"
	"github.com/rs/xid"

	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/common/scookie"
	"github.com/bentoml/yatai/common/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const sessionKeyState = "state"

var githubConfig *oauth2.Config

type GithubUser struct {
	Login             string    `json:"login"`
	ID                int       `json:"id"`
	NodeID            string    `json:"node_id"`
	AvatarURL         string    `json:"avatar_url"`
	GravatarID        string    `json:"gravatar_id"`
	URL               string    `json:"url"`
	HtmlUrl           string    `json:"html_url"`
	FollowersURL      string    `json:"followers_url"`
	FollowingURL      string    `json:"following_url"`
	GistsURL          string    `json:"gists_url"`
	StarredURL        string    `json:"starred_url"`
	SubscriptionsURL  string    `json:"subscriptions_url"`
	OrganizationsURL  string    `json:"organizations_url"`
	ReposURL          string    `json:"repos_url"`
	EventsURL         string    `json:"events_url"`
	ReceivedEventsURL string    `json:"received_events_url"`
	Type              string    `json:"type"`
	SiteAdmin         bool      `json:"site_admin"`
	Name              string    `json:"name"`
	Company           string    `json:"company"`
	Blog              string    `json:"blog"`
	Location          string    `json:"location"`
	Email             string    `json:"email"`
	Hireable          bool      `json:"hireable"`
	Bio               string    `json:"bio"`
	TwitterUserName   string    `json:"twitter_username"`
	PublicRepos       int       `json:"public_repos"`
	PublicGists       int       `json:"public_gits"`
	Followers         int       `json:"followers"`
	Following         int       `json:"following"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func getGithubOAuthURL(ctx *gin.Context) (*oauth2.Config, string) {
	scheme := "http"
	if config.YataiConfig.Server.EnableHTTPS {
		scheme = "https"
	}
	redirectUri := ctx.Query("redirect")
	if redirectUri == "" {
		redirectUri = "/"
	}
	redirectUrl := fmt.Sprintf("%s://%s/callback/github?redirect=%s", scheme, ctx.Request.Host, url.PathEscape(redirectUri))
	clientId := config.YataiConfig.OAuth.Github.ClientId
	clientSecret := config.YataiConfig.OAuth.Github.ClientSecret

	githubConfig = &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
		Scopes: []string{
			"user",
		},
		Endpoint: github.Endpoint,
	}

	state := xid.New().String()
	return githubConfig, state
}

func GithubOAuthLogin(ctx *gin.Context) {
	config, state := getGithubOAuthURL(ctx)
	redirectURL := config.AuthCodeURL(state)

	session := sessions.Default(ctx)
	session.Set(sessionKeyState, state)
	err := session.Save()
	if err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, redirectURL)
}

func GithubOAuthCallBack(ctx *gin.Context) {
	session := sessions.Default(ctx)
	state := session.Get(sessionKeyState)
	if state != ctx.Query("state") {
		_ = ctx.AbortWithError(http.StatusUnauthorized, errors.New("state error"))
		return
	}

	code := ctx.Query("code")
	token, err := githubConfig.Exchange(ctx, code)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	client := githubConfig.Client(ctx, token)
	// nolint:noctx
	userInfo, err := client.Get("https://api.github.com/user")
	if err != nil {
		_ = ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer userInfo.Body.Close()

	info, err := ioutil.ReadAll(userInfo.Body)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var githubUser GithubUser
	err = json.Unmarshal(info, &githubUser)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	user, err := services.UserService.GetByEmail(ctx, githubUser.Email)
	userIsNotFound := utils.IsNotFound(err)
	if err != nil && !userIsNotFound {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if userIsNotFound {
		userName := githubUser.Name
		total := 1000

		for i := 0; i < total; i++ {
			_, err = services.UserService.GetByName(ctx, userName)
			userIsNotFound = utils.IsNotFound(err)
			if err != nil && !userIsNotFound {
				_ = ctx.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			if userIsNotFound {
				break
			}
			userName = fmt.Sprintf("%s-%d", githubUser.Name, i)
		}

		user, err = services.UserService.Create(ctx, services.CreateUserOption{
			Name:           userName,
			FirstName:      githubUser.Name,
			LastName:       "",
			Email:          githubUser.Email,
			Password:       "",
			GithubUsername: githubUser.Name,
		})
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	} else {
		user, err = services.UserService.Update(ctx, user, services.UpdateUserOption{
			GithubUsername: utils.StringPtr(githubUser.Name),
		})
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	err = scookie.SetUsernameToCookie(ctx, user.Name)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	redirectUri := ctx.Query("redirect")
	if redirectUri == "" {
		redirectUri = "/"
	}

	ctx.Redirect(http.StatusSeeOther, redirectUri)
}

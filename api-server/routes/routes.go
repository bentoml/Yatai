package routes

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/pkg/errors"
	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"

	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/config"
	"github.com/bentoml/yatai/api-server/controllers/controllersv1"
	"github.com/bentoml/yatai/api-server/controllers/web"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/scookie"
	"github.com/bentoml/yatai/common/yataicontext"
)

var pwd, _ = os.Getwd()

var staticDirs = map[string]string{
	"/swagger": path.Join(pwd, "statics/swagger-ui"),
	"/static":  path.Join(config.GetUIDistDir(), "static"),
	"/libs":    path.Join(config.GetUIDistDir(), "libs"),
}

var staticFiles = map[string]string{
	"/favicon.ico": path.Join(config.GetUIDistDir(), "favicon.ico"),
}

const WebsocketConnectContextKey = "websocket-connect"

func NewRouter() (*fizz.Fizz, error) {
	tonic.SetRenderHook(func(c *gin.Context, statusCode int, payload interface{}) {
		if _, exists := c.Get(WebsocketConnectContextKey); exists {
			return
		}
		var status int
		if c.Writer.Written() {
			status = c.Writer.Status()
		} else {
			status = statusCode
		}
		if payload != nil {
			if gin.IsDebugging() {
				c.IndentedJSON(status, payload)
			} else {
				c.JSON(status, payload)
			}
		} else {
			c.String(status, "")
		}
	}, "")

	engine := gin.New()

	store := cookie.NewStore([]byte(config.YataiConfig.Server.SessionSecretKey))
	engine.Use(sessions.Sessions("yatai-session-v1", store))

	engine.GET("/logout", web.Logout)

	fizzApp := fizz.NewFromEngine(engine)

	// Override type names.
	// fizz.Generator().OverrideTypeName(reflect.TypeOf(Fruit{}), "SweetFruit")

	// Initialize the information of
	// the API that will be served with
	// the specification.
	infos := &openapi.Info{
		Title:       "yatai api server",
		Description: `This is yatai api server.`,
		Version:     "1.0.0",
	}
	// Create a new route that serve the OpenAPI spec.
	fizzApp.GET("/openapi.json", nil, fizzApp.OpenAPI(infos, "json"))

	wsRootGroup := fizzApp.Group("/ws/v1", "websocket v1", "websocket v1")
	wsRootGroup.Use(func(c *gin.Context) {
		c.Set(WebsocketConnectContextKey, true)
		c.Next()
	})
	wsRootGroup.Use(requireLogin)
	wsRootGroup.GET("/subscription/resource", []fizz.OperationOption{
		fizz.ID("Subscribe resource"),
		fizz.Summary("Subscribe resource"),
	}, tonic.Handler(controllersv1.SubscriptionController.SubscribeResource, 200))

	wsRootGroup.GET("/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/tail", []fizz.OperationOption{
		fizz.ID("Tail deployment pod log"),
		fizz.Summary("Tail deployment pod log"),
	}, tonic.Handler(controllersv1.LogController.TailDeploymentPodLog, 200))

	wsRootGroup.GET("/clusters/:clusterName/tail", []fizz.OperationOption{
		fizz.ID("Tail cluster pod log"),
		fizz.Summary("Tail cluster pod log"),
	}, tonic.Handler(controllersv1.LogController.TailClusterPodLog, 200))

	wsRootGroup.GET("/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/terminal", []fizz.OperationOption{
		fizz.ID("Deployment pod terminal"),
		fizz.Summary("Deployment pod terminal"),
	}, tonic.Handler(controllersv1.TerminalController.GetDeploymentPodTerminal, 200))

	wsRootGroup.GET("/clusters/:clusterName/terminal", []fizz.OperationOption{
		fizz.ID("Cluster pod terminal"),
		fizz.Summary("Cluster pod terminal"),
	}, tonic.Handler(controllersv1.TerminalController.GetClusterPodTerminal, 200))

	wsRootGroup.GET("/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/kube_events", []fizz.OperationOption{
		fizz.ID("Deployment kube events"),
		fizz.Summary("Deployment kube events"),
	}, tonic.Handler(controllersv1.KubeController.GetDeploymentKubeEvents, 200))

	wsRootGroup.GET("/clusters/:clusterName/kube_events", []fizz.OperationOption{
		fizz.ID("Cluster kube events"),
		fizz.Summary("Cluster kube events"),
	}, tonic.Handler(controllersv1.KubeController.GetPodKubeEvents, 200))

	wsRootGroup.GET("/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/pods", []fizz.OperationOption{
		fizz.ID("Ws deployment pods"),
		fizz.Summary("Ws deployment pods"),
	}, tonic.Handler(controllersv1.DeploymentController.WsPods, 200))

	wsRootGroup.GET("/clusters/:clusterName/pods", []fizz.OperationOption{
		fizz.ID("Ws cluster pods"),
		fizz.Summary("Ws cluster pods"),
	}, tonic.Handler(controllersv1.ClusterController.WsPods, 200))

	clusterGroup := engine.Group("/api/v1/clusters/:clusterName")
	clusterGroup.Use(requireLogin)

	clusterGroup.GET("/grafana/*path", controllersv1.GrafanaController.Proxy)
	clusterGroup.POST("/grafana/*path", controllersv1.GrafanaController.Proxy)
	clusterGroup.PUT("/grafana/*path", controllersv1.GrafanaController.Proxy)
	clusterGroup.PATCH("/grafana/*path", controllersv1.GrafanaController.Proxy)
	clusterGroup.HEAD("/grafana/*path", controllersv1.GrafanaController.Proxy)
	clusterGroup.DELETE("/grafana/*path", controllersv1.GrafanaController.Proxy)

	bentoGroup := engine.Group("/api/v1/bento_repositories/:bentoRepositoryName/bentos/:version")
	bentoGroup.Use(requireLogin)

	bentoGroup.PUT("/upload", controllersv1.BentoController.Upload)
	bentoGroup.GET("/download", controllersv1.BentoController.Download)

	modelGroup := engine.Group("/api/v1/model_repositories/:modelRepositoryName/models/:version")
	modelGroup.Use(requireLogin)

	modelGroup.PUT("/upload", controllersv1.ModelController.Upload)
	modelGroup.GET("/download", controllersv1.ModelController.Download)

	publicApiRootGroup := fizzApp.Group("/api/v1", "api v1", "api v1")
	apiRootGroup := fizzApp.Group("/api/v1", "api v1", "api v1")
	apiRootGroup.Use(requireLogin)

	// Setup routes.
	authRoutes(publicApiRootGroup)
	userRoutes(apiRootGroup)
	organizationRoutes(apiRootGroup)
	apiTokenRoutes(apiRootGroup)
	labelRoutes(apiRootGroup)
	clusterRoutes(apiRootGroup)
	bentoRepositoryRoutes(apiRootGroup)
	modelRepositoryRoutes(apiRootGroup)
	terminalRecordRoutes(apiRootGroup)

	publicApiRootGroup.GET("/version", []fizz.OperationOption{
		fizz.ID("Get version"),
		fizz.Summary("Get version"),
	}, tonic.Handler(controllersv1.VersionController.GetVersion, 200))

	apiRootGroup.GET("/news", []fizz.OperationOption{
		fizz.ID("Get news"),
		fizz.Summary("Get news"),
	}, tonic.Handler(controllersv1.NewsController.GetNews, 200))

	apiRootGroup.GET("/bentos", []fizz.OperationOption{
		fizz.ID("List all bentos"),
		fizz.Summary("List all bentos"),
	}, tonic.Handler(controllersv1.BentoController.ListAll, 200))

	apiRootGroup.GET("/image_build_status_unsynced_bentos", []fizz.OperationOption{
		fizz.ID("List image build status unsynced bentos"),
		fizz.Summary("List image build status unsynced bentos"),
	}, tonic.Handler(controllersv1.BentoController.ListImageBuildStatusUnsynced, 200))

	apiRootGroup.GET("/models", []fizz.OperationOption{
		fizz.ID("List all models"),
		fizz.Summary("List all models"),
	}, tonic.Handler(controllersv1.ModelController.ListAll, 200))

	publicApiRootGroup.POST("/setup", []fizz.OperationOption{
		fizz.ID("Setup admin user, org, cluster for selfhosted mode"),
		fizz.Summary("Setup admin user, org, cluster for selfhosted mode"),
	}, tonic.Handler(controllersv1.SelfHostedController.Setup, 200))

	if len(fizzApp.Errors()) != 0 {
		return nil, fmt.Errorf("fizz errors: %v", fizzApp.Errors())
	}

	for p, root := range staticDirs {
		engine.Static(p, root)
	}

	for f, root := range staticFiles {
		engine.StaticFile(f, root)
	}

	engine.NoRoute(func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.Request.URL.Path, "/api/") {
			ctx.JSON(http.StatusNotFound, &schemasv1.MsgSchema{Message: fmt.Sprintf("not found this router with method %s", ctx.Request.Method)})
			return
		}

		for p := range staticDirs {
			if strings.HasPrefix(ctx.Request.URL.Path, p) {
				ctx.JSON(http.StatusNotFound, &schemasv1.MsgSchema{Message: fmt.Sprintf("not found this router with method %s", ctx.Request.Method)})
				return
			}
		}

		web.Index(ctx)
	})

	return fizzApp, nil
}

func getLoginUser(ctx *gin.Context) (user *models.User, err error) {
	apiTokenStr := ctx.GetHeader(consts.YataiApiTokenHeaderName)

	// nolint: gocritic
	if apiTokenStr != "" {
		var apiToken *models.ApiToken
		apiToken, err = services.ApiTokenService.GetByToken(ctx, apiTokenStr)
		if err != nil {
			err = errors.Wrap(err, "get api token")
			return
		}
		if apiToken.IsExpired() {
			err = errors.New("the api token is expired")
			return
		}
		user, err = services.UserService.GetAssociatedUser(ctx, apiToken)
		if err != nil {
			err = errors.Wrap(err, "get user by api token")
			return
		}
		now := time.Now()
		now_ := &now
		apiToken, err = services.ApiTokenService.Update(ctx, apiToken, services.UpdateApiTokenOption{
			LastUsedAt: &now_,
		})
		if err != nil {
			err = errors.Wrap(err, "update api token")
			return
		}
		user.ApiToken = apiToken
	} else {
		username := scookie.GetUsernameFromCookie(ctx)
		if username == "" {
			err = errors.New("username in cookie is empty")
			return
		}
		user, err = services.UserService.GetByName(ctx, username)
		if err != nil {
			err = errors.Wrapf(err, "get user by name in cookie %s", username)
			return
		}
	}

	yataicontext.SetUserName(ctx, user.Name)
	services.SetLoginUser(ctx, user)
	return
}

func requireLogin(ctx *gin.Context) {
	_, loginErr := getLoginUser(ctx)
	if loginErr != nil {
		msg := schemasv1.MsgSchema{Message: loginErr.Error()}
		ctx.AbortWithStatusJSON(http.StatusForbidden, &msg)
		return
	}

	// https://github.com/gorilla/handlers/pull/187
	if ctx.GetHeader("Upgrade") == "" {
		ctx.Next()
	}
}

func authRoutes(publicGrp *fizz.RouterGroup) {
	grp := publicGrp.Group("/auth", "auth", "auth api")
	grp.Use(requireLogin)
	publicGrp = publicGrp.Group("/auth", "auth", "auth api")

	publicGrp.POST("/register", []fizz.OperationOption{
		fizz.ID("Register an user"),
		fizz.Summary("Register an user"),
	}, tonic.Handler(controllersv1.AuthController.Register, 200))

	publicGrp.POST("/login", []fizz.OperationOption{
		fizz.ID("Login an user"),
		fizz.Summary("Login an user"),
	}, tonic.Handler(controllersv1.AuthController.Login, 200))

	grp.GET("/current", []fizz.OperationOption{
		fizz.ID("Get current user"),
		fizz.Summary("Get current user"),
	}, tonic.Handler(controllersv1.AuthController.GetCurrentUser, 200))

	grp.PATCH("/reset_password", []fizz.OperationOption{
		fizz.ID("Reset password"),
		fizz.Summary("Reset password"),
	}, tonic.Handler(controllersv1.AuthController.ResetPassword, 200))
}

func userRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/users", "users", "users api")

	resourceGrp := grp.Group("/:userName", "user resource", "user resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get an user"),
		fizz.Summary("Get an user"),
	}, tonic.Handler(controllersv1.UserController.Get, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List users"),
		fizz.Summary("List users"),
	}, tonic.Handler(controllersv1.UserController.List, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create an user"),
		fizz.Summary("Create an user"),
	}, tonic.Handler(controllersv1.UserController.Create, 200))
}

func organizationRoutes(grp *fizz.RouterGroup) {
	resourceGrp := grp.Group("/current_org", "organization resource", "organization resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get an organization"),
		fizz.Summary("Get an organization"),
	}, tonic.Handler(controllersv1.OrganizationController.Get, 200))

	resourceGrp.GET("/major_cluster", []fizz.OperationOption{
		fizz.ID("Get an organization major cluster"),
		fizz.Summary("Get an organization major cluster"),
	}, tonic.Handler(controllersv1.OrganizationController.GetMajorCluster, 200))

	resourceGrp.GET("/model_modules", []fizz.OperationOption{
		fizz.ID("Get an organization model modules"),
		fizz.Summary("Get an organization model modules"),
	}, tonic.Handler(controllersv1.OrganizationController.ListModelModules, 200))

	resourceGrp.GET("/events", []fizz.OperationOption{
		fizz.ID("List current organization events"),
		fizz.Summary("List current organization events"),
	}, tonic.Handler(controllersv1.OrganizationController.ListEvents, 200))

	resourceGrp.GET("/event_operation_names", []fizz.OperationOption{
		fizz.ID("List current organization event operation names"),
		fizz.Summary("List current organization event operation names"),
	}, tonic.Handler(controllersv1.OrganizationController.ListEventOperationNames, 200))

	resourceGrp.PATCH("", []fizz.OperationOption{
		fizz.ID("Update an organization"),
		fizz.Summary("Update an organization"),
	}, tonic.Handler(controllersv1.OrganizationController.Update, 200))

	grp.GET("/members", []fizz.OperationOption{
		fizz.ID("List organization members"),
		fizz.Summary("Get organization members"),
	}, tonic.Handler(controllersv1.OrganizationMemberController.List, 200))

	grp.POST("/members", []fizz.OperationOption{
		fizz.ID("Create an organization member"),
		fizz.Summary("Create an organization member"),
	}, tonic.Handler(controllersv1.OrganizationMemberController.Create, 200))

	grp.DELETE("/members", []fizz.OperationOption{
		fizz.ID("Remove an organization member"),
		fizz.Summary("Remove an organization member"),
	}, tonic.Handler(controllersv1.OrganizationMemberController.Delete, 200))

	grp.GET("/deployments", []fizz.OperationOption{
		fizz.ID("List organization deployments"),
		fizz.Summary("List organization deployments"),
	}, tonic.Handler(controllersv1.DeploymentController.ListOrganizationDeployments, 200))

	grp.GET("/orgs", []fizz.OperationOption{
		fizz.ID("List organizations"),
		fizz.Summary("List organizations"),
	}, tonic.Handler(controllersv1.OrganizationController.List, 200))

	grp.POST("/orgs", []fizz.OperationOption{
		fizz.ID("Create organization"),
		fizz.Summary("Create organization"),
	}, tonic.Handler(controllersv1.OrganizationController.Create, 200))

	// clusterRoutes(resourceGrp)
	// bentoRepositoryRoutes(resourceGrp)
	// modelRepositoryRoutes(resourceGrp)
}

func apiTokenRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/api_tokens", "api tokens", "api tokens")

	resourceGrp := grp.Group("/:apiTokenUid", "api token resource", "api token resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a api token"),
		fizz.Summary("Get a api token"),
	}, tonic.Handler(controllersv1.ApiTokenController.Get, 200))

	resourceGrp.PATCH("", []fizz.OperationOption{
		fizz.ID("Update a api token"),
		fizz.Summary("Update a api token"),
	}, tonic.Handler(controllersv1.ApiTokenController.Update, 200))

	resourceGrp.DELETE("", []fizz.OperationOption{
		fizz.ID("Delete a api token"),
		fizz.Summary("Delete a api token"),
	}, tonic.Handler(controllersv1.ApiTokenController.Delete, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List api tokens"),
		fizz.Summary("List api tokens"),
	}, tonic.Handler(controllersv1.ApiTokenController.List, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create api token"),
		fizz.Summary("Create api token"),
	}, tonic.Handler(controllersv1.ApiTokenController.Create, 200))
}

func labelRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/labels", "labels", "labels")
	grp.GET("", []fizz.OperationOption{
		fizz.ID("List Labels"),
		fizz.Summary("List Labels"),
	}, tonic.Handler(controllersv1.LabelController.List, 200))
}

func clusterRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/clusters", "clusters", "clusters")

	resourceGrp := grp.Group("/:clusterName", "cluster resource", "cluster resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a cluster"),
		fizz.Summary("Get a cluster"),
	}, tonic.Handler(controllersv1.ClusterController.Get, 200))

	resourceGrp.PATCH("", []fizz.OperationOption{
		fizz.ID("Update a cluster"),
		fizz.Summary("Update a cluster"),
	}, tonic.Handler(controllersv1.ClusterController.Update, 200))

	resourceGrp.GET("/docker_registry_ref", []fizz.OperationOption{
		fizz.ID("Get a cluster docker registry ref"),
		fizz.Summary("Get a cluster docker registry ref"),
	}, tonic.Handler(controllersv1.ClusterController.GetDockerRegistryRef, 200))

	resourceGrp.GET("/members", []fizz.OperationOption{
		fizz.ID("List cluster members"),
		fizz.Summary("List cluster members"),
	}, tonic.Handler(controllersv1.ClusterMemberController.List, 200))

	resourceGrp.POST("/members", []fizz.OperationOption{
		fizz.ID("Create a cluster member"),
		fizz.Summary("Create a cluster member"),
	}, tonic.Handler(controllersv1.ClusterMemberController.Create, 200))

	resourceGrp.DELETE("/members", []fizz.OperationOption{
		fizz.ID("Remove a cluster member"),
		fizz.Summary("Remove a cluster member"),
	}, tonic.Handler(controllersv1.ClusterMemberController.Delete, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List clusters"),
		fizz.Summary("List clusters"),
	}, tonic.Handler(controllersv1.ClusterController.List, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create cluster"),
		fizz.Summary("Create cluster"),
	}, tonic.Handler(controllersv1.ClusterController.Create, 200))

	deploymentRoutes(resourceGrp)
}

func bentoRepositoryRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/bento_repositories", "bento repositories", "bento repositories")

	resourceGrp := grp.Group("/:bentoRepositoryName", "bento repository resource", "bento repository resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a bento repository"),
		fizz.Summary("Get a bento repository"),
	}, tonic.Handler(controllersv1.BentoRepositoryController.Get, 200))

	resourceGrp.PATCH("", []fizz.OperationOption{
		fizz.ID("Update a bento repository"),
		fizz.Summary("Update a bento repository"),
	}, tonic.Handler(controllersv1.BentoRepositoryController.Update, 200))

	resourceGrp.GET("/deployments", []fizz.OperationOption{
		fizz.ID("List bento repository deployments"),
		fizz.Summary("List bento repository deployments"),
	}, tonic.Handler(controllersv1.BentoRepositoryController.ListDeployment, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List bento repositories"),
		fizz.Summary("List bento repositories"),
	}, tonic.Handler(controllersv1.BentoRepositoryController.List, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create bento repository"),
		fizz.Summary("Create bento repository"),
	}, tonic.Handler(controllersv1.BentoRepositoryController.Create, 200))

	bentoRoutes(resourceGrp)
}

func bentoRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/bentos", "bentos", "bentos")

	resourceGrp := grp.Group("/:version", "bento resource", "bento resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a bento"),
		fizz.Summary("Get a bento"),
	}, tonic.Handler(controllersv1.BentoController.Get, 200))

	resourceGrp.PATCH("", []fizz.OperationOption{
		fizz.ID("Update a bento"),
		fizz.Summary("Update a bento"),
	}, tonic.Handler(controllersv1.BentoController.Update, 200))

	resourceGrp.PATCH("/update_image_build_status_syncing_at", []fizz.OperationOption{
		fizz.ID("Update a bento image build status syncing_at"),
		fizz.Summary("Update a bento image build status syncing_at"),
	}, tonic.Handler(controllersv1.BentoController.UpdateBentoImageBuildStatusSyncingAt, 200))

	resourceGrp.PATCH("/update_image_build_status", []fizz.OperationOption{
		fizz.ID("Update a bento image build status"),
		fizz.Summary("Update a bento image build status"),
	}, tonic.Handler(controllersv1.BentoController.UpdateBentoImageBuildStatus, 200))

	resourceGrp.GET("/models", []fizz.OperationOption{
		fizz.ID("List bento models"),
		fizz.Summary("List bento models"),
	}, tonic.Handler(controllersv1.BentoController.ListModel, 200))

	resourceGrp.GET("/deployments", []fizz.OperationOption{
		fizz.ID("List bento deployments"),
		fizz.Summary("List bento deployments"),
	}, tonic.Handler(controllersv1.BentoController.ListDeployment, 200))

	resourceGrp.PATCH("/presign_upload_url", []fizz.OperationOption{
		fizz.ID("Pre sign bento upload URL"),
		fizz.Summary("Pre sign bento upload URL"),
	}, tonic.Handler(controllersv1.BentoController.PreSignUploadUrl, 200))

	resourceGrp.PATCH("/presign_download_url", []fizz.OperationOption{
		fizz.ID("Pre sign bento download URL"),
		fizz.Summary("Pre sign bento download URL"),
	}, tonic.Handler(controllersv1.BentoController.PreSignDownloadUrl, 200))

	resourceGrp.PATCH("/start_upload", []fizz.OperationOption{
		fizz.ID("Start upload a bento"),
		fizz.Summary("Start upload a bento"),
	}, tonic.Handler(controllersv1.BentoController.StartUpload, 200))

	resourceGrp.PATCH("/finish_upload", []fizz.OperationOption{
		fizz.ID("Finish upload a bento"),
		fizz.Summary("Finish upload a bento"),
	}, tonic.Handler(controllersv1.BentoController.FinishUpload, 200))

	resourceGrp.GET("/image_builder_pods", []fizz.OperationOption{
		fizz.ID("List bento image builder pods"),
		fizz.Summary("List bento image builder pods"),
	}, tonic.Handler(controllersv1.BentoController.ListImageBuilderPods, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List bentos"),
		fizz.Summary("List bentos"),
	}, tonic.Handler(controllersv1.BentoController.List, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create a bento"),
		fizz.Summary("Create a bento"),
	}, tonic.Handler(controllersv1.BentoController.Create, 200))
}

func deploymentRoutes(grp *fizz.RouterGroup) {
	namespacedGrp := grp.Group("/namespaces/:kubeNamespace/deployments", "deployments", "deployments")
	grp = grp.Group("/deployments", "cluster deployments", "cluster deployments")

	resourceGrp := namespacedGrp.Group("/:deploymentName", "deployment resource", "deployment resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a deployment"),
		fizz.Summary("Get a deployment"),
	}, tonic.Handler(controllersv1.DeploymentController.Get, 200))

	resourceGrp.PATCH("", []fizz.OperationOption{
		fizz.ID("Update a deployment"),
		fizz.Summary("Update a deployment"),
	}, tonic.Handler(controllersv1.DeploymentController.Update, 200))

	resourceGrp.POST("/sync_status", []fizz.OperationOption{
		fizz.ID("Sync a deployment status"),
		fizz.Summary("Sync a deployment status"),
	}, tonic.Handler(controllersv1.DeploymentController.SyncStatus, 200))

	resourceGrp.POST("/terminate", []fizz.OperationOption{
		fizz.ID("Terminate a deployment"),
		fizz.Summary("Terminate a deployment"),
	}, tonic.Handler(controllersv1.DeploymentController.Terminate, 200))

	resourceGrp.DELETE("", []fizz.OperationOption{
		fizz.ID("Delete a deployment"),
		fizz.Summary("Delete a deployment"),
	}, tonic.Handler(controllersv1.DeploymentController.Delete, 200))

	resourceGrp.GET("/terminal_records", []fizz.OperationOption{
		fizz.ID("List deployment terminal records"),
		fizz.Summary("List deployment terminal records"),
	}, tonic.Handler(controllersv1.DeploymentController.ListTerminalRecords, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List cluster deployments"),
		fizz.Summary("List cluster deployments"),
	}, tonic.Handler(controllersv1.DeploymentController.ListClusterDeployments, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create deployment"),
		fizz.Summary("Create deployment"),
	}, tonic.Handler(controllersv1.DeploymentController.Create, 200))

	deploymentRevisionRoutes(resourceGrp)
}

func deploymentRevisionRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/revisions", "deployment revisions", "deployment revisions")

	resourceGrp := grp.Group("/:revisionUid", "deployment revision resource", "deployment revision resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a deployment revision"),
		fizz.Summary("Get a deployment revision"),
	}, tonic.Handler(controllersv1.DeploymentRevisionController.Get, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List deployment revisions"),
		fizz.Summary("List deployment revisions"),
	}, tonic.Handler(controllersv1.DeploymentRevisionController.List, 200))
}

func terminalRecordRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/terminal_records", "terminal records", "terminal records")

	resourceGrp := grp.Group("/:uid", "terminal record resource", "terminal record resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a terminal record"),
		fizz.Summary("Get a terminal record"),
	}, tonic.Handler(controllersv1.TerminalRecordController.Get, 200))

	resourceGrp.GET("/download", []fizz.OperationOption{
		fizz.ID("Download a terminal record"),
		fizz.Summary("Download a terminal record"),
	}, tonic.Handler(controllersv1.TerminalRecordController.Download, 200))
}

func modelRepositoryRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/model_repositories", "model repositories", "model repositories")

	resourceGrp := grp.Group("/:modelRepositoryName", "model repository resource", "model repository resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a model repository"),
		fizz.Summary("Get a model repository"),
	}, tonic.Handler(controllersv1.ModelRepositoryController.Get, 200))

	resourceGrp.PATCH("", []fizz.OperationOption{
		fizz.ID("Update a model repository"),
		fizz.Summary("Update a model repository"),
	}, tonic.Handler(controllersv1.ModelRepositoryController.Update, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List model repositories"),
		fizz.Summary("List model repositories"),
	}, tonic.Handler(controllersv1.ModelRepositoryController.List, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create a model repository"),
		fizz.Summary("Create a model repository"),
	}, tonic.Handler(controllersv1.ModelRepositoryController.Create, 200))

	modelRoutes(resourceGrp)
}

func modelRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/models", "models", "models")

	resourceGrp := grp.Group("/:version", "model resource", "model resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a model"),
		fizz.Summary("Get a model"),
	}, tonic.Handler(controllersv1.ModelController.Get, 200))

	resourceGrp.PATCH("", []fizz.OperationOption{
		fizz.ID("Update a model"),
		fizz.Summary("Update a model"),
	}, tonic.Handler(controllersv1.ModelController.Update, 200))

	resourceGrp.GET("/bentos", []fizz.OperationOption{
		fizz.ID("List model bentos"),
		fizz.Summary("List model bentos"),
	}, tonic.Handler(controllersv1.ModelController.ListBento, 200))

	resourceGrp.GET("/deployments", []fizz.OperationOption{
		fizz.ID("List model deployments"),
		fizz.Summary("List model deployments"),
	}, tonic.Handler(controllersv1.ModelController.ListDeployment, 200))

	resourceGrp.PATCH("/presign_upload_url", []fizz.OperationOption{
		fizz.ID("Pre sign model upload URL"),
		fizz.Summary("Pre sign model upload URL"),
	}, tonic.Handler(controllersv1.ModelController.PreSignUploadUrl, 200))

	resourceGrp.PATCH("/presign_download_url", []fizz.OperationOption{
		fizz.ID("Pre sign model download URL"),
		fizz.Summary("Pre sign model download URL"),
	}, tonic.Handler(controllersv1.ModelController.PreSignDownloadUrl, 200))

	resourceGrp.PATCH("/start_upload", []fizz.OperationOption{
		fizz.ID("Start upload a model"),
		fizz.Summary("Start upload a model"),
	}, tonic.Handler(controllersv1.ModelController.StartUpload, 200))

	resourceGrp.PATCH("/finish_upload", []fizz.OperationOption{
		fizz.ID("Finish upload a model"),
		fizz.Summary("Finish upload a model"),
	}, tonic.Handler(controllersv1.ModelController.FinishUpload, 200))

	resourceGrp.GET("/image_builder_pods", []fizz.OperationOption{
		fizz.ID("List model image builder pods"),
		fizz.Summary("List model image builder pods"),
	}, tonic.Handler(controllersv1.ModelController.ListImageBuilderPods, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List models"),
		fizz.Summary("List models"),
	}, tonic.Handler(controllersv1.ModelController.List, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create a model"),
		fizz.Summary("Create a model"),
	}, tonic.Handler(controllersv1.ModelController.Create, 200))
}

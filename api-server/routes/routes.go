package routes

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/bentoml/yatai/api-server/controllers/web"

	"github.com/bentoml/yatai/api-server/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/pkg/errors"
	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"

	"github.com/bentoml/yatai/api-server/controllers/controllersv1"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/scookie"
	"github.com/bentoml/yatai/common/yataicontext"
	"github.com/bentoml/yatai/schemas/schemasv1"
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

func NewRouter() (*fizz.Fizz, error) {
	engine := gin.New()

	store := cookie.NewStore([]byte(config.YataiConfig.Server.SessionSecretKey))
	engine.Use(sessions.Sessions("yatai-session-v1", store))

	oauthGrp := engine.Group("oauth")
	oauthGrp.GET("/github", controllersv1.GithubOAuthLogin)

	callbackGrp := engine.Group("callback")
	callbackGrp.GET("/github", controllersv1.GithubOAuthCallBack)

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
	wsRootGroup.GET("/subscription/resource", []fizz.OperationOption{
		fizz.ID("Subscribe resource"),
		fizz.Summary("Subscribe resource"),
	}, requireLogin, tonic.Handler(controllersv1.SubscriptionController.SubscribeResource, 200))

	wsRootGroup.GET("/orgs/:orgName/clusters/:clusterName/deployments/:deploymentName/tail", []fizz.OperationOption{
		fizz.ID("Tail pods log"),
		fizz.Summary("Tail pods log"),
	}, requireLogin, tonic.Handler(controllersv1.LogController.TailPodsLog, 200))

	wsRootGroup.GET("/orgs/:orgName/clusters/:clusterName/deployments/:deploymentName/terminal", []fizz.OperationOption{
		fizz.ID("Deployment terminal"),
		fizz.Summary("Deployment terminal"),
	}, requireLogin, tonic.Handler(controllersv1.TerminalController.GetDeploymentTerminal, 200))

	wsRootGroup.GET("/orgs/:orgName/clusters/:clusterName/deployments/:deploymentName/kube_events", []fizz.OperationOption{
		fizz.ID("Deployment kube events"),
		fizz.Summary("Deployment kube events"),
	}, requireLogin, tonic.Handler(controllersv1.KubeController.GetDeploymentKubeEvents, 200))

	wsRootGroup.GET("/orgs/:orgName/clusters/:clusterName/deployments/:deploymentName/pods", []fizz.OperationOption{
		fizz.ID("Ws pods"),
		fizz.Summary("Ws pods"),
	}, requireLogin, tonic.Handler(controllersv1.DeploymentController.WsPods, 200))

	ginGroup := engine.Group("/api/v1/orgs/:orgName/clusters/:clusterName")

	ginGroup.GET("/grafana/*path", requireLogin, controllersv1.GrafanaController.Proxy)
	ginGroup.POST("/grafana/*path", requireLogin, controllersv1.GrafanaController.Proxy)
	ginGroup.PUT("/grafana/*path", requireLogin, controllersv1.GrafanaController.Proxy)
	ginGroup.PATCH("/grafana/*path", requireLogin, controllersv1.GrafanaController.Proxy)
	ginGroup.HEAD("/grafana/*path", requireLogin, controllersv1.GrafanaController.Proxy)
	ginGroup.DELETE("/grafana/*path", requireLogin, controllersv1.GrafanaController.Proxy)

	apiRootGroup := fizzApp.Group("/api/v1", "api v1", "api v1")

	// Setup routes.
	apiRootGroup.GET("/yatai_component_operator_helm_charts", []fizz.OperationOption{
		fizz.ID("List yatai component operator helm charts"),
		fizz.Summary("List yatai component operator helm charts"),
	}, requireLogin, tonic.Handler(controllersv1.YataiComponentController.ListOperatorHelmCharts, 200))

	authRoutes(apiRootGroup)
	userRoutes(apiRootGroup)
	organizationRoutes(apiRootGroup)
	terminalRecordRoutes(apiRootGroup)

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
	apiToken := ctx.GetHeader(consts.YataiApiTokenHeaderName)

	// nolint: gocritic
	if apiToken != "" {
		user, err = services.UserService.GetByApiToken(ctx, apiToken)
		if err != nil {
			err = errors.Wrap(err, "get user by api token")
			return
		}
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
	ctx.Next()
}

func authRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/auth", "auth", "auth api")

	grp.POST("/register", []fizz.OperationOption{
		fizz.ID("Register an user"),
		fizz.Summary("Register an user"),
	}, tonic.Handler(controllersv1.AuthController.Register, 200))

	grp.POST("/login", []fizz.OperationOption{
		fizz.ID("Login an user"),
		fizz.Summary("Login an user"),
	}, tonic.Handler(controllersv1.AuthController.Login, 200))

	grp.GET("/current", []fizz.OperationOption{
		fizz.ID("Get current user"),
		fizz.Summary("Get current user"),
	}, requireLogin, tonic.Handler(controllersv1.AuthController.GetCurrentUser, 200))

	grp.PUT("/current/api_token", []fizz.OperationOption{
		fizz.ID("Generate current user api_token"),
	}, requireLogin, tonic.Handler(controllersv1.AuthController.GenerateApiToken, 200))

	grp.DELETE("/current/api_token", []fizz.OperationOption{
		fizz.ID("Delete current user api_token"),
	}, requireLogin, tonic.Handler(controllersv1.AuthController.DeleteApiToken, 200))
}

func userRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/users", "users", "users api")

	resourceGrp := grp.Group("/:userName", "user resource", "user resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get an user"),
		fizz.Summary("Get an user"),
	}, requireLogin, tonic.Handler(controllersv1.UserController.Get, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List users"),
		fizz.Summary("List users"),
	}, requireLogin, tonic.Handler(controllersv1.UserController.List, 200))
}

func organizationRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/orgs", "organizations", "organizations")

	resourceGrp := grp.Group("/:orgName", "organization resource", "organization resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get an organization"),
		fizz.Summary("Get an organization"),
	}, requireLogin, tonic.Handler(controllersv1.OrganizationController.Get, 200))

	resourceGrp.PATCH("", []fizz.OperationOption{
		fizz.ID("Update an organization"),
		fizz.Summary("Update an organization"),
	}, requireLogin, tonic.Handler(controllersv1.OrganizationController.Update, 200))

	resourceGrp.GET("/members", []fizz.OperationOption{
		fizz.ID("List organization members"),
		fizz.Summary("Get organization members"),
	}, requireLogin, tonic.Handler(controllersv1.OrganizationMemberController.List, 200))

	resourceGrp.POST("/members", []fizz.OperationOption{
		fizz.ID("Create an organization member"),
		fizz.Summary("Create an organization member"),
	}, requireLogin, tonic.Handler(controllersv1.OrganizationMemberController.Create, 200))

	resourceGrp.DELETE("/members", []fizz.OperationOption{
		fizz.ID("Remove an organization member"),
		fizz.Summary("Remove an organization member"),
	}, requireLogin, tonic.Handler(controllersv1.OrganizationMemberController.Delete, 200))

	resourceGrp.GET("/deployments", []fizz.OperationOption{
		fizz.ID("List organization deployments"),
		fizz.Summary("List organization deployments"),
	}, requireLogin, tonic.Handler(controllersv1.DeploymentController.ListOrganizationDeployments, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List organizations"),
		fizz.Summary("List organizations"),
	}, requireLogin, tonic.Handler(controllersv1.OrganizationController.List, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create organization"),
		fizz.Summary("Create organization"),
	}, requireLogin, tonic.Handler(controllersv1.OrganizationController.Create, 200))

	clusterRoutes(resourceGrp)
	bentoRoutes(resourceGrp)
}

func clusterRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/clusters", "clusters", "clusters")

	resourceGrp := grp.Group("/:clusterName", "cluster resource", "cluster resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a cluster"),
		fizz.Summary("Get a cluster"),
	}, requireLogin, tonic.Handler(controllersv1.ClusterController.Get, 200))

	resourceGrp.PATCH("", []fizz.OperationOption{
		fizz.ID("Update a cluster"),
		fizz.Summary("Update a cluster"),
	}, requireLogin, tonic.Handler(controllersv1.ClusterController.Update, 200))

	resourceGrp.GET("/members", []fizz.OperationOption{
		fizz.ID("List cluster members"),
		fizz.Summary("List cluster members"),
	}, requireLogin, tonic.Handler(controllersv1.ClusterMemberController.List, 200))

	resourceGrp.POST("/members", []fizz.OperationOption{
		fizz.ID("Create a cluster member"),
		fizz.Summary("Create a cluster member"),
	}, requireLogin, tonic.Handler(controllersv1.ClusterMemberController.Create, 200))

	resourceGrp.DELETE("/members", []fizz.OperationOption{
		fizz.ID("Remove a cluster member"),
		fizz.Summary("Remove a cluster member"),
	}, requireLogin, tonic.Handler(controllersv1.ClusterMemberController.Delete, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List clusters"),
		fizz.Summary("List clusters"),
	}, requireLogin, tonic.Handler(controllersv1.ClusterController.List, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create cluster"),
		fizz.Summary("Create cluster"),
	}, requireLogin, tonic.Handler(controllersv1.ClusterController.Create, 200))

	deploymentRoutes(resourceGrp)
	yataiComponentRoutes(resourceGrp)
}

func yataiComponentRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/yatai_components", "yatai components", "yatai components")

	resourceGrp := grp.Group("/:componentType", "yatai component resource", "yatai component resource")

	resourceGrp.DELETE("", []fizz.OperationOption{
		fizz.ID("Delete a yatai component"),
		fizz.Summary("Delete a yatai component"),
	}, requireLogin, tonic.Handler(controllersv1.YataiComponentController.Delete, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List yatai components"),
		fizz.Summary("List yatai components"),
	}, requireLogin, tonic.Handler(controllersv1.YataiComponentController.List, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create yatai component"),
		fizz.Summary("Create yatai component"),
	}, requireLogin, tonic.Handler(controllersv1.YataiComponentController.Create, 200))
}

func bentoRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/bentos", "bentos", "bentos")

	resourceGrp := grp.Group("/:bentoName", "bento resource", "bento resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a bento"),
		fizz.Summary("Get a bento"),
	}, requireLogin, tonic.Handler(controllersv1.BentoController.Get, 200))

	resourceGrp.PATCH("", []fizz.OperationOption{
		fizz.ID("Update a bento"),
		fizz.Summary("Update a bento"),
	}, requireLogin, tonic.Handler(controllersv1.BentoController.Update, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List bentos"),
		fizz.Summary("List bentos"),
	}, requireLogin, tonic.Handler(controllersv1.BentoController.List, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create bento"),
		fizz.Summary("Create bento"),
	}, requireLogin, tonic.Handler(controllersv1.BentoController.Create, 200))

	bentoVersionRoutes(resourceGrp)
}

func bentoVersionRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/versions", "bento versions", "bento versions")

	resourceGrp := grp.Group("/:version", "bento version resource", "bento version resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a bento version"),
		fizz.Summary("Get a bento version"),
	}, requireLogin, tonic.Handler(controllersv1.BentoVersionController.Get, 200))

	resourceGrp.PATCH("/presign_s3_upload_url", []fizz.OperationOption{
		fizz.ID("Pre sign bento version s3 upload URL"),
		fizz.Summary("Pre sign bento version s3 upload URL"),
	}, requireLogin, tonic.Handler(controllersv1.BentoVersionController.PreSignS3UploadUrl, 200))

	resourceGrp.PATCH("/start_upload", []fizz.OperationOption{
		fizz.ID("Start upload a bento version"),
		fizz.Summary("Start upload a bento version"),
	}, requireLogin, tonic.Handler(controllersv1.BentoVersionController.StartUpload, 200))

	resourceGrp.PATCH("/finish_upload", []fizz.OperationOption{
		fizz.ID("Finish upload a bento version"),
		fizz.Summary("Finish upload a bento version"),
	}, requireLogin, tonic.Handler(controllersv1.BentoVersionController.FinishUpload, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List bento versions"),
		fizz.Summary("List bento versions"),
	}, requireLogin, tonic.Handler(controllersv1.BentoVersionController.List, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create bento version"),
		fizz.Summary("Create bento version"),
	}, requireLogin, tonic.Handler(controllersv1.BentoVersionController.Create, 200))
}

func deploymentRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/deployments", "deployments", "deployments")

	resourceGrp := grp.Group("/:deploymentName", "deployment resource", "deployment resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a deployment"),
		fizz.Summary("Get a deployment"),
	}, requireLogin, tonic.Handler(controllersv1.DeploymentController.Get, 200))

	resourceGrp.PATCH("", []fizz.OperationOption{
		fizz.ID("Update a deployment"),
		fizz.Summary("Update a deployment"),
	}, requireLogin, tonic.Handler(controllersv1.DeploymentController.Update, 200))

	resourceGrp.GET("/terminal_records", []fizz.OperationOption{
		fizz.ID("List deployment terminal records"),
		fizz.Summary("List deployment terminal records"),
	}, requireLogin, tonic.Handler(controllersv1.DeploymentController.ListTerminalRecords, 200))

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List deployments"),
		fizz.Summary("List deployments"),
	}, requireLogin, tonic.Handler(controllersv1.DeploymentController.ListClusterDeployments, 200))

	grp.POST("", []fizz.OperationOption{
		fizz.ID("Create deployment"),
		fizz.Summary("Create deployment"),
	}, requireLogin, tonic.Handler(controllersv1.DeploymentController.Create, 200))

	deploymentSnapshotRoutes(resourceGrp)
}

func deploymentSnapshotRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/snapshots", "deployment snapshots", "deployment snapshots")

	grp.GET("", []fizz.OperationOption{
		fizz.ID("List deployment snapshots"),
		fizz.Summary("List deployment snapshots"),
	}, requireLogin, tonic.Handler(controllersv1.DeploymentSnapshotController.List, 200))
}

func terminalRecordRoutes(grp *fizz.RouterGroup) {
	grp = grp.Group("/terminal_records", "terminal records", "terminal records")

	resourceGrp := grp.Group("/:uid", "terminal record resource", "terminal record resource")

	resourceGrp.GET("", []fizz.OperationOption{
		fizz.ID("Get a terminal record"),
		fizz.Summary("Get a terminal record"),
	}, requireLogin, tonic.Handler(controllersv1.TerminalRecordController.Get, 200))

	resourceGrp.GET("/download", []fizz.OperationOption{
		fizz.ID("Download a terminal record"),
		fizz.Summary("Download a terminal record"),
	}, requireLogin, tonic.Handler(controllersv1.TerminalRecordController.Download, 200))
}

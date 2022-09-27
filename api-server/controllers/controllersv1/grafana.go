package controllersv1

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/huandu/xstrings"

	"github.com/bentoml/grafana-operator/api/integreatly/v1alpha1"

	commonconsts "github.com/bentoml/yatai-common/consts"
	"github.com/bentoml/yatai/api-server/config"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/common/reqcli"
)

type grafanaController struct {
	clusterController
}

var GrafanaController = grafanaController{}

type ProxyGrafanaSchema struct {
	GetClusterSchema
	Path string `path:"path"`
}

var (
	staticSuffixes    = []string{"js", "css", "svg", "png", "woff2"}
	pathPrefixPattern = regexp.MustCompile("^/")
	clustersCache     sync.Map
)

func getGrafanaCacheKey(orgName, clusterName string) string {
	return fmt.Sprintf("grafana:%s:%s", orgName, clusterName)
}

func (c *grafanaController) Proxy(ctx *gin.Context) {
	schema := &ProxyGrafanaSchema{
		GetClusterSchema: GetClusterSchema{
			GetOrganizationSchema: GetOrganizationSchema{
				OrgName: ctx.Param("orgName"),
			},
			ClusterName: ctx.Param("clusterName"),
		},
		Path: ctx.Param("path"),
	}

	_, _, suffix := xstrings.LastPartition(schema.Path, ".")
	suffix = strings.ToLower(suffix)
	isStatic := false
	for _, s := range staticSuffixes {
		if s == suffix {
			isStatic = true
			break
		}
	}

	var cluster *models.Cluster
	var err error

	clusterCacheKey := fmt.Sprintf("cluster:%s:%s", schema.OrgName, schema.ClusterName)
	cluster_, ok := clustersCache.Load(clusterCacheKey)
	if !ok {
		cluster, err = schema.GetCluster(ctx)
		if err != nil {
			_ = ctx.AbortWithError(500, err)
			return
		}
		clustersCache.Store(clusterCacheKey, cluster)
	} else {
		cluster = cluster_.(*models.Cluster)
	}

	if !isStatic {
		if err = ClusterController.canView(ctx, cluster); err != nil {
			_ = ctx.AbortWithError(400, err)
			return
		}
	}

	grafanaCacheKey := getGrafanaCacheKey(schema.OrgName, schema.ClusterName)
	grafana := &v1alpha1.Grafana{}
	exists, err := services.CacheService.Get(ctx, grafanaCacheKey, grafana)
	if err != nil {
		_ = ctx.AbortWithError(500, err)
		return
	}
	if !exists {
		grafana, err = services.ClusterService.GetGrafana(ctx, cluster)
		if err != nil {
			_ = ctx.AbortWithError(500, err)
			return
		}

		var org *models.Organization
		org, err = services.OrganizationService.GetAssociatedOrganization(ctx, cluster)
		if err != nil {
			return
		}
		var majorCluster *models.Cluster
		majorCluster, err = services.OrganizationService.GetMajorCluster(ctx, org)
		if err != nil {
			return
		}

		if majorCluster.ID == cluster.ID && config.YataiConfig.InCluster && !config.YataiConfig.IsSaaS {
			grafana.Spec.Ingress.Hostname = fmt.Sprintf("yatai-grafana.%s", commonconsts.KubeNamespaceYataiComponents)
		}

		err = services.CacheService.Set(ctx, grafanaCacheKey, grafana)
		if err != nil {
			_ = ctx.AbortWithError(500, err)
			return
		}
	}

	grafanaHostname := grafana.Spec.Ingress.Hostname

	path := fmt.Sprintf("/%s", pathPrefixPattern.ReplaceAllString(schema.Path, ""))

	oldReq := ctx.Request
	oldUrl := oldReq.URL

	url_ := oldUrl
	url_.Scheme = "http"
	url_.Host = grafanaHostname
	url_.Path = path

	req := &http.Request{
		Method:        oldReq.Method,
		URL:           url_,
		Proto:         oldReq.Proto,
		Body:          oldReq.Body,
		Header:        oldReq.Header,
		Form:          oldReq.Form,
		PostForm:      oldReq.PostForm,
		MultipartForm: oldReq.MultipartForm,
	}

	req.SetBasicAuth(grafana.Spec.Config.Security.AdminUser, grafana.Spec.Config.Security.AdminPassword)

	cli := reqcli.GetDefaultHttpClient()

	resp, err := cli.Do(req)
	if err != nil {
		_ = ctx.AbortWithError(resp.StatusCode, err)
		return
	}
	defer resp.Body.Close()

	err = writeProxyResp(ctx.Writer, resp)
	if err != nil {
		_ = ctx.AbortWithError(400, err)
	}
}

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func writeProxyResp(w http.ResponseWriter, resp *http.Response) error {
	for _, h := range hopHeaders {
		resp.Header.Del(h)
	}
	header := w.Header()
	for k, vs := range resp.Header {
		for _, v := range vs {
			header.Set(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err := io.Copy(w, resp.Body)
	return err
}

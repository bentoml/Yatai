package controllersv1

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/bentoml/grafana-operator/api/integreatly/v1alpha1"

	"github.com/bentoml/yatai/api-server/models"

	"github.com/gin-gonic/gin"
	"github.com/huandu/xstrings"

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
	grafanasCache     sync.Map
	clustersCache     sync.Map
)

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

	clusterCacheKey := fmt.Sprintf("%s:%s", schema.OrgName, schema.ClusterName)
	cluster_, ok := clustersCache.Load(clusterCacheKey)
	if !ok {
		cluster, err = schema.GetCluster(ctx)
		if err != nil {
			_ = ctx.AbortWithError(400, err)
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

	var grafana *v1alpha1.Grafana
	grafana_, ok := grafanasCache.Load(cluster.ID)
	if !ok {
		grafana, err = services.ClusterService.GetGrafana(ctx, cluster)
		if err != nil {
			_ = ctx.AbortWithError(400, err)
			return
		}
		grafanasCache.Store(cluster.ID, grafana)
	} else {
		grafana = grafana_.(*v1alpha1.Grafana)
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

	if !isStatic {
	}
	resp, err := cli.Do(req)
	if err != nil {
		_ = ctx.AbortWithError(resp.StatusCode, err)
		return
	}
	defer resp.Body.Close()

	//mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	//if err != nil {
	//	return err
	//}
	//
	//enableGzip := strings.Contains(req.Header.Get("Accept-Encoding"), "gzip")
	//
	//if mediaType == "text/html" {
	//	return writeHTMLProxyResp(schema, ctx.Writer, resp, enableGzip)
	//}
	err = writeProxyResp(ctx.Writer, resp)
	if err != nil {
		_ = ctx.AbortWithError(400, err)
	}
	return
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

func writeHTMLProxyResp(schema *ProxyGrafanaSchema, w http.ResponseWriter, resp *http.Response, enableGzip bool) error {
	htmlBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if enableGzip {
		gzReader, err := gzip.NewReader(bytes.NewReader(htmlBytes))
		if err != nil {
			return err
		}
		htmlBytes, err = ioutil.ReadAll(gzReader)
		if err != nil {
			return err
		}
	}
	htmlBytes = bytes.Replace(htmlBytes, []byte(`<base href="/" />`), []byte(fmt.Sprintf(`<base href="/api/v1/orgs/%s/clusters/%s/grafana/" />`, schema.OrgName, schema.ClusterName)), -1)
	htmlBytes = bytes.Replace(htmlBytes, []byte(`"gravatarUrl":"/avatar/`), []byte(fmt.Sprintf(`"grafanaUrl":"/api/v1/orgs/%s/clusters/%s/grafana/avatar/`, schema.OrgName, schema.ClusterName)), -1)
	//doc, err := goquery.NewDocumentFromReader(resp.Body)
	//if err != nil {
	//	return err
	//}
	//
	//doc.Find("base").SetAttr("href", fmt.Sprintf(`/api/v1/orgs/%s/clusters/%s/grafana/`, schema.OrgName, schema.ClusterName))
	//
	//html, err := doc.Html()
	//if err != nil {
	//	return err
	//}
	//
	//htmlBytes := []byte(html)

	for _, h := range hopHeaders {
		resp.Header.Del(h)
	}

	header := w.Header()
	for k, vs := range resp.Header {
		for _, v := range vs {
			header.Set(k, v)
		}
	}

	//contentLength := strconv.Itoa(len(htmlBytes))
	//header.Set("Content-Length", contentLength)
	header.Del("Content-Length")
	w.WriteHeader(resp.StatusCode)

	if enableGzip {
		gzWriter := gzip.NewWriter(w)
		defer gzWriter.Close()
		w = gzipResponseWriter{Writer: gzWriter, ResponseWriter: w}
	}

	written, err := io.Copy(w, bytes.NewReader(htmlBytes))
	fmt.Println(written)
	return err
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

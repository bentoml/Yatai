package controllersv1

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

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
)

func (c *grafanaController) Proxy(ctx *gin.Context, schema *ProxyGrafanaSchema) error {
	_, _, suffix := xstrings.LastPartition(schema.Path, ".")
	suffix = strings.ToLower(suffix)
	isStatic := false
	for _, s := range staticSuffixes {
		if s == suffix {
			isStatic = true
			break
		}
	}

	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return err
	}
	if err = ClusterController.canView(ctx, cluster); err != nil {
		return err
	}

	grafana, err := services.ClusterService.GetGrafana(ctx, cluster)
	if err != nil {
		return err
	}

	grafanaHostname := grafana.Spec.Ingress.Hostname

	path := fmt.Sprintf("/%s", pathPrefixPattern.ReplaceAllString(schema.Path, ""))

	oldReq := ctx.Request
	oldUrl := oldReq.URL

	url_ := &url.URL{
		Scheme:      "http",
		Host:        grafanaHostname,
		Path:        path,
		ForceQuery:  oldUrl.ForceQuery,
		RawQuery:    oldUrl.RawQuery,
		Fragment:    oldUrl.Fragment,
		RawFragment: oldUrl.RawFragment,
	}

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
		return err
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
	return writeProxyResp(ctx.Writer, resp)
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

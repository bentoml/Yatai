package tracing

import (
	"context"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/common/yataicontext"
)

func GetSpan(ctx context.Context) opentracing.Span {
	sp_ := ctx.Value(consts.TracingContextKey)
	if sp_ == nil {
		return nil
	}
	if sp, ok := sp_.(opentracing.Span); ok {
		return sp
	}
	return nil
}

func SetSpan(ctx *gin.Context, span opentracing.Span) {
	ctx.Set(consts.TracingContextKey, span)
}

func startSpan(ctx context.Context, operationName string, options_ ...opentracing.StartSpanOption) opentracing.Span {
	options := append([]opentracing.StartSpanOption{
		opentracing.Tag{Key: "goroutine", Value: runtime.NumGoroutine()},
		opentracing.Tag{Key: "file", Value: utils.FileWithLineNum()},
		opentracing.Tag{Key: "user", Value: yataicontext.GetUserName(ctx)},
	}, options_...)

	return opentracing.StartSpan(operationName, options...)
}

func StartSpan(ctx context.Context, operationName string) (context.Context, opentracing.Span) {
	var options []opentracing.StartSpanOption

	parentSpan := GetSpan(ctx)
	if parentSpan != nil {
		options = append(options, opentracing.ChildOf(parentSpan.Context()))
	}

	span := startSpan(ctx, operationName, options...)
	// nolint: ineffassign,staticcheck
	return context.WithValue(ctx, consts.TracingContextKey, span), span
}

func StartSpanWithParent(ctx context.Context, parent opentracing.SpanContext, operationName, method, path string) opentracing.Span {
	options := []opentracing.StartSpanOption{
		opentracing.Tag{Key: ext.SpanKindRPCServer.Key, Value: ext.SpanKindRPCServer.Value},
		opentracing.Tag{Key: string(ext.HTTPMethod), Value: method},
		opentracing.Tag{Key: string(ext.HTTPUrl), Value: path},
	}

	if parent != nil {
		options = append(options, opentracing.ChildOf(parent))
	}

	return startSpan(ctx, operationName, options...)
}

func StartSpanWithHeader(ctx context.Context, header *http.Header, operationName, method, path string) opentracing.Span {
	var wireContext opentracing.SpanContext
	if header != nil {
		wireContext, _ = opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(*header))
	}

	return StartSpanWithParent(ctx, wireContext, operationName, method, path)
}

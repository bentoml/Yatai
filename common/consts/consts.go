package consts

const (
	DefaultNewsURL                         = "https://raw.githubusercontent.com/bentoml/yatai-homepage-news/main/news.json"
	DefaultETCDTimeoutSeconds              = 5
	DefaultETCDDialKeepaliveTimeSeconds    = 30
	DefaultETCDDialKeepaliveTimeoutSeconds = 10

	AppNameMaxLength     = 20
	AppCompNameMaxLength = 20
	AppCompMaxReplicas   = 10

	YataiDebugImg             = "yatai.ai/yatai-infras/debug"
	YataiKubectlNamespace     = "default"
	YataiKubectlContainerName = "main"
	YataiKubectlImage         = "yatai.ai/yatai-infras/k8s"

	TracingContextKey = "tracing-context"
	// nolint: gosec
	YataiApiTokenHeaderName = "X-YATAI-API-TOKEN"

	BentoServicePort       = 3000
	BentoServicePortEnvKey = "PORT"

	NoneStr = "None"

	AmazonS3Endpoint = "s3.amazonaws.com"

	// nolint: gosec
	YataiK8sBotApiTokenName = "yatai-k8s-bot"
)

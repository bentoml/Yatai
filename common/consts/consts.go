package consts

const (
	DefaultETCDTimeoutSeconds              = 5
	DefaultETCDDialKeepaliveTimeSeconds    = 30
	DefaultETCDDialKeepaliveTimeoutSeconds = 10

	AppNameMaxLength     = 20
	AppCompNameMaxLength = 20
	AppCompMaxReplicas   = 10
)

const (
	TracingContextKey = "tracing-context"
	// nolint: gosec
	YataiApiTokenHeaderName = "X-YATAI-API-TOKEN"

	BentoServicePort = 5000

	NoneStr = "None"
)

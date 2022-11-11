package tracking

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bentoml/yatai-common/reqcli"
)

const (
	TRACKING_SERVER         = "http://t.bentoml.com"
	TRACKER_TIMEOUT         = time.Duration(1) * time.Second
	YATAI_TRACKING_LOGLEVEL = "__YATAI_TRACKING_LOGLEVEL"
	YATAI_DONOT_TRACK       = "YATAI_DONOT_TRACK"
)

var trackingLogger = NewTrackerLogger()

func NewTrackerLogger() *log.Logger {
	out := os.Getenv(YATAI_TRACKING_LOGLEVEL)
	var logLevel log.Level
	switch strings.ToLower(out) {
	case "info":
		logLevel = log.InfoLevel
	case "debug":
		logLevel = log.DebugLevel
	default:
		logLevel = log.FatalLevel
	}
	logger := log.New()
	logger.SetLevel(logLevel)
	return logger
}

func doNotTrack() bool {
	out := os.Getenv(YATAI_DONOT_TRACK)
	return strings.ToLower(out) == "true"
}

func isTrackingDebug() bool {
	out := os.Getenv(YATAI_TRACKING_LOGLEVEL)
	return out == "debug"
}

// Marshal the data and sent to tracking server
func track(ctx context.Context, data interface{}, eventType YataiEventType) {
	trackingLogger := trackingLogger.WithField("eventType", eventType)

	jsonData, err := json.Marshal(data)
	if err != nil {
		trackingLogger.Error(err, "Failed to marshal data")
		return
	}
	var prettyJSON bytes.Buffer
	_ = json.Indent(&prettyJSON, jsonData, "", " ")
	trackingLogger.Info("Tracking Payload: ", prettyJSON.String())

	if !doNotTrack() {
		type JitsuResponse struct {
			Status string `json:"status"`
		}
		var resp JitsuResponse
		_, err := reqcli.NewJsonRequestBuilder().Method("POST").Url(TRACKING_SERVER).Payload(bytes.NewBuffer(jsonData)).Result(&resp).Do(ctx)
		if err != nil {
			trackingLogger.Error(err, "sending request failed.")
		}

		if resp.Status == "ok" {
			trackingLogger.Info("Tracking Request sent.")
		} else {
			trackingLogger.Errorf("Tracking Request failed. Status [%s]", resp.Status)
		}
	}
}

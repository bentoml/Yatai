package tracking

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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

func donot_track() bool {
	out := os.Getenv(YATAI_DONOT_TRACK)
	return strings.ToLower(out) == "true"
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

	if !donot_track() {
		client := http.Client{Timeout: TRACKER_TIMEOUT}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, TRACKING_SERVER, bytes.NewBuffer(jsonData))
		if err != nil {
			trackingLogger.Error(err, "failed to create new request.")
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			trackingLogger.Error(err, "sending request failed.")
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			trackingLogger.Info("Tracking Request sent.")
		} else {
			trackingLogger.Errorf("Tracking Request failed. Status [%s]", resp.Status)
			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyString := string(bodyBytes)
			trackingLogger.Error(bodyString)
		}
	}
}

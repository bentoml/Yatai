package tracking

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	TRACKING_SERVER         = "http://t.bentoml.com"
	YATAI_TRACKING_LOGLEVEL = "__YATAI_TRACKING_LOGLEVEL"
	YATAI_DONOT_TRACK       = "YATAI_DONOT_TRACK"
)

var trackerLog = NewTrackerLogger()

func NewTrackerLogger() *log.Logger {
	out := os.Getenv(YATAI_TRACKING_LOGLEVEL)
	var logLevel log.Level
	if strings.ToLower(out) == "info" {
		logLevel = log.InfoLevel
	} else if strings.ToLower(out) == "debug" {
		logLevel = log.DebugLevel
	} else {
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
func track(data interface{}, eventType YataiEventType) {
	trackerLog := trackerLog.WithField("eventType", eventType)

	jsonData, err := json.Marshal(data)
	if err != nil {
		trackerLog.Error(err, "Failed to marshal data")
		return
	}

	var prettyJSON bytes.Buffer
	_ = json.Indent(&prettyJSON, jsonData, "", " ")
	trackerLog.Info("Tracking Payload: ", prettyJSON.String())

	if !donot_track() {
		resp, err := http.Post(TRACKING_SERVER, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			trackerLog.Error(err, "failed to send data to tracking server.")
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			trackerLog.Info("Tracking Request sent.")
		} else {
			trackerLog.Errorf("Tracking Request failed. Status [%s]", resp.Status)
			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyString := string(bodyBytes)
			trackerLog.Error(bodyString)
		}
	}
}

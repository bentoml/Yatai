package tracking

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	JITSU_SERVER              = "http://localhost:8000/api/v1/s2s/event"
	JITSUE_KEY                = "s2s.l5z3kpmcuevukrepeg70ht.j2imnduhnssuh4yb7gt9"
	YATAI_TRACKING_DEBUG_MODE = "__YATAI_TRACKING_DEBUG_MODE"
	YATAI_DONOT_TRACK         = "YATAI_DONOT_TRACK"
)

func is_debug_mode() bool {
	out := os.Getenv(YATAI_TRACKING_DEBUG_MODE)
	return strings.ToLower(out) == "true"
}

func donot_track() bool {
	out := os.Getenv(YATAI_DONOT_TRACK)
	return strings.ToLower(out) == "true"
}

// Marshal the data and sent to tracking server
func track(data interface{}, eventType string) {
	trackerLog := log.WithField("eventType", eventType)

	jsonData, err := json.Marshal(data)
	if err != nil {
		if is_debug_mode() {
			trackerLog.Error(err, "Failed to marshal data")
		}
		return
	}

	if is_debug_mode() {
		var prettyJSON bytes.Buffer
		_ = json.Indent(&prettyJSON, jsonData, "", " ")
		trackerLog.Info("Tracking Payload: ", prettyJSON.String())
	}

	if !donot_track() {
		//TODO: change to t.bentoml.com
		request_url := fmt.Sprintf("%s?token=%s", JITSU_SERVER, JITSUE_KEY)
		resp, err := http.Post(request_url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil && is_debug_mode() {
			trackerLog.Error(err, "failed to send data to tracking server.")
		}
		defer resp.Body.Close()

		if is_debug_mode() {
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
}

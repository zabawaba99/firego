package firego

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
)

// EventTypeError is the type that is set on an Event struct if an
// error occurs while watching a Firebase reference.
const EventTypeError = "event_error"

// Event represents a notification received when watching a
// firebase reference.
type Event struct {
	// Type of event that was received
	Type string
	// Path to the data that changed
	Path string
	// Data that changed
	Data    interface{}
	payload string
}

func (e Event) Value(v interface{}) (path string, err error) {
	var p struct {
		Path string      `json:"path"`
		Data interface{} `json:"data"`
	}
	p.Data = &v
	err = json.Unmarshal([]byte(e.payload), &p)
	if err != nil {
		path = ""
	} else {
		path = p.Path
	}
	return
}

// StopWatching stops tears down all connections that are watching.
func (fb *Firebase) StopWatching() {
	if fb.isWatching() {
		// signal connection to terminal
		fb.stopWatching <- struct{}{}
		// flip the bit back to not watching
		fb.setWatching(false)
	}
}

func (fb *Firebase) isWatching() bool {
	fb.watchMtx.Lock()
	v := fb.watching
	fb.watchMtx.Unlock()
	return v
}

func (fb *Firebase) setWatching(v bool) {
	fb.watchMtx.Lock()
	fb.watching = v
	fb.watchMtx.Unlock()
}

func readPayload(scanner *bufio.Reader, payload []string) error {
	lineCount := 0
	for {
		line, err := scanner.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.Trim(line, " \r\n")
		if len(line) == 0 {
			// empty line
			if lineCount == len(payload) {
				return nil // everything OK
			} else {
				return errors.New("Bad formated body")
			}
		}
		if lineCount < len(payload) {
			payload[lineCount] = line
			lineCount++
		}
	}
}

// Watch listens for changes on a firebase instance and
// passes over to the given chan.
//
// Only one connection can be established at a time. The
// second call to this function without a call to fb.StopWatching
// will close the channel given and return nil immediately.
func (fb *Firebase) Watch(notifications chan Event) error {

	if fb.isWatching() {
		close(notifications)
		return nil
	}
	// set watching flag
	fb.setWatching(true)

	// build SSE request
	req, err := http.NewRequest("GET", fb.String(), nil)
	if err != nil {
		fb.setWatching(false)
		return err
	}
	req.Header.Add("Accept", "text/event-stream")

	// do request
	resp, err := fb.client.Do(req)
	if err != nil {
		fb.setWatching(false)
		return err
	}

	// start parsing response body
	go func() {
		// build scanner for response body
		scanner := bufio.NewReader(resp.Body)
		var (
			scanErr        error
			closedManually bool
			mtx            sync.Mutex
		)

		// monitor the stopWatching channel
		// if we're told to stop, close the response Body
		go func() {
			<-fb.stopWatching

			mtx.Lock()
			closedManually = true
			mtx.Unlock()

			resp.Body.Close()
		}()
	scanning:
		for scanErr == nil {
			// split event string
			// 		event: put
			// 		data: {"path":"/","data":{"foo":"bar"}}

			payload := make([]string, 2)
			scanErr = readPayload(scanner, payload)
			if scanErr != nil {
				break scanning
			}

			var eventType string
			if !strings.HasPrefix(payload[0], "event:") {
				scanErr = errors.New("First line does not start with event:")
				break scanning
			}
			eventType = strings.TrimPrefix(payload[0], "event:")
			eventType = strings.Trim(eventType, " \r\n")

			var eventData string
			if !strings.HasPrefix(payload[1], "data:") {
				scanErr = errors.New("Second line does not start with data:")
				break scanning
			}
			eventData = strings.TrimPrefix(payload[1], "data:")
			eventData = strings.Trim(eventData, " \r\n")

			// create a base event
			event := Event{
				Type:    eventType,
				payload: eventData,
			}

			// should be reacting differently based off the type of event
			switch event.Type {
			case "put", "patch": // we've got extra data we've got to parse

				// the extra data is in json format
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(eventData), &data); err != nil {
					scanErr = err
					break scanning
				}

				// set the extra fields
				event.Path = data["path"].(string)
				event.Data = data["data"]

				// ship it
				notifications <- event
			case "keep-alive":
				// received ping - nothing to do here
			case "cancel":
				// The data for this event is null
				// This event will be sent if the Security and Firebase Rules
				// cause a read at the requested location to no longer be allowed

				// send the cancel event
				notifications <- event
				break scanning
			case "auth_revoked":
				// The data for this event is a string indicating that a the credential has expired
				// This event will be sent when the supplied auth parameter is no longer valid

				// TODO: handle
			case "rules_debug":
				log.Printf("Rules-Debug: %s\n", eventData)
			}
		}

		// check error type
		mtx.Lock()
		closed := closedManually
		mtx.Unlock()
		if !closed && scanErr != nil {
			notifications <- Event{
				Type: EventTypeError,
				Data: scanErr,
			}
		}

		// call stop watching to reset state and cleanup routines
		fb.StopWatching()
		close(notifications)

	}()
	return nil
}

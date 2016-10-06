package firego

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

const (
	// EventTypePut is the event type sent when new data is inserted to the
	// Firebase instance.
	EventTypePut = "put"
	// EventTypePatch is the event type sent when data at the Firebase instance is
	// updated.
	EventTypePatch = "patch"
	// EventTypeError is the event type sent when an unknown error is encountered.
	EventTypeError = "event_error"
	// EventTypeAuthRevoked is the event type sent when the supplied auth parameter
	// is no longer valid.
	EventTypeAuthRevoked = "auth_revoked"

	eventTypeKeepAlive  = "keep-alive"
	eventTypeCancel     = "cancel"
	eventTypeRulesDebug = "rules_debug"
)

// Event represents a notification received when watching a
// firebase reference.
type Event struct {
	// Type of event that was received
	Type string
	// Path to the data that changed
	Path string
	// Data that changed
	Data interface{}

	rawData string
}

// Value converts the raw payload of the event into the given interface.
func (e Event) Value(v interface{}) error {
	var tmp struct {
		Data interface{} `json:"data"`
	}
	tmp.Data = &v
	return json.Unmarshal([]byte(e.rawData), &tmp)
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

	stop := make(chan struct{})
	events, err := fb.watch(stop)
	if err != nil {
		return err
	}

	var closedManually bool

	// monitor the stopWatching channel
	// if we're told to stop, close the response Body
	go func() {
		<-fb.stopWatching

		closedManually = true
		close(stop)
	}()

	go func() {
		for event := range events {
			if event.Type == EventTypeError && closedManually {
				break
			}

			notifications <- event
		}

		close(notifications)
	}()

	return nil
}

func (fb *Firebase) watch(stop chan struct{}) (chan Event, error) {
	// build SSE request
	req, err := http.NewRequest("GET", fb.String(), nil)
	if err != nil {
		fb.setWatching(false)
		return nil, err
	}
	req.Header.Add("Accept", "text/event-stream")

	// do request
	resp, err := fb.client.Do(req)
	if err != nil {
		fb.setWatching(false)
		return nil, err
	}

	notifications := make(chan Event)

	go func() {
		<-stop
		defer resp.Body.Close()
	}()

	// start parsing response body
	go func() {

		// build scanner for response body
		scanner := bufio.NewReader(resp.Body)
		var scanErr error

	scanning:
		for scanErr == nil {
			// split event string
			// 		event: put
			// 		data: {"path":"/","data":{"foo":"bar"}}

			var evt string
			var dat string

			// scan for 'event:'
			evt, scanErr = scanner.ReadString('\n')
			if scanErr != nil {
				break scanning
			}
			evt = strings.TrimSuffix(evt, "\n")

			// scan for 'data:'
			dat, scanErr = scanner.ReadString('\n')
			if scanErr != nil {
				break scanning
			}
			dat = strings.TrimSuffix(dat, "\n")

			// strip off last '\n'
			_, scanErr = scanner.ReadString('\n')
			if scanErr != nil {
				break scanning
			}

			// create a base event
			event := Event{
				Type:    strings.Replace(evt, "event: ", "", 1),
				rawData: strings.Replace(dat, "data: ", "", 1),
			}

			// should be reacting differently based off the type of event
			switch event.Type {
			case EventTypePut, EventTypePatch:
				// we've got extra data we've got to parse
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(strings.Replace(dat, "data: ", "", 1)), &data); err != nil {
					scanErr = err
					break scanning
				}

				// set the extra fields
				event.Path = data["path"].(string)
				event.Data = data["data"]

				// ship it
				notifications <- event
			case eventTypeKeepAlive:
				// received ping - nothing to do here
			case eventTypeCancel:
				// The data for this event is null
				// This event will be sent if the Security and Firebase Rules
				// cause a read at the requested location to no longer be allowed

				// send the cancel event
				notifications <- event
				break scanning
			case EventTypeAuthRevoked:
				// The data for this event is a string indicating that a the credential has expired
				// This event will be sent when the supplied auth parameter is no longer valid
				event.Data = strings.Replace(dat, "data: ", "", 1)
				notifications <- event
				break scanning
			case eventTypeRulesDebug:
				log.Printf("Rules-Debug: %s\n%s\n", evt, dat)
			}
		}

		if scanErr != nil {
			notifications <- Event{
				Type: EventTypeError,
				Data: scanErr,
			}
		}

		// cleanup routines
		close(notifications)
	}()
	return notifications, nil
}

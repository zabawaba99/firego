package database

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
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

	rawData []byte
}

// Value converts the raw payload of the event into the given interface.
func (e Event) Value(v interface{}) error {
	var tmp struct {
		Data interface{} `json:"data"`
	}
	tmp.Data = &v
	return json.Unmarshal(e.rawData, &tmp)
}

// StopWatching stops tears down all connections that are watching.
func (db *Database) StopWatching() {
	db.watchMtx.Lock()
	defer db.watchMtx.Unlock()

	if db.watching {
		// flip the bit back to not watching
		db.watching = false
		// signal connection to terminal
		db.stopWatching <- struct{}{}
	}
}

func (db *Database) setWatching(v bool) {
	db.watchMtx.Lock()
	db.watching = v
	db.watchMtx.Unlock()
}

// Watch listens for changes on a firebase instance and
// passes over to the given chan.
//
// Only one connection can be established at a time. The
// second call to this function without a call to db.StopWatching
// will close the channel given and return nil immediately.
func (db *Database) Watch(notifications chan Event) error {
	db.watchMtx.Lock()
	if db.watching {
		db.watchMtx.Unlock()
		close(notifications)
		return nil
	}
	db.watching = true
	db.watchMtx.Unlock()

	stop := make(chan struct{})
	events, err := db.watch(stop)
	if err != nil {
		return err
	}

	var closedManually bool

	go func() {
		<-db.stopWatching
		closedManually = true
		stop <- struct{}{}
	}()

	go func() {
		defer func() {
			close(notifications)
			close(stop)
		}()

		for event := range events {
			if closedManually {
				return
			}

			notifications <- event
		}
	}()

	return nil
}

func readLine(rdr *bufio.Reader, prefix string) ([]byte, error) {
	// read event: line
	line, err := rdr.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	// empty line check for empty prefix
	if len(prefix) == 0 {
		line = bytes.TrimSpace(line)
		if len(line) != 0 {
			return nil, errors.New("expected empty line")
		}
		return line, nil
	}

	// check line has event prefix
	if !bytes.HasPrefix(line, []byte(prefix)) {
		return nil, errors.New("missing prefix")
	}

	// trim space
	line = line[len(prefix):]
	return bytes.TrimSpace(line), nil
}

func (db *Database) watch(stop chan struct{}) (chan Event, error) {
	// build SSE request
	req, err := http.NewRequest("GET", db.String(), nil)
	if err != nil {
		db.setWatching(false)
		return nil, err
	}
	req.Header.Add("Accept", "text/event-stream")

	// do request
	resp, err := db.client.Do(req)
	if err != nil {
		db.setWatching(false)
		return nil, err
	}

	notifications := make(chan Event)

	go func() {
		<-stop
		resp.Body.Close()
	}()

	heartbeat := make(chan struct{})
	go func() {
		for {
			select {
			case <-heartbeat:
				// do nothing
			case <-time.After(db.watchHeartbeat):
				resp.Body.Close()
				return
			}
		}
	}()

	// start parsing response body
	go func() {
		defer func() {
			resp.Body.Close()
			close(notifications)
		}()

		// build scanner for response body
		scanner := bufio.NewReader(resp.Body)
		sendError := func(err error) {
			notifications <- Event{
				Type: EventTypeError,
				Data: err,
			}
		}
		for {
			select {
			case heartbeat <- struct{}{}:
			default:
			}
			// scan for 'event:'
			evt, err := readLine(scanner, "event: ")
			if err != nil {
				sendError(err)
				return
			}

			// scan for 'data:'
			dat, err := readLine(scanner, "data: ")
			if err != nil {
				sendError(err)
				return
			}

			// read the empty line
			_, err = readLine(scanner, "")
			if err != nil {
				sendError(err)
				return
			}

			// create a base event
			event := Event{
				Type:    string(evt),
				Data:    string(dat),
				rawData: dat,
			}

			// should be reacting differently based off the type of event
			switch event.Type {
			case EventTypePut, EventTypePatch:
				// we've got extra data we've got to parse
				var data map[string]interface{}
				if err := json.Unmarshal(event.rawData, &data); err != nil {
					sendError(err)
					return
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
				return
			case EventTypeAuthRevoked:
				// The data for this event is a string indicating that a the credential has expired
				// This event will be sent when the supplied auth parameter is no longer valid
				notifications <- event
				return
			case eventTypeRulesDebug:
				log.Printf("Rules-Debug: %s\n%s\n", evt, dat)
			}
		}
	}()
	return notifications, nil
}

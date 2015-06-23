package firego

import (
	"bufio"
	"encoding/json"
	"log"
	"strings"
)

// Event represents a notification received when watching a
// firebase reference
type Event struct {
	// Type of event that was received
	Type string
	// Path to the data that changed
	Path string
	// Data that changed
	Data interface{}
}

// StopWatching stops tears down all connections that are watching
func (fb *Firebase) StopWatching() {
	if fb.watching {
		// signal connection to terminal
		fb.stopWatching <- struct{}{}
		// flip the bit back to not watching
		fb.watching = false
	}
}

// Watch listens for changes on a firebase instance and
// passes over to the given chan.
//
// Only one connection can be established at a time. The
// second call to this function without a call to fb.StopWatching
// will close the channel given and return nil immediately
func (fb *Firebase) Watch(notifications chan Event) error {
	if fb.watching {
		close(notifications)
		return nil
	}

	// build SSE request
	req, err := fb.makeRequest("GET", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "text/event-stream")

	// do request
	resp, err := fb.client.Do(req)
	if err != nil {
		return err
	}

	// set watching flag
	fb.watching = true

	// start parsing response body
	go func() {
		// build scanner for response body
		scanner := bufio.NewReader(resp.Body)
		var scanResult error

		// monitor the stopWatching channel
		// if we're told to stop, close the response Body
		go func() {
			<-fb.stopWatching
			resp.Body.Close()
		}()
	scanning:
		for scanResult == nil {
			// split event string
			// 		event: put
			// 		data: {"path":"/","data":{"foo":"bar"}}

			var evt []byte
			var dat []byte
			isPrefix := true
			var result []byte

			// For possible results larger than 64 * 1024 bytes (MaxTokenSize)
			// we need bufio#ReadLine()
			// 1. step: scan for the 'event:' part. ReadLine() oes not return the \n
			// so we have to add it to our result buffer.
			evt, isPrefix, scanResult = scanner.ReadLine()
			if scanResult != nil {
				break scanning
			}
			result = append(result, evt...)
			result = append(result, '\n')

			// 2. step: scan for the 'data:' part. Firebase returns just one 'data:'
			// part, but the value can be very large. If we exceed a certain length
			// isPrefix will be true until all data is read.
			for {
				dat, isPrefix, scanResult = scanner.ReadLine()
				if scanResult != nil {
					break scanning
				}
				result = append(result, dat...)
				if !isPrefix {
					break
				}
			}
			// Again we add the \n
			result = append(result, '\n')
			_, _, scanResult = scanner.ReadLine()
			if scanResult != nil {
				break scanning
			}

			txt := string(result)
			parts := strings.Split(txt, "\n")

			// create a base event
			event := Event{
				Type: strings.Replace(parts[0], "event: ", "", 1),
			}

			// should be reacting differently based off the type of event
			switch event.Type {
			case "put", "patch": // we've got extra data we've got to parse

				// the extra data is in json format
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(strings.Replace(parts[1], "data: ", "", 1)), &data); err != nil {
					log.Fatal(err)
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
				log.Printf("Rules-Debug: %s\n", txt)
			}
		}

		// call stop watching to reset state and cleanup routines
		fb.StopWatching()
		close(notifications)

	}()
	return nil
}

package firego

import (
	"encoding/json"
	"errors"
	"fmt"
)

// TransactionFn is used to run a transaction on a Firebase reference.
// See Firebase.Transaction for more information.
type TransactionFn func(currentSnapshot interface{}) (result interface{}, err error)

func (fb *Firebase) runTransaction(fn TransactionFn) error {
	// fetch etag and current value
	headers, body, err := fb.doRequest("GET", nil, withHeader("X-Firebase-ETag", "true"))
	if err != nil {
		return err
	}

	etag := headers.Get("ETag")
	if len(etag) == 0 {
		return errors.New("no etag returned by Firebase")
	}

	var currentSnapshot interface{}
	if err := json.Unmarshal(body, &currentSnapshot); err != nil {
		return fmt.Errorf("failed to unmarshal Firebase response. %s", err)
	}

	// run transaction
	result, err := fn(currentSnapshot)
	if err != nil {
		return err
	}

	newBody, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction result. %s", err)
	}

	// attempt to update it
	_, _, err = fb.doRequest("POST", newBody, withHeader("if-match", etag))
	return err
}

// Transaction runs a transaction on the data at this location. The TransactionFn parameter
// will be called, possibly multiple times, with the current data at this location.
// It is responsible for inspecting that data and specifying either the desired new data
// at the location or that the transaction should be aborted.
//
// Since the provided function may be called repeatedly for the same transaction, be extremely careful of
// any side effects that may be triggered by this method.
//
// Best practices for this method are to rely only on the data that is passed in.
func (fb *Firebase) Transaction(fn TransactionFn) error {
	err := fb.runTransaction(fn)
	// TODO: maybe this should be configurable - should we have a limit?
	for i := 0; i < 10 && err != nil; i++ {
		err = fb.runTransaction(fn)
	}

	if err != nil {
		return fmt.Errorf("failed to run transaction. %s", err)
	}
	return nil
}

package database

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/zabawaba99/firego/sync"
)

// ChildEventFunc is the type of function that is called for every
// new child added under a firebase reference. The snapshot argument
// contains the data that was added. The previousChildKey argument
// contains the key of the previous child that this function was called for.
type ChildEventFunc func(snapshot DataSnapshot, previousChildKey string)

// ChildAdded listens on the firebase instance and executes the callback
// for every child that is added.
//
// You cannot set the same function twice on a Firebase reference, if you do
// the first function will be overridden and you will not be able to close the
// connection.
func (db *Database) ChildAdded(fn ChildEventFunc) error {
	return db.addEventFunc(fn, fn.childAdded)
}

func (fn ChildEventFunc) childAdded(localDB *sync.Database, prevKey *string, notifications chan Event) error {
	for event := range notifications {
		if event.Type == EventTypeError {
			err, ok := event.Data.(error)
			if !ok {
				err = fmt.Errorf("Got error from event %#v", event)
			}
			return err
		}

		if event.Type != EventTypePut {
			continue
		}

		child := strings.Split(event.Path[1:], "/")[0]
		if event.Data == nil {
			localDB.Del(child)
			continue
		}

		if _, ok := localDB.Get("").Child(child); ok {
			// if the child isn't being added, forget it
			continue
		}

		m, ok := event.Data.(map[string]interface{})
		if child == "" && ok {
			// if we were given a map at the root then we have
			// to send an event per child
			for _, k := range sortedKeys(m) {
				v := m[k]
				node := sync.NewNode(k, v)
				localDB.Add(k, node)
				fn(newSnapshot(node), *prevKey)
				*prevKey = k
			}
			continue
		}

		// we have a single event to process
		node := sync.NewNode(child, event.Data)
		localDB.Add(strings.Trim(child, "/"), node)

		fn(newSnapshot(node), *prevKey)
		*prevKey = child
	}
	return nil
}

// ChildChanged listens on the firebase instance and executes the callback
// for every child that is changed.
//
// You cannot set the same function twice on a Firebase reference, if you do
// the first function will be overridden and you will not be able to close the
// connection.
func (db *Database) ChildChanged(fn ChildEventFunc) error {
	return db.addEventFunc(fn, fn.childChanged)
}

func (fn ChildEventFunc) childChanged(localDB *sync.Database, prevKey *string, notifications chan Event) error {
	first, ok := <-notifications
	if !ok {
		return errors.New("channel closed")
	}

	localDB.Add("", sync.NewNode("", first.Data))
	for event := range notifications {
		if event.Type == EventTypeError {
			err, ok := event.Data.(error)
			if !ok {
				err = fmt.Errorf("Got error from event %#v", event)
			}
			return err
		}

		path := strings.Trim(event.Path, "/")
		if event.Data == nil {
			localDB.Del(path)
			continue
		}

		child := strings.Split(path, "/")[0]
		node := sync.NewNode(child, event.Data)

		dbNode := localDB.Get("")
		if _, ok := dbNode.Child(child); child != "" && !ok {
			// if the child is new, ignore it.
			localDB.Add(path, node)
			continue
		}

		if m, ok := event.Data.(map[string]interface{}); child == "" && ok {
			// we've got children so send an event per child
			for _, k := range sortedKeys(m) {
				v := m[k]
				node := sync.NewNode(k, v)
				newPath := strings.TrimPrefix(child+"/"+k, "/")
				if _, ok := dbNode.Child(k); !ok {
					localDB.Add(newPath, node)
					continue
				}

				localDB.Update(newPath, node)
				fn(newSnapshot(node), *prevKey)
				*prevKey = k
			}
			continue
		}

		localDB.Update(path, node)
		fn(newSnapshot(localDB.Get(child)), *prevKey)
		*prevKey = child
	}
	return nil
}

// ChildRemoved listens on the firebase instance and executes the callback
// for every child that is deleted.
//
// You cannot set the same function twice on a Firebase reference, if you do
// the first function will be overridden and you will not be able to close the
// connection.
func (db *Database) ChildRemoved(fn ChildEventFunc) error {
	return db.addEventFunc(fn, fn.childRemoved)
}

func (fn ChildEventFunc) childRemoved(localDB *sync.Database, prevKey *string, notifications chan Event) error {
	first, ok := <-notifications
	if !ok {
		return errors.New("channel closed")
	}

	node := sync.NewNode("", first.Data)
	localDB.Add("", node)

	for event := range notifications {
		if event.Type == EventTypeError {
			err, ok := event.Data.(error)
			if !ok {
				err = fmt.Errorf("Got error from event %#v", event)
			}
			return err
		}

		path := strings.Trim(event.Path, "/")
		node := sync.NewNode(path, event.Data)

		if event.Type == EventTypePatch {
			localDB.Update(path, node)
			continue
		}

		if event.Data != nil {
			localDB.Add(path, node)
			continue
		}

		if path == "" {
			// if node that is being listened to is deleted,
			// an event should be triggered for every child
			children := localDB.Get("").Children
			orderedChildren := make([]string, len(children))
			var i int
			for k := range children {
				orderedChildren[i] = k
				i++
			}

			sort.Strings(orderedChildren)

			for _, k := range orderedChildren {
				node := localDB.Get(k)
				fn(newSnapshot(node), "")
				localDB.Del(k)
			}

			localDB.Del(path)
			continue
		}

		node = localDB.Get(path)
		fn(newSnapshot(node), "")
		localDB.Del(path)
	}
	return nil
}

type handleSSEFunc func(*sync.Database, *string, chan Event) error

func (db *Database) addEventFunc(fn ChildEventFunc, handleSSE handleSSEFunc) error {
	db.eventMtx.Lock()
	defer db.eventMtx.Unlock()

	stop := make(chan struct{})
	key := fmt.Sprintf("%v", fn)
	if _, ok := db.eventFuncs[key]; ok {
		return nil
	}

	db.eventFuncs[key] = stop
	notifications, err := db.watch(stop)
	if err != nil {
		return err
	}

	localDB := sync.NewDB()
	prevKey := new(string)
	var run func(notifications chan Event, backoff time.Duration)
	run = func(notifications chan Event, backoff time.Duration) {
		db.eventMtx.Lock()
		if _, ok := db.eventFuncs[key]; !ok {
			db.eventMtx.Unlock()
			// the func has been removed
			return
		}
		db.eventMtx.Unlock()

		if err := handleSSE(localDB, prevKey, notifications); err == nil {
			// we returned gracefully
			return
		}

		// give firebase some time
		backoff *= 2
		time.Sleep(backoff)

		// try and reconnect
		for notifications, err = db.watch(stop); err != nil; time.Sleep(backoff) {
			db.eventMtx.Lock()
			if _, ok := db.eventFuncs[key]; !ok {
				db.eventMtx.Unlock()
				// func has been removed
				return
			}
			db.eventMtx.Unlock()
		}

		// give this another shot
		run(notifications, backoff)
	}

	go run(notifications, db.watchHeartbeat)
	return nil
}

// RemoveEventFunc removes the given function from the firebase
// reference.
func (db *Database) RemoveEventFunc(fn ChildEventFunc) {
	db.eventMtx.Lock()
	defer db.eventMtx.Unlock()

	key := fmt.Sprintf("%v", fn)
	stop, ok := db.eventFuncs[key]
	if !ok {
		return
	}

	delete(db.eventFuncs, key)
	close(stop)
}

func sortedKeys(m map[string]interface{}) []string {
	orderedKeys := make([]string, len(m))
	var i int
	for k := range m {
		orderedKeys[i] = k
		i++
	}

	sort.Strings(orderedKeys)
	return orderedKeys
}

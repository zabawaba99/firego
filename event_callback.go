package firego

import (
	"fmt"
	"sort"
	"strings"
)

// ChildEventFunc is the type of function that is called for every
// new child added under a firebase reference. The snapshot argument
// contains the data that was added. The previousChildKey argument
// contains the key of the previous child that this function was called for.
type ChildEventFunc func(snapshot *DataSnapshot, previousChildKey string)

// ChildAdded listens on the firebase instance and executes the callback
// for every child that is added.
func (fb *Firebase) ChildAdded(fn ChildEventFunc) error {
	handleSSE := func(notifications chan Event) {
		first, ok := <-notifications
		if !ok {
			return
		}

		var pk string
		db := newDatabase()
		children, ok := first.Data.(map[string]interface{})
		if ok {
			// we've got children so send an event per child
			orderedChildren := make([]string, len(children))
			var i int
			for k := range children {
				orderedChildren[i] = k
				i++
			}

			sort.Strings(orderedChildren)

			for _, k := range orderedChildren {
				v := children[k]
				snapshot := newSnapshot(v)
				db.add(k, snapshot)
				fn(snapshot, pk)
				pk = k
			}
		}

		for event := range notifications {
			if event.Type != "put" {
				continue
			}

			child := strings.Split(event.Path[1:], "/")[0]
			if event.Data == nil {
				// delete
				db.del(child)
				continue
			}

			if _, ok := db.rootNode.Child(child); ok {
				// if the child isn't being added, forget it
				continue
			}

			snapshot := newSnapshot(event.Data)
			db.add(sanitizePath(child), snapshot)

			fn(snapshot, pk)
			pk = child
		}
	}

	return fb.addEventFunc(fn, handleSSE)
}

func (fb *Firebase) ChildRemoved(fn ChildEventFunc) error {
	handleSSE := func(notifications chan Event) {
		first, ok := <-notifications
		if !ok {
			return
		}

		db := newDatabase()
		db.add("", newSnapshot(first.Data))

		for event := range notifications {
			path := sanitizePath(event.Path)
			snapshot := newSnapshot(event.Data)

			if event.Type == "patch" {
				db.update(path, snapshot)
				continue
			}

			if event.Data != nil {
				db.add(path, snapshot)
				continue
			}

			// if node is not root, notify for child and move along
			if path != "" {
				snapshot = db.get(path)
				fn(snapshot, "")
				db.del(path)

				continue
			}

			// if node that is being listened to is deleted,
			// an event should be triggered for every child
			orderedChildren := make([]string, len(db.rootNode.children))
			var i int
			for k := range db.rootNode.children {
				orderedChildren[i] = k
				i++
			}

			sort.Strings(orderedChildren)

			for _, k := range orderedChildren {
				fn(db.get(k), "")
				db.del(k)
			}

			db.del(path)

		}
	}
	return fb.addEventFunc(fn, handleSSE)
}

func (fb *Firebase) addEventFunc(fn ChildEventFunc, handleSSE func(chan Event)) error {
	fb.eventMtx.Lock()
	defer fb.eventMtx.Unlock()

	stop := make(chan struct{})
	key := fmt.Sprintf("%v", fn)
	if _, ok := fb.eventFuncs[key]; ok {
		return nil
	}

	fb.eventFuncs[key] = stop

	notifications, err := fb.watch(stop)
	if err != nil {
		return err
	}

	go handleSSE(notifications)
	return nil
}

// RemoveEventFunc removes the given function from the firebase
// reference.
func (fb *Firebase) RemoveEventFunc(fn ChildEventFunc) {
	fb.eventMtx.Lock()
	defer fb.eventMtx.Unlock()

	key := fmt.Sprintf("%v", fn)
	stop, ok := fb.eventFuncs[key]
	if !ok {
		return
	}

	delete(fb.eventFuncs, key)
	close(stop)
}

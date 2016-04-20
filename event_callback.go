package firego

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zabawaba99/firego/sync"
)

// ChildEventFunc is the type of function that is called for every
// new child added under a firebase reference. The snapshot argument
// contains the data that was added. The previousChildKey argument
// contains the key of the previous child that this function was called for.
type ChildEventFunc func(snapshot DataSnapshot, previousChildKey string)

// ChildAdded listens on the firebase instance and executes the callback
// for every child that is added.
func (fb *Firebase) ChildAdded(fn ChildEventFunc) error {
	return fb.addEventFunc(fn, childAdded)
}

func childAdded(fn ChildEventFunc, notifications chan Event) {
	first, ok := <-notifications
	if !ok {
		return
	}

	var pk string
	db := sync.NewDB()
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
			node := sync.NewNode(k, v)
			db.Add(k, node)
			fn(newSnapshot(node), pk)
			pk = k
		}
	}

	for event := range notifications {
		if event.Type != EventTypePut {
			continue
		}

		child := strings.Split(event.Path[1:], "/")[0]
		if event.Data == nil {
			// delete
			db.Del(child)
			continue
		}

		if _, ok := db.Get("").Child(child); ok {
			// if the child isn't being added, forget it
			continue
		}

		node := sync.NewNode(child, event.Data)
		db.Add(strings.Trim(child, "/"), node)

		fn(newSnapshot(node), pk)
		pk = child
	}
}

// ChildRemoved listens on the firebase instance and executes the callback
// for every child that is deleted.
func (fb *Firebase) ChildRemoved(fn ChildEventFunc) error {
	return fb.addEventFunc(fn, childRemoved)
}

func childRemoved(fn ChildEventFunc, notifications chan Event) {
	first, ok := <-notifications
	if !ok {
		return
	}

	db := sync.NewDB()
	node := sync.NewNode("", first.Data)
	db.Add("", node)

	for event := range notifications {
		path := strings.Trim(event.Path, "/")
		node := sync.NewNode(path, event.Data)

		if event.Type == EventTypePatch {
			db.Update(path, node)
			continue
		}

		if event.Data != nil {
			db.Add(path, node)
			continue
		}

		if path == "" {
			// if node that is being listened to is deleted,
			// an event should be triggered for every child
			children := db.Get("").Children
			orderedChildren := make([]string, len(children))
			var i int
			for k := range children {
				orderedChildren[i] = k
				i++
			}

			sort.Strings(orderedChildren)

			for _, k := range orderedChildren {
				node := db.Get(k)
				fn(newSnapshot(node), "")
				db.Del(k)
			}

			db.Del(path)
			continue
		}

		node = db.Get(path)
		fn(newSnapshot(node), "")
		db.Del(path)
	}
}

func (fb *Firebase) addEventFunc(fn ChildEventFunc, handleSSE func(ChildEventFunc, chan Event)) error {
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

	go handleSSE(fn, notifications)
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

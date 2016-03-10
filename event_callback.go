package firego

import (
	"fmt"
	"sort"
	"strings"
)

type DataSnapshot interface{}
type ChildEventFunc func(snapshot DataSnapshot, previousChildKey string)

func (fb *Firebase) ChildAdded(fn ChildEventFunc) error {
	stop := make(chan struct{})
	key := fmt.Sprintf("%v", fn)
	fb.eventFuncs[key] = stop

	notifications, err := fb.watch(stop)
	if err != nil {
		return err
	}

	go func() {
		first, ok := <-notifications
		if !ok {
			return
		}

		pk := ""
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
				fn(DataSnapshot(v), pk)
				pk = k
			}
		} else {
			children = map[string]interface{}{}
		}

		for event := range notifications {
			if event.Type != "put" {
				continue
			}

			child := strings.Split(event.Path[1:], "/")[0]
			if event.Data == nil {
				// delete
				delete(children, child)
				continue
			}

			_, ok := children[child]
			if ok {
				// if the child isn't being added, forget it
				continue
			}

			fn(DataSnapshot(event.Data), pk)
			pk = child
			children[child] = true
		}
	}()

	return nil
}

func (fb *Firebase) RemoveEventFunc(fn ChildEventFunc) {
	key := fmt.Sprintf("%v", fn)
	stop, ok := fb.eventFuncs[key]
	if !ok {
		return
	}

	delete(fb.eventFuncs, key)
	close(stop)
}

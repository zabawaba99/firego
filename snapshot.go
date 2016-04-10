package firego

import (
	"fmt"
	"strconv"
)

// DataSnapshot instances contains data from a Firebase reference.
type DataSnapshot struct {
	key       string
	value     interface{}
	children  map[string]*DataSnapshot
	parent    *DataSnapshot
	sliceKids bool
}

func newSnapshot(key string, data interface{}) *DataSnapshot {
	d := &DataSnapshot{
		key:      key,
		children: map[string]*DataSnapshot{},
	}

	switch data := data.(type) {
	case map[string]interface{}:
		for k, v := range data {
			child := newSnapshot(k, v)
			child.parent = d
			d.children[k] = child
		}
	case []interface{}:
		d.sliceKids = true
		for i, v := range data {
			child := newSnapshot(strconv.FormatInt(int64(i), 10), v)
			child.parent = d
			d.children[child.key] = child
		}
	case string, int, int8, int16, int32, int64, float32, float64, bool:
		d.value = data
	case nil:
		// do nothing
	default:
		fmt.Printf("Type(%T) not supported\nIf you see this log please report an issue on https://github.com/zabawaba99/firego", data)
	}

	return d
}

// Key retrieves the key for the source location of this snapshot
func (d *DataSnapshot) Key() string {
	return d.key
}

// Value retrieves the data contained in this snapshot.
func (d *DataSnapshot) Value() interface{} {
	return d.value
}

func (d *DataSnapshot) Child(name string) (*DataSnapshot, bool) {
	s, ok := d.children[name]
	return s, ok
}

func (d *DataSnapshot) isNil() bool {
	return d.value == nil && len(d.children) == 0
}

func (d *DataSnapshot) merge(newSnapshot *DataSnapshot) {
	for k, v := range newSnapshot.children {
		d.children[k] = v
	}
	d.value = newSnapshot.value
}

func (d *DataSnapshot) prune() *DataSnapshot {
	if len(d.children) > 0 || d.value != nil {
		return nil
	}

	parent := d.parent
	d.parent = nil
	return parent
}

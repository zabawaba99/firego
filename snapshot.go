package firego

import "fmt"

// DataSnapshot instances contains data from a Firebase reference.
type DataSnapshot struct {
	value     interface{}
	children  map[string]*DataSnapshot
	parent    *DataSnapshot
	sliceKids bool
}

func newSnapshot(data interface{}) *DataSnapshot {
	d := &DataSnapshot{children: map[string]*DataSnapshot{}}

	switch data := data.(type) {
	case map[string]interface{}:
		for k, v := range data {
			child := newSnapshot(v)
			child.parent = d
			d.children[k] = child
		}
	case []interface{}:
		d.sliceKids = true
		for i, v := range data {
			child := newSnapshot(v)
			child.parent = d
			d.children[fmt.Sprint(i)] = child
		}
	case string, int, int8, int16, int32, int64, float32, float64, bool:
		d.value = data
	case nil:
		// do nothing
	default:
		panic(fmt.Sprintf("Type(%T) not supported\n", data))
	}

	return d
}

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

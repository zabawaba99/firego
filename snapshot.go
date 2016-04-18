package firego

import "github.com/zabawaba99/firego/sync"

// DataSnapshot instances contains data from a Firebase reference.
type DataSnapshot struct {
	node *sync.Node
}

func newSnapshot(node *sync.Node) DataSnapshot {
	return DataSnapshot{node: node}
}

// Key retrieves the key for the source location of this snapshot
func (d *DataSnapshot) Key() string {
	return d.node.Key
}

// Value retrieves the data contained in this snapshot.
func (d *DataSnapshot) Value() interface{} {
	return d.node.Value
}

// Child gets a DataSnapshot for the location at the specified relative path.
// The relative path can either be a simple child key (e.g. 'fred') or a deeper
// slash-separated path (e.g. 'fred/name/first').
func (d *DataSnapshot) Child(name string) (DataSnapshot, bool) {
	node, ok := d.node.Child(name)
	return newSnapshot(node), ok
}

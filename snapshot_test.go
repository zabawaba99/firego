package firego

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDataSnapshotWithKids(children map[string]*DataSnapshot) *DataSnapshot {
	d := &DataSnapshot{}
	for _, child := range children {
		child.parent = d
	}
	d.children = children
	return d
}

func equalNodes(expected, actual *DataSnapshot) error {
	if ec, ac := len(expected.children), len(actual.children); ec != ac {
		return fmt.Errorf("Children count is not the same\n\tExpected: %d\n\tActual: %d", ec, ac)
	}

	if len(expected.children) == 0 {
		if !assert.ObjectsAreEqualValues(expected.value, actual.value) {
			return fmt.Errorf("Node values not equal\n\tExpected: %T %v\n\tActual: %T %v", expected.value, expected.value, actual.value, actual.value)
		}
		return nil
	}

	for child, n := range expected.children {
		n2, ok := actual.children[child]
		if !ok {
			return fmt.Errorf("Expected node to have child: %s", child)
		}

		err := equalNodes(n, n2)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestNewSnapshot(t *testing.T) {

	for _, test := range []struct {
		name     string
		snapshot *DataSnapshot
	}{
		{
			name:     "scalars/string",
			snapshot: newSnapshot("", "foo"),
		},
		{
			name:     "scalars/number",
			snapshot: newSnapshot("", 2),
		},
		{
			name:     "scalars/decimal",
			snapshot: newSnapshot("", 2.2),
		},
		{
			name:     "scalars/boolean",
			snapshot: newSnapshot("", false),
		},
		{
			name:     "arrays/strings",
			snapshot: newSnapshot("", []interface{}{"foo", "bar"}),
		},
		{
			name:     "arrays/booleans",
			snapshot: newSnapshot("", []interface{}{true, false}),
		},
		{
			name:     "arrays/numbers",
			snapshot: newSnapshot("", []interface{}{1, 2, 3}),
		},
		{
			name:     "arrays/decimals",
			snapshot: newSnapshot("", []interface{}{1.1, 2.2, 3.3}),
		},
		{
			name: "objects/simple",
			snapshot: newDataSnapshotWithKids(map[string]*DataSnapshot{
				"foo": newSnapshot("", "bar"),
			}),
		},
		{
			name: "objects/complex",
			snapshot: newDataSnapshotWithKids(map[string]*DataSnapshot{
				"foo":  newSnapshot("", "bar"),
				"foo1": newSnapshot("", 2),
				"foo2": newSnapshot("", true),
				"foo3": newSnapshot("", 3.42),
			}),
		},
		{
			name: "objects/nested",
			snapshot: newDataSnapshotWithKids(map[string]*DataSnapshot{
				"dinosaurs": newDataSnapshotWithKids(map[string]*DataSnapshot{
					"bruhathkayosaurus": newDataSnapshotWithKids(map[string]*DataSnapshot{
						"appeared": newSnapshot("", -70000000),
						"height":   newSnapshot("", 25),
						"length":   newSnapshot("", 44),
						"order":    newSnapshot("", "saurischia"),
						"vanished": newSnapshot("", -70000000),
						"weight":   newSnapshot("", 135000),
					}),
					"lambeosaurus": newDataSnapshotWithKids(map[string]*DataSnapshot{
						"appeared": newSnapshot("", -76000000),
						"height":   newSnapshot("", 2.1),
						"length":   newSnapshot("", 12.5),
						"order":    newSnapshot("", "ornithischia"),
						"vanished": newSnapshot("", -75000000),
						"weight":   newSnapshot("", 5000),
					}),
				}),
				"scores": newDataSnapshotWithKids(map[string]*DataSnapshot{
					"bruhathkayosaurus": newSnapshot("", 55),
					"lambeosaurus":      newSnapshot("", 21),
				}),
			}),
		},
		{
			name: "objects/with_arrays",
			snapshot: newDataSnapshotWithKids(map[string]*DataSnapshot{
				"regular":  newSnapshot("", "item"),
				"booleans": newSnapshot("", []interface{}{false, true}),
				"numbers":  newSnapshot("", []interface{}{1, 2}),
				"decimals": newSnapshot("", []interface{}{1.1, 2.2}),
				"strings":  newSnapshot("", []interface{}{"foo", "bar"}),
			}),
		},
	} {
		data, err := ioutil.ReadFile("fixtures/" + test.name + ".json")
		require.NoError(t, err, test.name)

		var v interface{}
		require.NoError(t, json.Unmarshal(data, &v), test.name)

		n := newSnapshot("", v)
		assert.NoError(t, equalNodes(test.snapshot, n), test.name)
	}
}

func TestPrune(t *testing.T) {
	/*
		Children:	0
		Value:		Non nil
		Parent: 	nil
	*/
	n := newSnapshot("", "foo")
	assert.Nil(t, n.prune())

	/*
		Children:	0
		Value:		Non nil
		Parent: 	Non nil
	*/
	n = newSnapshot("", "foo")
	n.parent = newSnapshot("", 1)
	assert.Nil(t, n.prune())

	/*
		Children:	0
		Value:		nil
		Parent: 	Non nil
	*/
	n = &DataSnapshot{}
	parent := newDataSnapshotWithKids(map[string]*DataSnapshot{"foo": n})
	parentFromPrune := n.prune()

	assert.NotNil(t, parentFromPrune)
	assert.Equal(t, parent, parentFromPrune)
	assert.Nil(t, n.parent)
	assert.Nil(t, n.children)

	/*
		Children:	1
		Value:		nil
		Parent: 	Non nil
	*/
	n = newDataSnapshotWithKids(map[string]*DataSnapshot{"c1": n})
	parent = newDataSnapshotWithKids(map[string]*DataSnapshot{"foo": n})
	assert.Nil(t, n.prune())

	/*
		Children:	1
		Value:		nil
		Parent: 	nil
	*/
	n = newDataSnapshotWithKids(map[string]*DataSnapshot{"c1": n})
	assert.Nil(t, n.prune())

	/*
		Children:	2
		Value:		nil
		Parent: 	Non nil
	*/
	n = newDataSnapshotWithKids(map[string]*DataSnapshot{
		"c1": newSnapshot("", 1),
		"c2": newSnapshot("", 2),
	})
	n.parent = newSnapshot("", "hello!")
	assert.Nil(t, n.prune())
}

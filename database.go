package firego

import "strings"

type database struct {
	rootNode *DataSnapshot
}

func newDatabase() *database {
	return &database{
		rootNode: &DataSnapshot{
			children: map[string]*DataSnapshot{},
		},
	}
}

func (d *database) add(path string, n *DataSnapshot) {
	if path == "" {
		d.rootNode = n
		return
	}

	rabbitHole := strings.Split(path, "/")
	current := d.rootNode
	for i := 0; i < len(rabbitHole)-1; i++ {
		step := rabbitHole[i]
		next, ok := current.children[step]
		if !ok {
			next = &DataSnapshot{
				parent:   current,
				children: map[string]*DataSnapshot{},
			}
			current.children[step] = next
		}
		next.value = nil // no long has a value since it now has a child
		current, next = next, nil
	}

	lastPath := rabbitHole[len(rabbitHole)-1]
	current.children[lastPath] = n
	n.parent = current
}

func (d *database) update(path string, n *DataSnapshot) {
	current := d.rootNode
	rabbitHole := strings.Split(path, "/")

	for i := 0; i < len(rabbitHole); i++ {
		path := rabbitHole[i]
		if path == "" {
			// prevent against empty strings due to strings.Split
			continue
		}
		next, ok := current.children[path]
		if !ok {
			next = &DataSnapshot{parent: current, children: map[string]*DataSnapshot{}}
			current.children[path] = next
		}
		next.value = nil // no long has a value since it now has a child
		current, next = next, nil
	}

	current.merge(n)
}

func (d *database) del(path string) {
	if path == "" {
		d.rootNode = &DataSnapshot{
			children: map[string]*DataSnapshot{},
		}
		return
	}

	rabbitHole := strings.Split(path, "/")
	current := d.rootNode

	// traverse to target node's parent
	var delIdx int
	for ; delIdx < len(rabbitHole)-1; delIdx++ {
		next, ok := current.children[rabbitHole[delIdx]]
		if !ok {
			// item does not exist, no need to do anything
			return
		}

		current = next
	}

	endNode := current
	leafPath := rabbitHole[len(rabbitHole)-1]
	delete(endNode.children, leafPath)

	for tmp := endNode.prune(); tmp != nil; tmp = tmp.prune() {
		delIdx--
		endNode = tmp
	}

	if endNode != nil {
		delete(endNode.children, rabbitHole[delIdx])
	}
}

func (d *database) get(path string) *DataSnapshot {
	current := d.rootNode
	if path == "" {
		return current
	}

	rabbitHole := strings.Split(path, "/")
	for i := 0; i < len(rabbitHole); i++ {
		var ok bool
		current, ok = current.children[rabbitHole[i]]
		if !ok {
			return nil
		}
	}
	return current
}

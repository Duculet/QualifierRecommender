package schematree

import (
	"encoding/gob"
	"fmt"
	"sync"
)

// SchemaNode is a nodes of the Schema FP-Tree
type SchemaNode struct {
	ID            *IItem
	parent        *SchemaNode
	FirstChildren []SchemaNode
	Children      []*SchemaNode
	nextSameID    *SchemaNode // node traversal pointer
	Support       uint32      // total frequency of the node in the path
}

//newRootNode creates a new root node for a given propMap
func newRootNode(pMap propMap) SchemaNode {
	// return schemaNode{newRootiItem(), nil, make(map[*iItem]*schemaNode), nil, 0, nil}
	return SchemaNode{pMap.get("root"), nil, make([]SchemaNode, 5), []*SchemaNode{}, nil, 0}
}

const lockPrime = 97 // arbitrary prime number
var globalItemLocks [lockPrime]*sync.Mutex
var globalNodeLocks [lockPrime]*sync.RWMutex

// decodeGob decodes the schema node from its binary representation
func (node *SchemaNode) decodeGob(d *gob.Decoder, props []*IItem) error {
	// function scoping to allow for garbage collection
	// err := func() error {
	// ID
	var id uint32
	err := d.Decode(&id)
	if err != nil {
		return err
	}
	node.ID = props[int(id)]

	// traversal pointer repopulation
	node.nextSameID = node.ID.traversalPointer
	node.ID.traversalPointer = node

	// Support
	err = d.Decode(&node.Support)
	if err != nil {
		return err
	}

	// Children
	var length int
	err = d.Decode(&length)
	if err != nil {
		return err
	}

	node.FirstChildren = make([]SchemaNode, length, length)

	for i := range node.FirstChildren {
		node.FirstChildren[i] = SchemaNode{nil, node, nil, nil, nil, 0}
		err = node.FirstChildren[i].decodeGob(d, props)

		if err != nil {
			return err
		}
	}

	err = d.Decode(&length)
	if err != nil {
		return err
	}
	node.Children = make([]*SchemaNode, length, length)

	for i := range node.Children {
		node.Children[i] = &SchemaNode{nil, node, nil, nil, nil, 0}
		err = node.Children[i].decodeGob(d, props)

		if err != nil {
			return err
		}
	}

	return nil
}

// prefixContains checks if all properties of a given list are ancestors of a node
// internal! propertyPath *MUST* be sorted in sortOrder (i.e. descending support)
// thread-safe!
func (node *SchemaNode) prefixContains(propertyPath IList) bool {
	nextP := len(propertyPath) - 1                         // index of property expected to be seen next
	for cur := node; cur.parent != nil; cur = cur.parent { // walk from leaf towards root

		if cur.ID.SortOrder < propertyPath[nextP].SortOrder { // we already walked past the next expected property
			return false
		}
		if cur.ID == propertyPath[nextP] {
			nextP--
			if nextP < 0 { // we encountered all expected properties!
				return true
			}
		}
	}
	return false
}

func (node *SchemaNode) graphViz(minSup uint32) string {
	s := ""
	// // draw horizontal links
	// if node.nextSameID != nil && node.nextSameID.Support >= minSup {
	//  s += fmt.Sprintf("%v -> %v  [color=blue];\n", node, node.nextSameID)
	// }

	// // draw types
	// for k, v := range node.Types {
	//  s += fmt.Sprintf("%v -> \"%v\" [color=red,label=%v];\n", node, *k.Str, v)
	// }

	// draw children
	for _, child := range node.FirstChildren {
		if child.Support >= minSup {
			s += fmt.Sprintf("\"%p\" -> \"%p\" [label=%v;weight=%v];\n", node, &child, child.Support, child.ID.TotalCount)
			s += child.graphViz(minSup)
		}
	}
	if len(node.Children) > 0 {
		for _, child := range node.Children {
			if child.Support >= minSup {
				s += fmt.Sprintf("\"%p\" -> \"%p\" [label=%v;weight=%v];\n", node, child, child.Support, child.ID.TotalCount)
				s += child.graphViz(minSup)
			}
		}
	}

	return s
}

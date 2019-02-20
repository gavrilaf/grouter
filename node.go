package grouter

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	nodeRegular = iota + 1
	nodeVariable
	nodeCatchAll
	nodeCatchVariable
)

type node struct {
	name          string
	nodeType      int
	children      map[string]*node
	wildcardChild *node
	leaf          *leafNode
}

type leafNode struct {
	query url.Values
	value interface{}
}

func newRoot() *node {
	return &node{
		name:     "*",
		nodeType: nodeRegular,
	}
}

func (self *node) addChild(name string) (*node, error) {
	var newNode *node

	switch name[0:1] {
	case ":":
		name = strings.TrimSpace(name[1:])
		if len(name) == 0 {
			return nil, fmt.Errorf("Empty variable name")
		}

		wildcardChild := self.wildcardChild
		if wildcardChild != nil {
			if wildcardChild.name == name {
				return wildcardChild, nil
			} else {
				return nil, fmt.Errorf("Different variables on the same position: %v, %v", wildcardChild.name, name)
			}
		}

		newNode = &node{
			name:     name,
			nodeType: nodeVariable,
		}

		self.wildcardChild = newNode

	case "*":
		if self.wildcardChild != nil {
			return nil, fmt.Errorf("Variable and catchAll conflict: %v", self.wildcardChild.name)
		}

		name = strings.TrimSpace(name[1:])
		if len(name) == 0 {
			newNode = &node{
				name:     "",
				nodeType: nodeCatchAll,
			}
		} else {
			newNode = &node{
				name:     name,
				nodeType: nodeCatchVariable,
			}
		}

		self.wildcardChild = newNode

	default:
		if self.children == nil {
			self.children = make(map[string]*node)
		} else {
			child, ok := self.children[name]
			if ok {
				return child, nil
			}
		}

		newNode = &node{
			name:     name,
			nodeType: nodeRegular,
		}

		self.children[name] = newNode
	}

	return newNode, nil
}

func (self *node) getChild(name string) *node {
	fmt.Printf("getChild: %v, %v\n", self, name)

	child, ok := self.children[name]
	if !ok {
		return self.wildcardChild
	}

	return child
}

func (self *node) setLeaf(params url.Values, value interface{}) {
	self.leaf = &leafNode{
		query: params,
		value: value,
	}
}

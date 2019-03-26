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

type matchedLeaf struct {
	queryParams map[string]string
	value       interface{}
}

type leafNode struct {
	query    url.Values
	catchAll bool
	value    interface{}
}

type node struct {
	name          string
	nodeType      int
	children      map[string]*node
	wildcardChild *node
	leafs         []leafNode
}

// node

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
		name = strings.TrimSpace(name[1:])
		if len(name) == 0 {
			if self.wildcardChild != nil {
				if self.wildcardChild.nodeType == nodeCatchAll {
					return self.wildcardChild, nil
				} else {
					return nil, fmt.Errorf("Variable and catchAll conflict: %v", self.wildcardChild.name)
				}
			}

			newNode = &node{
				name:     "",
				nodeType: nodeCatchAll,
			}
		} else {
			if self.wildcardChild != nil {
				if self.wildcardChild.nodeType == nodeCatchVariable && self.wildcardChild.name == name {
					return self.wildcardChild, nil
				} else {
					return nil, fmt.Errorf("CatchAll variables conflict: %v", self.wildcardChild.name)
				}
			}

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
	child, ok := self.children[name]
	if !ok {
		return self.wildcardChild
	}

	return child
}

func (self *node) addLeaf(params url.Values, value interface{}) {
	_, catchAll := params["*"]
	if catchAll {
		delete(params, "*")
	}

	newLeaf := leafNode{
		query:    params,
		catchAll: catchAll,
		value:    value,
	}

	if self.leafs == nil {
		self.leafs = make([]leafNode, 1)
		self.leafs[0] = newLeaf
	} else {
		self.leafs = append(self.leafs, newLeaf)
	}
}

func (self *node) matchLeaf(params url.Values) *matchedLeaf {
	for _, leaf := range self.leafs {
		ok, parsedParams := leaf.matchQuery(params)
		if ok {
			return &matchedLeaf{
				queryParams: parsedParams,
				value:       leaf.value,
			}
		}
	}
	return nil
}

// leafNode

func (self *leafNode) matchQuery(params url.Values) (bool, map[string]string) {
	queryVars := make(map[string]string)

	if !self.catchAll && len(self.query) != len(params) {
		return false, queryVars
	}

	for key, v := range self.query {
		matchValue := strings.Join(v, ",")

		v, ok := params[key]
		if !ok {
			return false, queryVars
		}

		reqValue := strings.Join(v, ",")

		if matchValue == "*" {
			continue
		}

		if len(matchValue) > 1 && matchValue[0:1] == ":" {
			name := strings.TrimSpace(matchValue[1:])
			if len(name) > 0 {
				queryVars[name] = reqValue
			}

			continue
		}

		if matchValue != reqValue {
			return false, queryVars
		}
	}

	return true, queryVars
}

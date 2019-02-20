package grouter

import (
	"fmt"
	. "net/url"
	"strings"
)

type ParsedRoute struct {
	Url         string
	Pattern     string
	UrlParams   map[string]string
	QueryParams map[string]string
	Value       interface{}
}

type Router interface {
	AddRoute(url string, value interface{}) error
	Lookup(url string) (*ParsedRoute, error)
}

type router struct {
	hosts map[string]*node
}

func NewRouter() Router {
	return &router{hosts: make(map[string]*node)}
}

func (self *router) AddRoute(url string, value interface{}) error {
	u, err := Parse(url)
	if err != nil {
		return fmt.Errorf("Could not parse url %v, %v", url, err)
	}

	fmt.Printf("Add route - parsed url: %v, %v, %v\n", u.Hostname(), u.Path, u.Query())

	var root *node
	root, ok := self.hosts[u.Hostname()]
	if !ok {
		root = newRoot()
		self.hosts[u.Hostname()] = root
	}

	//fmt.Printf("Created root: %v\n", root)
	current := root

	comps := strings.Split(u.Path, "/")
	//fmt.Printf("Splitted components: %v\n", comps)
	for i := 0; i < len(comps); i++ {
		s := strings.TrimSpace(comps[i])
		if len(s) == 0 {
			continue
		}

		newNode, err := current.addChild(s)
		if err != nil {
			return fmt.Errorf("Could not add url %v, %v", url, err)
		}

		fmt.Printf("Add route - Current: %v, %v, %v\n", s, current, newNode)

		current = newNode
	}

	current.setLeaf(u.Query(), value)

	return nil
}

func (self *router) Lookup(url string) (*ParsedRoute, error) {
	u, err := Parse(url)
	if err != nil {
		return nil, fmt.Errorf("Could not parse url %v, %v", url, err)
	}

	fmt.Printf("Lookup - Parsed url: %v, %v, %v\n", u.Hostname(), u.Path, u.Query())

	current, ok := self.hosts[u.Hostname()]
	if !ok {
		return nil, nil
	}

	urlParams := make(map[string]string)

	comps := strings.Split(u.Path, "/")
	fmt.Printf("Lookup - Splitted components: %v\n", comps)
	for i := 0; i < len(comps); i++ {
		s := strings.TrimSpace(comps[i])
		if len(s) == 0 {
			continue
		}

		current = current.getChild(s)
		fmt.Printf("Lookup - Current: %v, %v\n", s, current)
		if current == nil {
			return nil, nil
		}

		if current.nodeType == nodeVariable {
			urlParams[current.name] = s
		}

		if current.nodeType == nodeCatchVariable {
			urlParams[current.name] = strings.Join(comps[i:], "/")
			break
		}

		if current.nodeType == nodeCatchAll {
			break
		}
	}

	if current != nil && current.leaf != nil {
		return &ParsedRoute{
			Url:         url,
			Pattern:     "",
			UrlParams:   urlParams,
			QueryParams: nil,
			Value:       current.leaf.value,
		}, nil
	}

	return nil, nil
}

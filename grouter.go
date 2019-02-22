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
	AddRoute(method string, url string, value interface{}) error
	Lookup(method string, url string) (*ParsedRoute, error)
}

type router struct {
	hosts map[string]map[string]*node
}

func NewRouter() Router {
	return &router{hosts: make(map[string]map[string]*node)}
}

func (self *router) AddRoute(method string, url string, value interface{}) error {
	method = strings.ToLower(method)
	url = strings.ToLower(url)

	u, err := Parse(url)
	if err != nil {
		return fmt.Errorf("Could not parse url %v, %v", url, err)
	}

	host, ok := self.hosts[u.Hostname()]
	if !ok {
		host = make(map[string]*node)
		self.hosts[u.Hostname()] = host
	}

	root, ok := host[method]
	if !ok {
		root = newRoot()
		host[method] = root
	}

	current := root

	comps := strings.Split(u.Path, "/")
	for i := 0; i < len(comps); i++ {
		s := strings.TrimSpace(comps[i])
		if len(s) == 0 {
			continue
		}

		newNode, err := current.addChild(s)
		if err != nil {
			return fmt.Errorf("Could not add url %v, %v", url, err)
		}
		current = newNode
	}

	current.addLeaf(u.Query(), value)

	return nil
}

func (self *router) Lookup(method string, url string) (*ParsedRoute, error) {
	method = strings.ToLower(method)
	url = strings.ToLower(url)

	u, err := Parse(url)
	if err != nil {
		return nil, fmt.Errorf("Could not parse url %v, %v", url, err)
	}

	host, ok := self.hosts[u.Hostname()]
	if !ok {
		return nil, nil
	}

	current, ok := host[method]
	if !ok {
		return nil, nil
	}

	urlParams := make(map[string]string)

	comps := strings.Split(u.Path, "/")
	for i := 0; i < len(comps); i++ {
		s := strings.TrimSpace(comps[i])
		if len(s) == 0 {
			continue
		}

		current = current.getChild(s)
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

	if current != nil {
		leaf := current.matchLeaf(u.Query())
		if leaf != nil {
			return &ParsedRoute{
				Url:         url,
				Pattern:     "",
				UrlParams:   urlParams,
				QueryParams: leaf.queryParams,
				Value:       leaf.value,
			}, nil
		}
	}

	return nil, nil
}

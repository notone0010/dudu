// Copyright 2022 NotOne Lv <aiphalv0010@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package dudu

// Model server http流量处理引擎

import (
	"log"
	"net/http"
	"sync"
	)

// ProxyHttpMux main logic.
type ProxyHttpMux struct {
	// Cluster contains role, enable and so on of the current server
	Cluster Cluster

	InsecureServingBindPort int
	SecureServingBindPort   int

	// LocalNetIFAddr is the network interface address the current local machine
	LocalNetIFAddr string

	mu       sync.RWMutex
	handlers HandlersChain
	m        map[string]http.Handler

	ph http.Handler
}

// NewProxyHttpMux new ProxyHttpMux.
func NewProxyHttpMux(cluster Cluster) *ProxyHttpMux {
	return &ProxyHttpMux{
		Cluster: cluster,
	}
}

// GetHandlers returns HandlersChain.
func (p *ProxyHttpMux) GetHandlers() HandlersChain {
	return p.handlers
}

// Use register middlewares.
func (p *ProxyHttpMux) Use(middlewares ...HandlerFunc) {
	p.handlers = append(p.handlers, middlewares...)
}

// Handle ...
func (p *ProxyHttpMux) Handle(pattern string, handler http.Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exist := p.m[pattern]; exist {
		panic("http: multiple registrations for " + pattern)
	}
	if p.m == nil {
		p.m = make(map[string]http.Handler)
	}
	p.m[pattern] = handler
}

// HandleFunc ...
func (p *ProxyHttpMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	p.Handle(pattern, handler)
}

// ServeHTTP implements the http.server interface.
func (p *ProxyHttpMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if router, ok := p.m[r.RequestURI]; ok {
		router.ServeHTTP(w, r)

		return
	}
	if p.ph == nil {
		log.Fatal("handle is nil, please call ProxyHttpMux.ProxyRequestHandler to register a proxy handler")

		return
	}
	p.ph.ServeHTTP(w, r)
}

// ProxyRequestHandler set proxy handler to http mux.
func (p *ProxyHttpMux) ProxyRequestHandler(handler http.HandlerFunc) {
	// newHandler := applyMiddlewares(handler, p.middlewares...)
	// p.ServeMux.Handle("/", newHandler)
	p.ph = handler
}

// Cluster contains configuration items related to cluster.
type Cluster struct {
	Enable         bool
	Role           string
	IsMasterHandle bool

	// Name is this cluster name
	Name string

	ClusterId string
	// LoadPolicy the current can choose load-balance policy when the role is master
	LoadPolicy string
}

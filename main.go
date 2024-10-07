package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleError(err)

	proxy := httputil.NewSingleHostReverseProxy(serverUrl)

	// Customize the transport for HTTPS handling (optional)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Not recommended for production
	}

	return &simpleServer{
		addr:  addr,
		proxy: proxy,
	}
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func handleError(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func (s *simpleServer) Address() string { return s.addr }

func (s *simpleServer) IsAlive() bool {
	// Perform a health check on the server by sending a GET request
	resp, err := http.Get(s.addr)
	if err != nil || resp.StatusCode >= 500 {
		return false
	}
	return true
}

func (s *simpleServer) Serve(rw http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(rw, r)
}

func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]

	// Check if the server is alive before forwarding the request
	for !server.IsAlive() {
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++

	// Reset roundRobinCount after a large number of requests to avoid overflow
	if lb.roundRobinCount > 1000000 {
		lb.roundRobinCount = 0
	}

	return server
}

func (lb *LoadBalancer) serveProxy(rw http.ResponseWriter, r *http.Request) {
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("forwarding your request to address %s\n", targetServer.Address())
	targetServer.Serve(rw, r)
}

func main() {
	servers := []Server{
		newSimpleServer("http://www.amazon.com"),
		newSimpleServer("http://www.samsung.com/"),
		newSimpleServer("http://www.x.com/"), // Change to valid internal URLs in a real use case
	}

	lb := NewLoadBalancer("8000", servers)

	handleRedirect := func(rw http.ResponseWriter, r *http.Request) {
		lb.serveProxy(rw, r)
	}

	http.HandleFunc("/", handleRedirect)

	fmt.Printf("serving requests at 'localhost:%s'\n", lb.port)
	err := http.ListenAndServe(":"+lb.port, nil)
	if err != nil {
		fmt.Printf("error starting server: %v\n", err)
		os.Exit(1)
	}
}

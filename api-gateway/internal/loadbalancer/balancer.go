package loadbalancer

import (
	"fmt"
	"net/url"
	"sync/atomic"

	consulapi "github.com/hashicorp/consul/api"
)

// LoadBalancer defines the interface for selecting a backend server.
type LoadBalancer interface {
	// Next takes a list of available service entries and returns the URL of the next one to use.
	Next(services []*consulapi.ServiceEntry) (*url.URL, error)
}

// RoundRobin is a simple round-robin load balancer.
// It keeps track of the index of the last used server.
type RoundRobin struct {
	current uint64 // Atomically updated index
}

// NewRoundRobin creates a new RoundRobin load balancer.
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{current: 0}
}

// Next implements the LoadBalancer interface for RoundRobin.
func (rr *RoundRobin) Next(services []*consulapi.ServiceEntry) (*url.URL, error) {
	if len(services) == 0 {
		return nil, fmt.Errorf("no available services for load balancing")
	}

	// I need to get the next index atomically.
	idx := atomic.AddUint64(&rr.current, 1)
	// Modulo operation to wrap around the list of services.
	selected := services[idx%uint64(len(services))]

	// I need to construct the URL for the selected service.
	// It's crucial to handle both Service.Address (if set) and AgentService.Address.
	// Prefer Service.Address if available, otherwise fallback to node address.
	address := selected.Service.Address
	if address == "" {
		address = selected.Node.Address // Fallback to node address
	}
	port := selected.Service.Port

	// I should default scheme to http if not specified in service meta, tags, etc.
	// In a real setup, this could be configurable via Consul service metadata.
	scheme := "http"

	targetURL := fmt.Sprintf("%s://%s:%d", scheme, address, port)
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target URL '%s': %w", targetURL, err)
	}

	return parsedURL, nil
}

// // Future: Implement other load balancing strategies
// type Random struct {}
// type LeastConnections struct {}

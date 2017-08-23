package builders

import (
	"github.com/docker/docker/api/types"
)

// Network creates a network with default values.
// Any number of network function builder can be pass to augment it.
func Network(builders ...func(*types.NetworkResource)) *types.NetworkResource {
	network := &types.NetworkResource{
		ID:     "networkID",
		Name:   "defaultNetworkID",
		Driver: "bridge",
		Scope:  "local",
	}

	for _, builder := range builders {
		builder(network)
	}

	return network
}

// NetworkID sets the network ID
func NetworkID(id string) func(*types.NetworkResource) {
	return func(network *types.NetworkResource) {
		network.ID = id
	}
}

// NetworkName sets the network name
func NetworkName(name string) func(*types.NetworkResource) {
	return func(network *types.NetworkResource) {
		network.Name = name
	}
}

// NetworkDriver sets the network driver
func NetworkDriver(driver string) func(network *types.NetworkResource) {
	return func(network *types.NetworkResource) {
		network.Driver = driver
	}
}

// NetworkLabels sets the network labels
func NetworkLabels(labels map[string]string) func(network *types.NetworkResource) {
	return func(network *types.NetworkResource) {
		network.Labels = labels
	}
}

// Ingress sets the network as ingress
func Ingress() func(*types.NetworkResource) {
	return func(network *types.NetworkResource) {
		network.Ingress = true
	}
}

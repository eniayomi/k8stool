package portforward

import (
	"io"
	"net"
)

// PortForwardOptions represents options for port forwarding
type PortForwardOptions struct {
	// Ports is a list of port mappings (local:remote)
	Ports []PortMapping `json:"ports"`

	// StopChannel is used to stop the port forwarding
	StopChannel chan struct{} `json:"-"`

	// ReadyChannel indicates when the port forward is ready
	ReadyChannel chan struct{} `json:"-"`

	// Streams configures the standard streams
	Streams Streams `json:"-"`
}

// PortMapping represents a port forwarding mapping
type PortMapping struct {
	// Local is the local port that will be used to forward to the pod
	Local uint16 `json:"local"`

	// Remote is the remote port that will be exposed from the pod
	Remote uint16 `json:"remote"`

	// Address is the local address to bind to
	Address string `json:"address,omitempty"`

	// Protocol is the protocol to use (tcp or udp)
	Protocol string `json:"protocol,omitempty"`
}

// Streams configures the standard streams for port forwarding
type Streams struct {
	// Out is used to write regular output
	Out io.Writer

	// ErrOut is used to write error output
	ErrOut io.Writer
}

// ForwardedPort represents an active forwarded port
type ForwardedPort struct {
	// Local is the local port that is being forwarded
	Local uint16

	// Remote is the remote port being forwarded to
	Remote uint16

	// Address is the local address bound to
	Address string

	// Protocol is the protocol being forwarded
	Protocol string

	// Listener is the local port listener
	Listener net.Listener
}

// PortForwardResult represents the result of a port forward operation
type PortForwardResult struct {
	// Ports contains information about the forwarded ports
	Ports []ForwardedPort `json:"ports"`

	// Error is any error that occurred during port forwarding
	Error error `json:"error,omitempty"`
}

// PortForwardDirection represents the direction of port forwarding
type PortForwardDirection string

const (
	// LocalToRemote forwards from local to remote
	LocalToRemote PortForwardDirection = "local-to-remote"
	// RemoteToLocal forwards from remote to local
	RemoteToLocal PortForwardDirection = "remote-to-local"
)

// PortForwardProtocol represents the protocol for port forwarding
type PortForwardProtocol string

const (
	// TCP protocol
	TCP PortForwardProtocol = "tcp"
	// UDP protocol
	UDP PortForwardProtocol = "udp"
)

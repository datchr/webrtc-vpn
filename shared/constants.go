// Package shared contains constants
package shared

const (
	// DataChannel names
	DataChannelControl = "vpn-control"
	DataChannelData    = "vpn-data"

	// Default values
	DefaultSignalingAddr    = "0.0.0.0:8443"
	DefaultTUNName          = "vpn0"
	DefaultIPPoolStart      = "10.0.0.1"
	DefaultIPPoolSize       = 254
	DefaultMaxPadding       = 512
	DefaultTimingJitterMs   = 50

	// Timeouts
	DefaultConnectTimeout   = 10 // seconds
	DefaultHeartbeatTimeout = 30 // seconds

	// Message types for control channel
	MsgTypePing       = "ping"
	MsgTypePong       = "pong"
	MsgTypeConfig     = "config"
	MsgTypeDisconnect = "disconnect"
)

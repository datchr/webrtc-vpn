// Package shared contains common types and utilities
package shared

import "time"

// ClientConfig represents client configuration
type ClientConfig struct {
	Server ServerClientConfig `yaml:"server"`
	TUN    TUNConfig          `yaml:"tun"`
	Crypto CryptoConfig       `yaml:"crypto"`
	Obfuscation ObfuscationConfig `yaml:"obfuscation"`
}

type ServerClientConfig struct {
	URL            string `yaml:"url"` // wss://server:port/connect
	ServerName     string `yaml:"server_name"`
	SkipVerifyTLS  bool   `yaml:"skip_verify_tls"`
	PreSharedKey   string `yaml:"psk"`
}

type TUNConfig struct {
	Name string `yaml:"name"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Server      ServerListenConfig `yaml:"server"`
	TUN         TUNConfig          `yaml:"tun"`
	IPPool      IPPoolConfig       `yaml:"ip_pool"`
	Crypto      CryptoConfig       `yaml:"crypto"`
	Obfuscation ObfuscationConfig  `yaml:"obfuscation"`
}

type ServerListenConfig struct {
	Addr     string `yaml:"addr"`       // 0.0.0.0:8443
	CertFile string `yaml:"cert_file"` // /path/to/cert.pem
	KeyFile  string `yaml:"key_file"`  // /path/to/key.pem
}

type IPPoolConfig struct {
	Start string `yaml:"start"`     // 10.0.0.1
	Size  int    `yaml:"size"`      // 254
	GW    string `yaml:"gateway"`   // 10.0.0.1
}

type CryptoConfig struct {
	EncryptionKey string `yaml:"encryption_key"` // base64-encoded 32 bytes
	PreSharedKey  string `yaml:"psk"`
}

type ObfuscationConfig struct {
	Enabled      bool          `yaml:"enabled"`
	Padding      bool          `yaml:"padding_enabled"`
	MaxPadding   int           `yaml:"max_padding_bytes"`  // 0-1024
	TimingJitter time.Duration `yaml:"timing_jitter_ms"`   // milliseconds
}

// SignalingMessage is exchanged during WebRTC setup
type SignalingMessage struct {
	Type     string `json:"type"` // "offer" or "answer"
	SDP      string `json:"sdp"`
	ClientID string `json:"client_id,omitempty"`
}

// ControlMessage is sent over the control DataChannel
type ControlMessage struct {
	Type      string `json:"type"` // "ping", "pong", "config", "disconnect"
	Timestamp int64  `json:"timestamp"`
	Payload   string `json:"payload,omitempty"`
}

// PeerInfo represents connected client info
type PeerInfo struct {
	ClientID      string
	AssignedIP    string
	ConnectedAt   time.Time
	LastHeartbeat time.Time
	BytesIn       int64
	BytesOut      int64
}

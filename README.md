# WebRTC VPN

Decentralized VPN based on WebRTC with obfuscation against DPI analysis (TSPU in RF).

## Architecture

```
┌─────────────────────┐           ┌──────────────────────┐
│  Windows Client     │           │   Ubuntu Server      │
│  ─────────────────  │           │  ──────────────────  │
│  • VPN App (Go)     │──WSS──→   │  • Signaling (Go)    │
│  • TUN/TAP драйвер  │   ←──WSS──│  • WebRTC Peer       │
│  • Обфускация       │           │  • TUN/TAP Interface │
│  • Шифрование       │           │  • IP pool (10.x.x.x)│
│                     │           │  • Маршрутизация     │
└─────────────────────┘           └──────────────────────┘
```

## Project Structure

```
webrtc-vpn/
├── client/                 # Windows VPN Client (client.go)
│   ├── main.go
│   ├── webrtc.go          # WebRTC peer connection
│   ├── signaling.go       # Connect to signaling server
│   ├── tuntap.go          # Virtual network interface (Windows)
│   ├── obfuscation.go     # DPI obfuscation
│   ├── config.yaml        # Client configuration
│   └── README.md
│
├── server/                 # Ubuntu Server (server.go)
│   ├── main.go
│   ├── signaling.go       # Signaling server (WSS, SDP exchange)
│   ├── tunnel.go          # TUN/TAP interface and routing
│   ├── peers.go           # Client connection management
│   ├── obfuscation.go     # DPI obfuscation
│   ├── config.yaml        # Server configuration
│   └── README.md
│
├── shared/                 # Shared utilities
│   ├── types.go           # Common data structures
│   ├── crypto.go          # Encryption/Decryption
│   └── constants.go       # Constants and defaults
│
├── go.mod
├── go.sum
├── README.md
└── ARCHITECTURE.md        # Detailed architecture
```

## Quick Start

### Prerequisites
- Go 1.21+
- On Windows client: Administrator privileges for TUN/TAP driver
- On Ubuntu server: `root` or `sudo` for TUN/TAP interface

### Server Setup (Ubuntu 192.168.3.41)

```bash
cd server
go mod download
go build -o vpn-server
sudo ./vpn-server
```

### Client Setup (Windows)

```bash
cd client
go mod download
go build -o vpn-client.exe
vpn-client.exe  # Run as Administrator
```

## Configuration

Each component has a `config.yaml` file:

### Server Config (server/config.yaml)

```yaml
server:
  signaling_addr: "0.0.0.0:8443"  # WSS server address
  tun_name: "vpn0"
  ip_pool_start: "10.0.0.1"
  ip_pool_size: 254
  
crypto:
  encryption_key: "your-base64-encoded-32-byte-key"
  
obfuscation:
  enabled: true
  padding_enabled: true
```

### Client Config (client/config.yaml)

```yaml
server:
  url: "wss://192.168.3.41:8443/connect"
  
tun:
  name: "vpn0"
  
obfuscation:
  enabled: true
```

## Security Features

- ✅ WSS (WebSocket Secure) for encrypted signaling
- ✅ AES-256-GCM for traffic encryption
- ✅ DTLS (Datagram TLS) from WebRTC
- ✅ DPI obfuscation (packet padding, timing randomization)
- ✅ Client authentication via pre-shared key
- ✅ Perfect Forward Secrecy (ephemeral keys)

## Development Phases

1. **Phase 1**: Signaling server + basic client connection
2. **Phase 2**: Improved client with proper error handling
3. **Phase 3**: TUN/TAP interface implementation
4. **Phase 4**: DPI obfuscation and hardening

## Deployment

### Production Server (France)

```bash
# On remote Ubuntu server
git clone https://github.com/datchr/webrtc-vpn.git
cd webrtc-vpn/server
cp config.yaml.example config.yaml
# Edit config.yaml with production settings
go build -o vpn-server
sudo systemctl enable vpn-server
sudo systemctl start vpn-server
```

## License

MIT

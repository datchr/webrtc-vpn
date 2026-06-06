# Server - Ubuntu VPN Backend

## For Deployment on Ubuntu Server (192.168.3.41 and later on remote)

### Setup

```bash
cd server
go mod download
go build -o vpn-server
```

### Run with TUN/TAP interface

```bash
# First, install wireguard-tools
sudo apt-get install wireguard-tools

# Run as root (needed for TUN/TAP)
sudo ./vpn-server
```

### Configuration

Edit `config.yaml`:

```yaml
server:
  addr: "0.0.0.0:8443"
  cert_file: "/path/to/cert.pem"
  key_file: "/path/to/key.pem"

tun:
  name: "vpn0"

ip_pool:
  start: "10.0.0.1"
  size: 254
  gateway: "10.0.0.1"

crypto:
  encryption_key: "<base64-32-byte-key>"
  psk: "your-pre-shared-key"

obfuscation:
  enabled: true
  padding_enabled: true
  max_padding_bytes: 512
  timing_jitter_ms: 50
```

### Generate SSL Certificates (for testing)

```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
```

### Logs

Check system logs:

```bash
journalctl -u vpn-server -f
```

## Architecture

Server components:
1. **signaling.go** - WSS endpoint for SDP offer/answer exchange
2. **tunnel.go** - TUN interface creation and packet routing
3. **peers.go** - Client connection management and IP pool allocation
4. **obfuscation.go** - DPI evasion features
5. **main.go** - Entry point and WebRTC setup

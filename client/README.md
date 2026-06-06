# Client - Windows VPN Application

## For Windows 10/11 (Administrator required)

### Prerequisites

- Go 1.21+
- Administrator privileges (for TUN/TAP driver)
- OpenVPN TAP driver or WireGuard TUN driver installed

### Setup

```bash
cd client
go mod download
go build -o vpn-client.exe
```

### Run

```bash
# Right-click and select "Run as Administrator"
vpn-client.exe
```

### Configuration

Edit `config.yaml`:

```yaml
server:
  url: "wss://192.168.3.41:8443/connect"
  server_name: "vpn.example.com"
  skip_verify_tls: true  # Only for testing!
  psk: "changeme"

tun:
  name: "vpn0"

crypto:
  encryption_key: "<same as server>"

obfuscation:
  enabled: true
```

### Verify Connection

```powershell
# Check TUN interface
Get-NetAdapter | Select-Object Name, Status

# Ping VPN gateway
ping 10.0.0.1

# Check IP configuration
route print
```

### Logs

Logs are written to `vpn-client.log`

## Architecture

Client components:
1. **main.go** - Application entry point and lifecycle
2. **webrtc.go** - WebRTC PeerConnection setup
3. **signaling.go** - Connect to signaling server, SDP exchange
4. **tuntap.go** - Windows TUN/TAP interface integration
5. **obfuscation.go** - DPI evasion features

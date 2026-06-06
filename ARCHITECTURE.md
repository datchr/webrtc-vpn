# WebRTC VPN - Detailed Architecture

## Overview

This VPN system uses WebRTC DataChannels for peer-to-peer encrypted communication, combined with TUN/TAP virtual network interfaces to provide transparent VPN functionality.

## Components

### 1. Signaling Server (server/signaling.go)

**Purpose**: Facilitate initial connection between client and server

**Features**:
- WSS (WebSocket Secure) endpoint for SDP offer/answer exchange
- Client registration and deregistration
- Minimal state - stateless design for scalability
- Support for multiple simultaneous connections

**Flow**:
1. Client connects to `/connect` WSS endpoint
2. Client sends WebRTC SDP Offer
3. Server creates PeerConnection and sends SDP Answer
4. Both sides exchange ICE candidates
5. DataChannel opens → tunnel begins

### 2. WebRTC Transport (client/webrtc.go, server/tunnel.go)

**Purpose**: Encrypted data channel for packets

**Features**:
- DTLS encryption (from WebRTC)
- Multiple DataChannels for different traffic types
- ICE candidate filtering (only IPv4 for now)
- NAT traversal via STUN/TURN servers

**Channels**:
- `vpn-control`: Control messages (ping, configuration)
- `vpn-data`: User traffic (encrypted packets)

### 3. TUN/TAP Interface

**Purpose**: Virtual network adapter for transparent VPN

#### Windows (client/tuntap.go)
- Uses WinTUN driver via `golang.zx2c4.com/wireguard`
- Creates virtual interface `vpn0`
- Assigns IP from server's pool (e.g., 10.0.0.2)
- Reads/writes IP packets directly

#### Linux (server/tunnel.go)
- Uses `/dev/net/tun` directly
- Creates `vpn0` interface
- Configures with `ip` command
- NAT rules to route traffic

### 4. Packet Processing

**Client side**:
```
Application
    ↓
OS Network Stack
    ↓
TUN Interface (vpn0)
    ↓
Read IP packet
    ↓
Encrypt + Send via WebRTC DataChannel
    ↓
Server
```

**Server side**:
```
Receive from WebRTC DataChannel
    ↓
Decrypt IP packet
    ↓
Inject into TUN Interface (vpn0)
    ↓
OS Network Stack
    ↓
Internet / Local Network
    ↓
Response comes back
    ↓
TUN Interface (vpn0)
    ↓
Read response packet
    ↓
Encrypt + Send to Client via WebRTC
    ↓
Client TUN Interface receives packet
    ↓
Application
```

### 5. DPI Obfuscation (shared/obfuscation.go)

**Defense layers**:

1. **Protocol Level**
   - Use WSS (WebSocket over TLS) instead of raw WebRTC
   - Appears as normal HTTPS traffic
   - TLS fingerprint randomization

2. **Packet Level**
   - Random padding up to 1KB per packet
   - Timing jitter (5-50ms random delays)
   - Packet size randomization

3. **Traffic Analysis**
   - Packet scheduling to break patterns
   - Continuous dummy packets during idle (optional)
   - Constant bandwidth simulation (optional)

4. **Encryption**
   - AES-256-GCM for all data
   - Key derivation via HKDF
   - IV randomization

## Security Model

### Authentication

1. **Pre-shared Key (PSK)**
   - Client and server share a secret
   - Used for HMAC-based authentication
   - PSK rotatable per configuration reload

2. **Certificate Pinning** (future)
   - Pin server certificate SHA-256
   - Prevents MITM attacks

### Encryption

**In Transit**:
- TLS 1.3 for WSS signaling
- DTLS for WebRTC DataChannel
- AES-256-GCM for payload encryption

**At Rest**: N/A (no data stored)

### Client Isolation

- Each client gets unique IP from pool
- Network namespace isolation (server side)
- No inter-client traffic allowed

## Scalability

### Single Server
- Supports ~100 concurrent connections with 4GB RAM
- Each connection: ~10MB resident (Pion WebRTC overhead)
- Bandwidth: limited by server's ISP

### Multiple Servers

Future architecture:
```
Clients → Load Balancer (anycast) → Multiple Signaling Servers
                                   → Each with own TUN interface
                                   → Or shared routing backend (not yet implemented)
```

## Performance

**Expected throughput**:
- Latency: +10-20ms vs direct connection
- Bandwidth: ~95% of available (WebRTC overhead ~5%)
- CPU: ~30% on single core per 100Mbps

## Future Enhancements

1. **Transport Diversity**
   - QUIC-based transport as alternative
   - TCP fallback mode
   - HTTP/2 server-push for downlink

2. **Advanced DPI Evasion**
   - Machine learning-based traffic shaping
   - Protocol mimicry (fake BitTorrent, etc.)
   - Multi-hop architecture

3. **Performance**
   - Kernel-space VPN driver
   - eBPF-based packet processing
   - Hardware acceleration

4. **Reliability**
   - Automatic reconnection
   - Connection migration
   - Redundant server connections

## Building

```bash
# Server
cd server
go build -o vpn-server -ldflags="-s -w"

# Client (Windows)
cd client
go build -o vpn-client.exe -ldflags="-s -w"

# Cross-compile from Linux to Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o vpn-client.exe
```

## Testing

```bash
# Start server
sudo ./server/vpn-server

# Start client (admin shell on Windows)
./client/vpn-client.exe

# Test connection
ping 10.0.0.1  # Ping server's VPN IP

# Test throughput
iperf3 -s            # On server
iperf3 -c 10.0.0.1   # From client
```

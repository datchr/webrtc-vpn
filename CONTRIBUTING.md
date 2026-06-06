# Contributing

We welcome contributions! Please follow these guidelines:

## Development Setup

```bash
git clone https://github.com/datchr/webrtc-vpn.git
cd webrtc-vpn
go mod download
```

## Running Tests

```bash
go test ./...
```

## Building

### Server
```bash
cd server
go build -o vpn-server -ldflags="-s -w"
```

### Client (Windows)
```bash
cd client
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o vpn-client.exe -ldflags="-s -w"
```

## Code Style

- Use `gofmt` for formatting
- Run `go vet ./...` before submitting
- Add comments for exported functions
- Keep error messages descriptive

## Security

- Never commit credentials or keys
- Use environment variables for sensitive data
- Test DPI evasion capabilities
- Document any known limitations

## Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add/update tests
5. Submit PR with detailed description

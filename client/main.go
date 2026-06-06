package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
	"crypto/shared"
)

var (
	configPath = flag.String("config", "config.yaml", "Path to config file")
	client     *VPNClient
)

type VPNClient struct {
	config      *shared.ClientConfig
	cipher      *shared.Cipher
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	logger      *log.Logger
	tunDevice   interface{} // Will be *TUNDevice on Windows
	signaling   *SignalingClient
}

func main() {
	flag.Parse()

	// Setup logging
	logFile, err := os.OpenFile("vpn-client.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "[VPN] ", log.LstdFlags|log.Lshortfile)

	// Load configuration
	cfg, err := loadConfig(*configPath)
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Initialize cipher
	cipher, err := shared.NewCipher(cfg.Crypto.EncryptionKey)
	if err != nil {
		logger.Fatalf("Failed to initialize cipher: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client = &VPNClient{
		config: cfg,
		cipher: cipher,
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		client.logger.Println("Received shutdown signal")
		client.shutdown()
	}()

	if err := client.run(); err != nil {
		client.logger.Fatalf("Client error: %v", err)
	}

	client.wg.Wait()
}

func (c *VPNClient) run() error {
	c.logger.Println("=== WebRTC VPN Client Started ===")

	// Connect to signaling server
	sig, err := NewSignalingClient(c.config.Server.URL, c.logger)
	if err != nil {
		return fmt.Errorf("failed to create signaling client: %w", err)
	}
	c.signaling = sig

	// Create TUN device
	tun, err := NewTUNDevice(c.config.TUN.Name, c.logger)
	if err != nil {
		return fmt.Errorf("failed to create TUN device: %w", err)
	}
	defer tun.Close()
	c.tunDevice = tun

	// Setup WebRTC peer
	wrtc, err := NewWebRTCPeer(c.config, c.logger)
	if err != nil {
		return fmt.Errorf("failed to setup WebRTC: %w", err)
	}

	// Perform signaling handshake
	if err := sig.Handshake(c.ctx, wrtc); err != nil {
		return fmt.Errorf("signaling handshake failed: %w", err)
	}

	c.logger.Println("Successfully connected to VPN server")

	// Start packet processing
	c.wg.Add(1)
	go c.handlePackets(wrtc, tun)

	return nil
}

func (c *VPNClient) handlePackets(wrtc *WebRTCPeer, tun interface{}) {
	defer c.wg.Done()
	// TODO: Implement packet processing
	c.logger.Println("Packet handler started")
}

func (c *VPNClient) shutdown() {
	c.logger.Println("Shutting down gracefully...")
	c.cancel()
	if c.signaling != nil {
		c.signaling.Close()
	}
	if c.tunDevice != nil {
		if dev, ok := c.tunDevice.(*TUNDevice); ok {
			dev.Close()
		}
	}
	c.logger.Println("Shutdown complete")
}

func loadConfig(path string) (*shared.ClientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg shared.ClientConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.TUN.Name == "" {
		cfg.TUN.Name = shared.DefaultTUNName
	}
	if cfg.Obfuscation.MaxPadding == 0 {
		cfg.Obfuscation.MaxPadding = shared.DefaultMaxPadding
	}

	return &cfg, nil
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/golang/glog"
	"gopkg.in/yaml.v3"
	"crypto/shared"
)

var (
	configPath = flag.String("config", "config.yaml", "Path to config file")
	server     *VPNServer
)

type VPNServer struct {
	config      *shared.ServerConfig
	cipher      *shared.Cipher
	peerManager *PeerManager
	tunDevice   *TUNDevice
	webrtcAPI   *webrtc.API
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func main() {
	flag.Parse()
	defer glog.Flush()

	// Load configuration
	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize cipher
	cipher, err := shared.NewCipher(cfg.Crypto.EncryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize cipher: %v", err)
	}

	// Create WebRTC API
	settingsEngine := webrtc.SettingEngine{}
	api := webrtc.NewAPI(webrtc.WithSettingEngine(settingsEngine))

	// Initialize TUN device
	tun, err := NewTUNDevice(cfg.TUN.Name, cfg.IPPool)
	if err != nil {
		log.Fatalf("Failed to create TUN device: %v", err)
	}
	defer tun.Close()

	// Initialize peer manager
	peerMgr := NewPeerManager(cfg.IPPool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server = &VPNServer{
		config:      cfg,
		cipher:      cipher,
		peerManager: peerMgr,
		tunDevice:   tun,
		webrtcAPI:   api,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\nShutting down...")
		server.shutdown()
	}()

	// Start signaling server
	if err := server.startSignalingServer(); err != nil {
		log.Fatalf("Failed to start signaling server: %v", err)
	}

	log.Printf("VPN Server started on %s", cfg.Server.Addr)

	// Start TUN packet handler
	server.wg.Add(1)
	go server.handleTUNPackets()

	server.wg.Wait()
}

func (s *VPNServer) startSignalingServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/connect", s.handleConnection)
	mux.HandleFunc("/health", s.handleHealth)

	server := &http.Server{
		Addr:         s.config.Server.Addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := server.ListenAndServeTLS(
			s.config.Server.CertFile,
			s.config.Server.KeyFile,
		); err != nil && err != http.ErrServerClosed {
			glog.Errorf("Server error: %v", err)
		}
	}()

	go func() {
		<-s.ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	return nil
}

func (s *VPNServer) handleConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For now, this is a placeholder
	// In production, upgrade to WebSocket here
	glog.V(1).Infof("New connection from %s", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
}

func (s *VPNServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","peers":%d}`, s.peerManager.Count())
}

func (s *VPNServer) handleTUNPackets() {
	defer s.wg.Done()
	// TODO: Implement packet processing from TUN interface
}

func (s *VPNServer) shutdown() {
	s.cancel()
	s.tunDevice.Close()
	s.wg.Wait()
	glog.Infof("Server shutdown complete")
}

func loadConfig(path string) (*shared.ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg shared.ServerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Server.Addr == "" {
		cfg.Server.Addr = shared.DefaultSignalingAddr
	}
	if cfg.TUN.Name == "" {
		cfg.TUN.Name = shared.DefaultTUNName
	}
	if cfg.IPPool.Start == "" {
		cfg.IPPool.Start = shared.DefaultIPPoolStart
	}
	if cfg.IPPool.Size == 0 {
		cfg.IPPool.Size = shared.DefaultIPPoolSize
	}
	if cfg.Obfuscation.MaxPadding == 0 {
		cfg.Obfuscation.MaxPadding = shared.DefaultMaxPadding
	}

	return &cfg, nil
}

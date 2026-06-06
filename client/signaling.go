package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"crypto/shared"
)

type SignalingClient struct {
	url        string
	ws         *websocket.Conn
	logger     *log.Logger
	shutdownCh chan struct{}
}

func NewSignalingClient(serverURL string, logger *log.Logger) (*SignalingClient, error) {
	if serverURL == "" {
		return nil, fmt.Errorf("server URL is required")
	}

	return &SignalingClient{
		url:        serverURL,
		logger:     logger,
		shutdownCh: make(chan struct{}),
	}, nil
}

func (sc *SignalingClient) Handshake(ctx context.Context, wrtc *WebRTCPeer) error {
	// Parse URL and create WebSocket connection
	u, err := url.Parse(sc.url)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	sc.logger.Printf("Connecting to signaling server: %s", sc.url)

	// Create WebSocket connection with timeout
	dialer := websocket.Dialer{
		HandshakeTimeout: time.Duration(shared.DefaultConnectTimeout) * time.Second,
	}

	header := make(map[string][]string)
	header["X-Client-ID"] = []string{"client-" + fmt.Sprint(time.Now().Unix())}

	ws, _, err := dialer.DialContext(ctx, u.String(), header)
	if err != nil {
		return fmt.Errorf("failed to connect to signaling server: %w", err)
	}
	sc.ws = ws
	defer ws.Close()

	sc.logger.Println("Connected to signaling server")

	// Create WebRTC offer
	offer, err := wrtc.CreateOffer()
	if err != nil {
		return fmt.Errorf("failed to create offer: %w", err)
	}

	// Send offer to server
	msg := shared.SignalingMessage{
		Type: "offer",
		SDP:  offer.SDP,
	}

	if err := ws.WriteJSON(msg); err != nil {
		return fmt.Errorf("failed to send offer: %w", err)
	}

	sc.logger.Println("Sent offer, waiting for answer...")

	// Receive answer
	var answer shared.SignalingMessage
	if err := ws.ReadJSON(&answer); err != nil {
		return fmt.Errorf("failed to receive answer: %w", err)
	}

	if answer.Type != "answer" {
		return fmt.Errorf("expected answer, got: %s", answer.Type)
	}

	sc.logger.Println("Received answer from server")

	// Set remote description
	if err := wrtc.SetRemoteDescription(answer.SDP); err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	sc.logger.Println("Handshake complete, waiting for connection...")

	// Wait for data channels to open or timeout
	select {
	case <-wrtc.readyCh:
		sc.logger.Println("Data channels opened successfully")
		return nil
	case <-time.After(time.Duration(shared.DefaultConnectTimeout) * time.Second):
		return fmt.Errorf("connection timeout")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (sc *SignalingClient) Close() error {
	close(sc.shutdownCh)
	if sc.ws != nil {
		return sc.ws.Close()
	}
	return nil
}

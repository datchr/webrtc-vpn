package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pion/webrtc/v4"
	"crypto/shared"
)

type WebRTCPeer struct {
	peerConn    *webrtc.PeerConnection
	logger      *log.Logger
	readyCh     chan struct{}
	mu          sync.Mutex
	dataChannel *webrtc.DataChannel
	controlCh   *webrtc.DataChannel
}

func NewWebRTCPeer(config *shared.ClientConfig, logger *log.Logger) (*WebRTCPeer, error) {
	// Create WebRTC API with optimized settings
	settingsEngine := webrtc.SettingEngine{}
	settingsEngine.SetNetworkTypes([]webrtc.NetworkType{webrtc.NetworkTypeUDP4})

	api := webrtc.NewAPI(webrtc.WithSettingEngine(settingsEngine))

	// Create peer connection
	peerConn, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create PeerConnection: %w", err)
	}

	wrtc := &WebRTCPeer{
		peerConn: peerConn,
		logger:   logger,
		readyCh:  make(chan struct{}),
	}

	// Set up connection state change handlers
	peerConn.OnConnectionStateChange(func(connState webrtc.PeerConnectionState) {
		logger.Printf("Connection state changed: %s", connState.String())
	})

	// Set up ICE connection state change handlers
	peerConn.OnICEConnectionStateChange(func(connState webrtc.ICEConnectionState) {
		logger.Printf("ICE connection state changed: %s", connState.String())
	})

	// Create data channels for control and data
	controlCh, err := peerConn.CreateDataChannel(shared.DataChannelControl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create control channel: %w", err)
	}
	wrtc.controlCh = controlCh
	wrtc.setupControlChannel(controlCh)

	dataCh, err := peerConn.CreateDataChannel(shared.DataChannelData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create data channel: %w", err)
	}
	wrtc.dataChannel = dataCh
	wrtc.setupDataChannel(dataCh)

	return wrtc, nil
}

func (w *WebRTCPeer) setupControlChannel(ch *webrtc.DataChannel) {
	ch.OnOpen(func() {
		w.logger.Println("Control channel opened")
	})

	ch.OnMessage(func(msg webrtc.DataChannelMessage) {
		var ctrlMsg shared.ControlMessage
		if err := json.Unmarshal(msg.Data, &ctrlMsg); err != nil {
			w.logger.Printf("Failed to parse control message: %v", err)
			return
		}

		switch ctrlMsg.Type {
		case shared.MsgTypePing:
			// Send pong back
			pong := shared.ControlMessage{
				Type:      shared.MsgTypePong,
				Timestamp: time.Now().Unix(),
			}
			if data, err := json.Marshal(pong); err == nil {
				ch.Send(webrtc.DataChannelMessage{Data: data})
			}
		case shared.MsgTypeDisconnect:
			w.logger.Println("Received disconnect command from server")
		}
	})
}

func (w *WebRTCPeer) setupDataChannel(ch *webrtc.DataChannel) {
	ch.OnOpen(func() {
		w.logger.Println("Data channel opened")
		close(w.readyCh)
	})

	ch.OnMessage(func(msg webrtc.DataChannelMessage) {
		// TODO: Process decrypted packets from TUN interface
	})
}

func (w *WebRTCPeer) CreateOffer() (*webrtc.SessionDescription, error) {
	offer, err := w.peerConn.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	if err := w.peerConn.SetLocalDescription(offer); err != nil {
		return nil, fmt.Errorf("failed to set local description: %w", err)
	}

	return &offer, nil
}

func (w *WebRTCPeer) SetRemoteDescription(sdp string) error {
	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdp,
	}

	return w.peerConn.SetRemoteDescription(answer)
}

func (w *WebRTCPeer) SendPacket(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.dataChannel == nil || w.dataChannel.ReadyState() != webrtc.DataChannelStateOpen {
		return fmt.Errorf("data channel not ready")
	}

	return w.dataChannel.Send(webrtc.DataChannelMessage{Data: data})
}

func (w *WebRTCPeer) Close() error {
	return w.peerConn.Close()
}

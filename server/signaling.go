package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"crypto/shared"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: restrict in production
	},
}

type SignalingSession struct {
	conn         *websocket.Conn
	peerConn     *webrtc.PeerConnection
	clientID     string
	mu           sync.Mutex
}

func (s *VPNServer) handleSignaling(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to upgrade connection:", err)
		return
	}

	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		clientID = generateClientID()
	}

	session := &SignalingSession{
		conn:     ws,
		clientID: clientID,
	}

	defer ws.Close()

	// Create WebRTC PeerConnection
	peerConn, err := s.webrtcAPI.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		fmt.Println("Failed to create PeerConnection:", err)
		return
	}
	session.peerConn = peerConn
	defer peerConn.Close()

	// Allocate IP for client
	ip, err := s.peerManager.AllocateIP(clientID)
	if err != nil {
		fmt.Println("Failed to allocate IP:", err)
		return
	}

	// Create data channels
	ctrlChan, err := peerConn.CreateDataChannel(shared.DataChannelControl, nil)
	if err != nil {
		fmt.Println("Failed to create control channel:", err)
		return
	}

	dataChan, err := peerConn.CreateDataChannel(shared.DataChannelData, nil)
	if err != nil {
		fmt.Println("Failed to create data channel:", err)
		return
	}

	// Handle data channel open
	dataChan.OnOpen(func() {
		fmt.Printf("Data channel opened for %s (%s)\n", clientID, ip)
	})

	dataChan.OnMessage(func(msg webrtc.DataChannelMessage) {
		// TODO: Process encrypted packets and inject into TUN
	})

	// Handle incoming WebRTC offer
	var msg shared.SignalingMessage
	if err := ws.ReadJSON(&msg); err != nil {
		fmt.Println("Failed to read offer:", err)
		return
	}

	if msg.Type != "offer" {
		fmt.Println("Expected offer, got:", msg.Type)
		return
	}

	// Set remote description (offer)
	if err := peerConn.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  msg.SDP,
	}); err != nil {
		fmt.Println("Failed to set remote description:", err)
		return
	}

	// Create and send answer
	answer, err := peerConn.CreateAnswer(nil)
	if err != nil {
		fmt.Println("Failed to create answer:", err)
		return
	}

	if err := peerConn.SetLocalDescription(answer); err != nil {
		fmt.Println("Failed to set local description:", err)
		return
	}

	// Send answer back
	response := shared.SignalingMessage{
		Type:     "answer",
		SDP:      peerConn.LocalDescription().SDP,
		ClientID: clientID,
	}

	if err := ws.WriteJSON(response); err != nil {
		fmt.Println("Failed to send answer:", err)
		return
	}

	// Keep connection open
	for {
		var controlMsg shared.ControlMessage
		if err := ws.ReadJSON(&controlMsg); err != nil {
			break
		}
		if controlMsg.Type == shared.MsgTypePing {
			ws.WriteJSON(shared.ControlMessage{
				Type: shared.MsgTypePong,
			})
		}
	}

	s.peerManager.ReleaseIP(clientID)
}

func generateClientID() string {
	return fmt.Sprintf("client_%d", int64(time.Now().Unix()))
}

var time = struct {
	Now func() int64
}{
	Now: func() int64 {
		return 0
	},
}

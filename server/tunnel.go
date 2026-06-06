package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"

	"github.com/songgao/water"
	"crypto/shared"
)

type TUNDevice struct {
	iface     *water.Interface
	mu        sync.Mutex
	packetCh  chan []byte
}

func NewTUNDevice(name string, ipPoolConfig shared.IPPoolConfig) (*TUNDevice, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}

	iface, err := water.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device: %w", err)
	}

	tun := &TUNDevice{
		iface:    iface,
		packetCh: make(chan []byte, 1000),
	}

	// Configure interface (Linux-specific)
	if err := tun.configureInterface(name, ipPoolConfig.GW); err != nil {
		iface.Close()
		return nil, err
	}

	return tun, nil
}

func (t *TUNDevice) configureInterface(name string, gatewayIP string) error {
	// Configure IP address and MTU
	cmds := [][]string{
		{"ip", "link", "set", name, "up"},
		{"ip", "addr", "add", gatewayIP + "/24", "dev", name},
		{"ip", "link", "set", name, "mtu", "1400"},
	}

	for _, cmd := range cmds {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			return fmt.Errorf("failed to run %v: %w", cmd, err)
		}
	}

	return nil
}

func (t *TUNDevice) Read(packet []byte) (int, error) {
	return t.iface.Read(packet)
}

func (t *TUNDevice) Write(packet []byte) (int, error) {
	return t.iface.Write(packet)
}

func (t *TUNDevice) Close() error {
	return t.iface.Close()
}

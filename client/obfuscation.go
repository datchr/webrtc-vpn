package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"
	"crypto/shared"
)

type ClientObfuscator struct {
	config shared.ObfuscationConfig
	logger *log.Logger
}

func NewClientObfuscator(config shared.ObfuscationConfig, logger *log.Logger) *ClientObfuscator {
	return &ClientObfuscator{
		config: config,
		logger: logger,
	}
}

func (co *ClientObfuscator) ApplyObfuscation(packet []byte) []byte {
	if !co.config.Enabled {
		return packet
	}

	// Add random padding
	if co.config.Padding {
		packet = co.addPadding(packet)
	}

	// Add timing jitter
	if co.config.TimingJitter > 0 {
		co.applyTimingJitter()
	}

	return packet
}

func (co *ClientObfuscator) addPadding(payload []byte) []byte {
	padSize := make([]byte, 1)
	rand.Read(padSize)
	size := int(padSize[0]) % (co.config.MaxPadding + 1)

	padding := make([]byte, size)
	rand.Read(padding)

	return append(payload, padding...)
}

func (co *ClientObfuscator) applyTimingJitter() {
	// Apply random delay to break timing patterns
	jitterMs := make([]byte, 1)
	rand.Read(jitterMs)
	delay := time.Duration(int(jitterMs[0])%int(co.config.TimingJitter)) * time.Millisecond
	time.Sleep(delay)
}

func (co *ClientObfuscator) String() string {
	return fmt.Sprintf("ClientObfuscator{enabled:%v, padding:%v, jitter:%dms}",
		co.config.Enabled, co.config.Padding, co.config.TimingJitter)
}

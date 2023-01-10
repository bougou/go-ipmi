package ipmi_test

import (
	"bytes"
	"context"
	"github.com/xstp/go-ipmi"
	"net"
	"sync"
	"testing"
	"time"
)

func TestUDPClientRaceCondition(t *testing.T) {
	// Start a mock UDP server that just echoes back whatever it receives
	ln, err := net.ListenPacket("udp", "0.0.0.0:31337")
	if err != nil {
		t.Fatalf("failed to start mock UDP server: %v", err)
	}
	defer ln.Close()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, _, err := ln.ReadFrom(buf)
			if err != nil {
				return
			}
			ln.WriteTo(buf[:n], ln.LocalAddr())
		}
	}()

	// Create a client that sends requests to the mock server
	client := ipmi.NewUDPClient("0.0.0.0", 31337)
	client.SetTimeout(time.Second)
	client.SetBufferSize(1024)

	// Send 100 requests concurrently
	var wg sync.WaitGroup
	for i := 0; i < 400; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 2 seconds should be enough to trigger
			cancelCtx, cancel := context.WithTimeout(context.Background(), 80*time.Second)
			defer cancel()

			_, err := client.Exchange(cancelCtx, bytes.NewReader([]byte("hello")))
			if err != nil {
				t.Errorf("failed to exchange data: %v", err)
			}

			select {
			case <-cancelCtx.Done():
				t.Errorf("context was cancelled: %v", cancelCtx.Err())
				return
			default:

			}
		}()
	}
	wg.Wait()
}

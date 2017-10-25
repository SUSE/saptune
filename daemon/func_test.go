package daemon

import (
	"os"
	"testing"
	"time"
)

func TestFunc(t *testing.T) {
	if !system.IsUserRoot() {
                t.Skip("the test requires root access")
        }
	// Start RPC server
	server := new(Server)
	if err := server.Listen(); err != nil {
		t.Fatal(err)
	}
	go server.MainLoop()
	// Expect RPC server to be ready in a second
	time.Sleep(1 * time.Second)
	client := new(Client)
	// Test all RPC functions
	if err := client.SetForceLatency(1); err != nil {
		t.Fatal(err)
	}
	if err := client.StopForceLatency(); err != nil {
		t.Fatal(err)
	}
	// Repeatedly shutdown server should not carry negative consequence
	server.Shutdown()
	server.Shutdown()
	// Server should shut down in a second
	time.Sleep(1 * time.Second)
	if err := client.StopForceLatency(); err == nil {
		t.Fatal("did not shutdown")
	}
}

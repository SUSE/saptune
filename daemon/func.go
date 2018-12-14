package daemon

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"
)

// FunctionHost hosts saptune daemon's RPC functions that are invoked by RPC client.
type FunctionHost struct {
	forceLatency      chan int    // forceLatency channel stores the latest setting daemon should apply in cpu_dma_latency.
	forceLatencyIsSet bool        // forceLatencyLoop is true only if the continuous loop that maintains force-latency value is running.
	mutex             *sync.Mutex // mutex protects internal states from being modified concurrently.
}

// NewFunctionHost returns an initialised function host.
func NewFunctionHost() *FunctionHost {
	return &FunctionHost{
		forceLatency: make(chan int),
		mutex:        new(sync.Mutex),
	}
}

// RPCObjNameFmt is a format string that helps client to construct RPC call function name.
var RPCObjNameFmt = reflect.TypeOf(FunctionHost{}).Name() + ".%s"

type DummyAttr bool // DummyAttr is a placeholder receiver's type in an RPC function

/*
maintainDMALatency opens a file handle to /dev/cpu_dma_latency and feeds it the latest desired value.
Blocks caller until channel is closed.
*/
func (host *FunctionHost) maintainDMALatency() {
	// The file handle must be maintained open in order to set the value
	latency := make([]byte, 4)
	dmaLatency, err := os.OpenFile("/dev/cpu_dma_latency", os.O_RDWR, 0600)
	if err != nil {
		log.Printf("FunctionHost.maintainDMALatency: failed to open cpu_dma_latency - %v", err)
	}
	// Close the file handle after the latency value is no longer maintained
	defer dmaLatency.Close()

	for {
		newValue, keepGoing := <-host.forceLatency
		binary.LittleEndian.PutUint32(latency, uint32(newValue))
		if keepGoing {
			log.Printf("FunctionHost.maintainDMALatency: setting cpu_dma_latency to new value %d", newValue)
			// write new value into file
			_, err = dmaLatency.Write(latency)
			if err != nil {
				log.Printf("FunctionHost.maintainDMALatency: writing to '/dev/cpu_dma_latency' failed - %v", err)
			}
		} else {
			// Caller is responsible for maintaining forceLatencyIsSet flag
			log.Print("FunctionHost.maintainDMALatency: stop maintaining cpu_dma_latency")
			return
		}
	}
}

/*
SetForceLatency starts a loop that manipulates cpu_dma_latency to match input value. If the loop is already
started, the loop will be informed about the new value via a channel.
*/
func (host *FunctionHost) SetForceLatency(newValue int, _ *DummyAttr) error {
	host.mutex.Lock()
	defer host.mutex.Unlock()
	if !host.forceLatencyIsSet {
		host.forceLatencyIsSet = true
		go host.maintainDMALatency()
	}
	host.forceLatency <- newValue
	/*
	 The RPC function does not wait till value is set before responding to client.
	 Should an error occur, the goroutine in background will log the error and quit.
	*/
	return nil
}

// StopForceLatency stops the background loop that maintains cpu_dma_latency by closing its channel.
func (host *FunctionHost) StopForceLatency(_ DummyAttr, _ *DummyAttr) error {
	host.mutex.Lock()
	defer host.mutex.Unlock()
	if host.forceLatencyIsSet {
		close(host.forceLatency)
		host.forceLatencyIsSet = false
	}
	return nil
}

func GetForceLatency() string {
	latency := make([]byte, 4)
	dmaLatency, err := os.OpenFile("/dev/cpu_dma_latency", os.O_RDONLY, 0600)
	if err != nil {
		log.Printf("GetForceLatency: failed to open cpu_dma_latency - %v", err)
	}
	_, err = dmaLatency.Read(latency)
	if err != nil {
		log.Printf("GetForceLatency: reading from '/dev/cpu_dma_latency' failed:", err)
	}
	// Close the file handle after the latency value is no longer maintained
	defer dmaLatency.Close()

	ret := fmt.Sprintf("%v", binary.LittleEndian.Uint32(latency))
	return ret
}

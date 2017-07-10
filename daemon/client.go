package daemon

import (
	"fmt"
	"net"
	"net/rpc"
)

// Client connects to RPC server via domain socket in order to call RPC functions.
type Client struct {
}

// doRPC establishes a new connection via domain socket and executes exactly one RPC call.
func (client *Client) doRPC(fun func(*rpc.Client) error) error {
	conn, err := net.Dial("unix", DomainSocketFile)
	if err != nil {
		return err
	}
	defer conn.Close()
	rpcClient := rpc.NewClient(conn)
	defer rpcClient.Close()
	if err := fun(rpcClient); err != nil {
		return err
	}
	return nil
}

// SetForceLatency instructs RPC server to maintain a new value for cpu_dma_latency.
func (client *Client) SetForceLatency(newValue int) error {
	return client.doRPC(func(rpcClient *rpc.Client) error {
		var dummy DummyAttr
		return rpcClient.Call(fmt.Sprintf(RPCObjNameFmt, "SetForceLatency"), newValue, &dummy)
	})
}

// StopForceLatency instructs RPC server to stop background loop that maintains cpu_dma_latency value.
func (client *Client) StopForceLatency() error {
	return client.doRPC(func(rpcClient *rpc.Client) error {
		var dummy DummyAttr
		return rpcClient.Call(fmt.Sprintf(RPCObjNameFmt, "StopForceLatency"), dummy, &dummy)
	})
}

package daemon

import (
	"log"
	"net"
	"net/rpc"
	"os"
)

const (
	// DomainSocketFile is the file name of unix domain socket server that saptune daemon listens on.
	DomainSocketFile = "/var/run/saptune"
)

// Server is run by saptune system service to handle certain tuning parameters that require long-term maintenance.
type Server struct {
	listener  net.Listener  // listener is the unix domain socket listener.
	rpcServer *rpc.Server   // rpcServer is the RPC server serving connections on domain socket.
	host      *FunctionHost // host is an object of all RPC functions served to RPC client.
}

// Listen establishes unix domain socket listener and starts RPC server.
func (srv *Server) Listen() (err error) {
	if err := os.RemoveAll(DomainSocketFile); err != nil {
		return err
	}
	srv.listener, err = net.Listen("unix", DomainSocketFile)
	if err != nil {
		return
	}
	srv.host = NewFunctionHost()
	srv.rpcServer = rpc.NewServer()
	srv.rpcServer.Register(srv.host)
	log.Printf("Server.Listen: listening on %s", DomainSocketFile)
	return
}

// Shutdown closes server listener so that main loop (if running) will terminate.
func (srv *Server) Shutdown() {
	if listener := srv.listener; listener != nil {
		listener.Close()
		srv.listener = nil
	}
}

// MainLoop accepts and handles incoming connections in a continuous loop. Blocks caller until listener closes.
func (srv *Server) MainLoop() {
	for {
		client, err := srv.listener.Accept()
		if err != nil {
			log.Printf("Server.MainLoop: quit now - %v", err)
			return
		}
		go srv.rpcServer.ServeConn(client)
	}
}

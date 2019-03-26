package note

import (
	"github.com/SUSE/saptune_v1/system"
	"testing"
)

func TestSUSESysOptimisation(t *testing.T) {
	if !system.IsUserRoot() {
		t.Skip("the test requires root access")
	}
	sysop := SUSESysOptimisation{SysconfigPrefix: OSPackageInGOPATH}
	if sysop.Name() == "" {
		t.Fatal(sysop.Name())
	}
	initSysop, err := sysop.Initialise()
	if err != nil {
		t.Fatal(err)
	}
	optimised, err := initSysop.(SUSESysOptimisation).Optimise()
	if err != nil {
		t.Fatal(err)
	}
	// As written in file saptune-note-SUSE-GUIDE-01, all of the parameters are tuned by default.
	o := optimised.(SUSESysOptimisation)
	if o.VMNumberHugePages < 128 || o.VMSwappiness > 25 || o.VMVfsCachePressure > 50 || o.VMOvercommitMemory != 1 ||
		o.VMOvercommitRatio < 70 || o.VMDirtyRatio > 10 || o.VMDirtyBackgroundRatio > 5 {
		t.Fatalf("%+v", o)
	}
	// All elevators are set to noop
	for name, elevator := range optimised.(SUSESysOptimisation).BlockDeviceSchedulers.SchedulerChoice {
		if name == "" || elevator != "noop" {
			t.Fatalf("%+v", o)
		}
	}
}

func TestSUSENetCPUOptimisation(t *testing.T) {
	if !system.IsUserRoot() {
		t.Skip("the test requires root access")
	}
	netop := SUSENetCPUOptimisation{SysconfigPrefix: OSPackageInGOPATH}
	if netop.Name() == "" {
		t.Fatal(netop.Name())
	}
	initNetop, err := netop.Initialise()
	if err != nil {
		t.Fatal(err)
	}
	optimised, err := initNetop.(SUSENetCPUOptimisation).Optimise()
	if err != nil {
		t.Fatal(err)
	}
	// As written in file saptune-note-SUSE-GUIDE-02, all of the parameters are tuned by default.
	o := optimised.(SUSENetCPUOptimisation)
	if o.NetCoreWmemMax < 12582912 || o.NetCoreRmemMax < 12582912 ||
		o.NetCoreNetdevMaxBacklog < 9000 || o.NetCoreSoMaxConn < 512 ||
		o.NetIpv4TcpRmem < 9437184 || o.NetIpv4TcpWmem < 9437184 ||
		o.NetIpv4TcpTimestamps != 0 || o.NetIpv4TcpSack != 0 || o.NetIpv4TcpDsack != 0 || o.NetIpv4TcpFack != 0 ||
		o.NetIpv4IpfragHighThres < 544288 || o.NetIpv4IpfragLowThres < 393216 ||
		o.NetIpv4TcpMaxSynBacklog < 8192 || o.NetIpv4TcpSynackRetries > 3 || o.NetIpv4TcpRetries2 > 6 ||
		o.NetTcpKeepaliveTime > 1000 || o.NetTcpKeepaliveProbes > 4 || o.NetTcpKeepaliveIntvl > 20 ||
		o.NetTcpTwRecycle != 1 || o.NetTcpTwReuse != 1 ||
		o.NetTcpFinTimeout > 30 || o.NetTcpMtuProbing != 1 ||
		o.NetIpv4TcpSyncookies != 1 || o.NetIpv4ConfAllAcceptSourceRoute != 0 || o.NetIpv4ConfAllAcceptRedirects != 0 || o.NetIpv4ConfAllRPFilter != 1 ||
		o.NetIpv4IcmpEchoIgnoreBroadcasts != 1 || o.NetIpv4IcmpIgnoreBogusErrorResponses != 1 || o.NetIpv4ConfAllLogMartians != 1 || o.KernelRandomizeVASpace != 2 ||
		o.KernelKptrRestrict != 1 || o.FSProtectedHardlinks != 1 || o.FSProtectedSymlinks != 1 ||
		o.KernelSchedChildRunsFirst != 1 {
		t.Fatalf("%+v", o)
	}
}

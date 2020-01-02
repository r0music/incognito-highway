package common

import "time"

const (
	BEACONID      byte = 255
	NumberOfShard      = 8
	BeaconRole         = "beacon"
	ShardRole          = "shard"

	CommitteeRole = "committee"
	PendingRole   = "pending"
	WaitingRole   = "waiting"
	NormalNode    = ""

	MaxCallDepth                = 2
	MaxBlocksPerRequest         = 100
	MaxTimePerRequest           = 10 * time.Second
	ChoosePeerBlockDelta        = 100
	RouteClientKeepaliveTime    = 20 * time.Minute
	RouteClientKeepaliveTimeout = 20 * time.Second
	RouteHighwayKeepaliveTime   = 40 * time.Second
	BroadcastMsgEnlistTimestep  = 5 * time.Minute

	DefaultRPCServerPort = 9330
	DefaultHWPeerID      = "/ip4/127.0.0.1/tcp/9330/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"
)

const (
	COMMITTEE byte = iota
	PENDING
	NORMAL
)

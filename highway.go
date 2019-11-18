//+build !test
package main

import (
	"highway/chain"
	"highway/common"
	"highway/p2p"
	"highway/process"
	"highway/route"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	// Setup logging
	initLogger()

	config, err := GetProxyConfig()
	if err != nil {
		logger.Errorf("%+v", err)
		return
	}
	config.printConfig()
	masterPeerID, err := peer.IDB58Decode(config.masternode)
	if err != nil {
		logger.Error(err)
		return
	}

	chainData := new(process.ChainData)
	chainData.Init("keylist.json", common.NumberOfShard, common.CommitteeSize, masterPeerID)

	// New libp2p host
	proxyHost := p2p.NewHost(config.version, config.host, config.proxyPort, config.privateKey)

	// Pubsub
	if err := process.InitPubSub(proxyHost.Host, config.supportShards, chainData); err != nil {
		logger.Error(err)
		return
	}
	logger.Info("Init pubsub ok")
	go process.GlobalPubsub.WatchingChain()

	// Highway manager: connect cross highways
	rman := route.NewManager(
		config.supportShards,
		config.bootstrap,
		masterPeerID,
		proxyHost.Host,
		proxyHost.GRPC,
	)
	go rman.Start()

	// Chain-facing connections
	go chain.ManageChainConnections(proxyHost.Host, rman, proxyHost.GRPC, chainData, config.supportShards)

	// Subscribe to receive new committee
	process.GlobalPubsub.GRPCSpecSub <- process.SubHandler{
		Topic:   "chain_committee",
		Handler: chainData.ProcessChainCommitteeMsg,
	}

	time.Sleep(1 * time.Second)
	proxyHost.GRPC.Serve() // NOTE: must serve after registering all services
}

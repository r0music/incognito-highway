package process

import (
	context "context"
	"highway/common"
	logger "highway/customizelog"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
	grpc "google.golang.org/grpc"
)

func (hc *HighwayClient) GetBlockByHeight(
	shardID int32,
	specific bool,
	from uint64,
	to uint64,
	heights []uint64,
	fromCandidate string,
) ([][]byte, error) {
	client, err := hc.getClientWithPublicKey(fromCandidate)
	if err != nil {
		return nil, err
	}
	if shardID != -1 {
		reply, err := client.GetBlockShardByHeight(
			context.Background(),
			&GetBlockShardByHeightRequest{
				Shard:      shardID,
				Specific:   specific,
				FromHeight: from,
				ToHeight:   to,
				Heights:    heights,
				FromPool:   false,
			},
		)
		if err != nil {
			return nil, errors.WithStack(err)
		} else {
			logger.Infof("Reply: %v", reply)
			return reply.GetData(), nil
		}
	}

	reply, err := client.GetBlockBeaconByHeight(
		context.Background(),
		&GetBlockBeaconByHeightRequest{
			Specific:   specific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	} else {
		logger.Infof("Reply: %v", reply)
		return reply.GetData(), nil
	}
}

func (hc *HighwayClient) getClientWithPublicKey(
	committeePublicKey string,
) (HighwayServiceClient, error) {
	peerID, exist := hc.chainData.PeerIDByCommitteePubkey[committeePublicKey]
	if !exist {
		logger.Infof("Committee Publickey %v", committeePublicKey)
		PK := common.CommitteePublicKey{}
		PK.FromString(committeePublicKey)
		pkstring, _ := PK.MiningPublicKey()
		logger.Infof("Committee Publickey by mining key %v", common.MiningKeyByCommitteeKey[pkstring])
		return nil, errors.Errorf("Can not find PeerID of this committee PublicKey %v", committeePublicKey)
	}
	client, err := hc.cc.GetServiceClient(peerID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (hc *HighwayClient) GetBlockShardByHeight(
	shardID int32,
	specific bool,
	from uint64,
	to uint64,
	heights []uint64,
) ([][]byte, error) {
	client, err := hc.getClientWithBlock(int(shardID), specific, to, heights)
	if err != nil {
		return nil, err
	}
	reply, err := client.GetBlockShardByHeight(
		context.Background(),
		&GetBlockShardByHeightRequest{
			Shard:      shardID,
			Specific:   specific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
	)
	logger.Infof("Reply: %v", reply)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply.Data, nil
}

func (hc *HighwayClient) GetBlockShardToBeaconByHeight(
	shardID int32,
	specific bool,
	from uint64,
	to uint64,
	heights []uint64,
) ([][]byte, error) {
	client, err := hc.getClientWithBlock(int(shardID), specific, to, heights)
	if err != nil {
		return nil, err
	}
	reply, err := client.GetBlockShardToBeaconByHeight(
		context.Background(),
		&GetBlockShardToBeaconByHeightRequest{
			FromShard:  shardID,
			Specific:   specific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
	)
	logger.Infof("Reply: %v", reply)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply.Data, nil
}

func (hc *HighwayClient) GetBlockBeaconByHeight(
	specific bool,
	from uint64,
	to uint64,
	heights []uint64,
) ([][]byte, error) {
	client, err := hc.getClientWithBlock(int(common.BEACONID), specific, to, heights)
	if err != nil {
		return nil, err
	}
	reply, err := client.GetBlockBeaconByHeight(
		context.Background(),
		&GetBlockBeaconByHeightRequest{
			Specific:   specific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
	)
	logger.Infof("Reply: %v", reply)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply.Data, nil
}

func (hc *HighwayClient) getClientWithBlock(
	cid int,
	specific bool,
	to uint64,
	heights []uint64,
) (HighwayServiceClient, error) {
	peerID := peer.ID("")
	maxHeight := to
	if specific {
		maxHeight = heights[len(heights)-1]
	}
	peerID, err := hc.choosePeerIDWithBlock(cid, maxHeight)
	logger.Infof("Chosen peer: %v", peerID)
	if err != nil {
		return nil, err
	}

	client, err := hc.cc.GetServiceClient(peerID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (hc *HighwayClient) choosePeerIDWithBlock(cid int, blk uint64) (peer.ID, error) {
	// TODO(0xakk0r0kamui): choose client from peer state
	if len(hc.peers[int(cid)]) < 1 {
		return peer.ID(""), errors.Errorf("empty peer list for cid %v, block %v", cid, blk)
	}
	peerID, err := hc.chainData.GetPeerHasBlk(blk, byte(cid))
	if err != nil {
		return peer.ID(""), err
	}
	return *peerID, nil
	// return hc.peers[int(cid)][0], nil
}

type PeerInfo struct {
	ID  peer.ID
	CID int // CommitteeID
}

type HighwayClient struct {
	NewPeers  chan PeerInfo
	chainData *ChainData

	cc    *ClientConnector
	peers map[int][]peer.ID
}

func NewHighwayClient(pr *p2pgrpc.GRPCProtocol) *HighwayClient {
	hc := &HighwayClient{
		NewPeers: make(chan PeerInfo, 1000),
		cc:       NewClientConnector(pr),
		peers:    map[int][]peer.ID{},
	}
	go hc.start()
	return hc
}

func (hc *HighwayClient) start() {
	for {
		select {
		case p := <-hc.NewPeers:
			logger.Infof("Append new peer: cid = %v, pid = %v", p.CID, p.ID)
			hc.peers[p.CID] = append(hc.peers[p.CID], p.ID)
		}
	}
}

func (cc *ClientConnector) GetServiceClient(peerID peer.ID) (HighwayServiceClient, error) {
	// TODO(@0xbunyip): check if connection is alive or not; maybe return a list of conn for HighwayClient to retry if fail to connect
	if _, ok := cc.conns[peerID]; !ok { // TODO(@0xbunyip): lock access to cc.conns
		conn, err := cc.pr.Dial(
			context.Background(),
			peerID,
			grpc.WithInsecure(),
			grpc.WithBlock(),
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		cc.conns[peerID] = conn
	}
	client := NewHighwayServiceClient(cc.conns[peerID])
	return client, nil
}

type ClientConnector struct {
	pr    *p2pgrpc.GRPCProtocol
	conns map[peer.ID]*grpc.ClientConn
}

func NewClientConnector(pr *p2pgrpc.GRPCProtocol) *ClientConnector {
	return &ClientConnector{
		pr:    pr,
		conns: map[peer.ID]*grpc.ClientConn{},
	}
}

package rpcserver

import "github.com/libp2p/go-libp2p-core/peer"

type Handler struct {
	rpcServer *RpcServer
}

type PeerMap interface {
	CopyPeersMap() map[byte][]peer.AddrInfo
}

func (s *Handler) GetPeers(
	req Request,
	res *Response,
) (
	err error,
) {
	peers := s.rpcServer.pmap.CopyPeersMap()
	addrs := []string{}

	// NOTE: assume all highways support all shards => get at 0
	for _, p := range peers[0] {
		// TODO(@0xbunyip): get p[idx] with public ip
		ma, err := peer.AddrInfoToP2pAddrs(&p)
		if err != nil {
			logger.Warnf("Invalid addr info: %+v", p)
			continue
		}

		addrs = append(addrs, ma[0].String())
	}

	res.PeerPerShard = map[string][]string{
		"all": addrs,
	}
	return
}

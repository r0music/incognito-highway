package chaindata

import (
	"bytes"
	"highway/common"
	"highway/proto"
	"sort"
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain"
	ic "github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

type ChainData struct {
	ListMsgPeerStateOfShard    map[byte]CommitteeState //AllPeerState
	CurrentNetworkState        NetworkState
	CurrentNetworkStateV2      NetworkStateV2
	MiningPubkeyByPeerID       map[peer.ID]string
	PeerIDByMiningPubkey       map[string]peer.ID
	ShardByMiningPubkey        map[string]byte
	ShardPendingByMiningPubkey map[string]byte
	Locker                     *sync.RWMutex
}

type PeerWithBlk struct {
	HW     peer.ID
	ID     peer.ID
	Height uint64
}

func (chainData *ChainData) Init(numberOfShard int) error {
	logger.Info("Init chaindata")
	chainData.ListMsgPeerStateOfShard = map[byte]CommitteeState{}
	chainData.Locker = &sync.RWMutex{}
	for i := 0; i < numberOfShard; i++ {
		chainData.ListMsgPeerStateOfShard[byte(i)] = map[string][]byte{}
	}
	chainData.CurrentNetworkState.Init(numberOfShard)
	chainData.CurrentNetworkStateV2.Init(numberOfShard)
	chainData.MiningPubkeyByPeerID = map[peer.ID]string{}
	chainData.PeerIDByMiningPubkey = map[string]peer.ID{}
	chainData.ShardPendingByMiningPubkey = map[string]byte{}
	chainData.ShardByMiningPubkey = map[string]byte{}
	return nil
}

func (chainData *ChainData) GetCommitteeIDOfValidator(validator common.ProcessedKey) (byte, error) {
	miningPubkey := string(validator)

	chainData.Locker.RLock()
	defer chainData.Locker.RUnlock()
	if cid, ok := chainData.ShardByMiningPubkey[miningPubkey]; ok {
		return cid, nil
	}
	return 0, errors.New("candidate " + miningPubkey + " not found 2")
}

func (chainData *ChainData) GetPeerHasBlk(
	blkHeight uint64,
	committeeID byte,
) (
	[]PeerWithBlk,
	error,
) {
	var exist bool
	var committeeState map[string]ChainState
	chainData.Locker.RLock()
	defer chainData.Locker.RUnlock()
	if committeeID == common.BEACONID {
		committeeState = chainData.CurrentNetworkState.BeaconState
	} else {
		if committeeState, exist = chainData.CurrentNetworkState.ShardState[committeeID]; !exist {
			return nil, errors.New("committeeID " + string(committeeID) + " not found")
		}
	}
	peers := []PeerWithBlk{}
	for miningPubkey, nodeState := range committeeState {
		HWID, err := chainData.CurrentNetworkState.GetHWIDOfPubKey(miningPubkey)
		if err != nil {
			logger.Error(err)
			continue
		}
		peer := PeerWithBlk{
			HW:     HWID,
			ID:     peer.ID(""),
			Height: nodeState.Height,
		}

		// peerID is not mandatory, for peers connected to other highways, we
		// don't really care about their peerID
		if peerID, ok := chainData.PeerIDByMiningPubkey[miningPubkey]; ok {
			peer.ID = peerID
			// logger.Warnf("Committee publickey %v not found in PeerID map", miningPubkey)
		}
		peers = append(peers, peer)
	}

	// Sort based on block height
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].Height > peers[j].Height
	})
	return peers, nil
}

func (chainData *ChainData) GetPeerHasBlkV2(
	blkHeight uint64,
	committeeID byte,
) (
	[]PeerWithBlk,
	error,
) {
	var exist bool
	var committeeState map[string]ChainState
	chainData.Locker.RLock()
	defer chainData.Locker.RUnlock()
	if committeeID == common.BEACONID {
		committeeState = chainData.CurrentNetworkStateV2.BeaconState
	} else {
		if committeeState, exist = chainData.CurrentNetworkStateV2.ShardState[committeeID]; !exist {
			return nil, errors.New("committeeID " + string(committeeID) + " not found")
		}
	}
	peers := []PeerWithBlk{}
	for pID, nodeState := range committeeState {
		HWID, err := chainData.CurrentNetworkStateV2.GetHWIDOfPeerID(pID)
		if err != nil {
			logger.Error(err)
			continue
		}
		peerID, err := peer.IDB58Decode(pID)
		if err != nil {
			logger.Error(err)
			continue
		}
		peer := PeerWithBlk{
			HW:     HWID,
			ID:     peerID,
			Height: nodeState.Height,
		}
		peers = append(peers, peer)
	}

	// Sort based on block height
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].Height > peers[j].Height
	})
	return peers, nil
}

func (chainData *ChainData) GetHWIDOfPeer(
	pID peer.ID,
) (
	peer.ID,
	error,
) {
	chainData.Locker.RLock()
	miningPubkey, ok := chainData.MiningPubkeyByPeerID[pID]
	chainData.Locker.RUnlock()
	if !ok {
		return "", errors.Errorf("Can not found info of this peerID %v", pID.String())
	}

	HWID, err := chainData.CurrentNetworkState.GetHWIDOfPubKey(miningPubkey)
	if err != nil {
		return "", err
	}
	return HWID, nil
}

// UpdateCommittee saves peerID, mining pubkey and committeeID of a validator
func (chainData *ChainData) UpdateCommittee(pubkey common.ProcessedKey, peerID peer.ID, cid byte) {
	// Convert from CommitteePubkey to MiningPubKey if user submitted one
	miningPubkey := string(pubkey)

	// Map between mining pubkey and peerID
	chainData.Locker.Lock()
	defer chainData.Locker.Unlock()
	chainData.MiningPubkeyByPeerID[peerID] = miningPubkey
	chainData.PeerIDByMiningPubkey[miningPubkey] = peerID
	chainData.ShardByMiningPubkey[miningPubkey] = cid
}

func (chainData *ChainData) UpdateStateV2WithMsgPeerState(
	committeeID byte,
	peerID string,
	msgPeerState *wire.MessagePeerState,
) error {
	chainData.Locker.Lock()
	chainData.CurrentNetworkStateV2.BeaconState[peerID] = MergeChainState(newChainStateFromMsgPeerState(msgPeerState, common.BEACONID), chainData.CurrentNetworkStateV2.BeaconState[peerID])
	for i := 0; i < 8; i++ {
		chainData.CurrentNetworkStateV2.ShardState[byte(i)][peerID] = MergeChainState(newChainStateFromMsgPeerState(msgPeerState, byte(i)), chainData.CurrentNetworkStateV2.ShardState[byte(i)][peerID])
	}
	// // }
	// logger.Infof("[newpeerstate] NetworkState:")
	// logger.Infof("[newpeerstate] Beacon:")
	// for k, v := range chainData.CurrentNetworkStateV2.BeaconState {
	// 	logger.Infof("[newpeerstate] Peer %v: Height %v", k, v.Height)
	// }
	// for k, v := range chainData.CurrentNetworkStateV2.ShardState {
	// 	logger.Infof("[newpeerstate] Shard %v:", k)
	// 	for pID, state := range v {
	// 		logger.Infof("[newpeerstate] Peer %v: Height %v", pID, state.Height)
	// 	}
	// }
	// logger.Infof("[newpeerstate]-----------------------------------")
	defer chainData.Locker.Unlock()
	return nil
}

func (chainData *ChainData) UpdateStateWithMsgPeerState(
	committeeID byte,
	committeePublicKey string,
	msgPeerState *wire.MessagePeerState,
) error {
	chainData.Locker.Lock()
	if committeeID == common.BEACONID {
		chainData.CurrentNetworkState.BeaconState[committeePublicKey] = newChainStateFromMsgPeerState(msgPeerState, committeeID)
	} else {
		chainData.CurrentNetworkState.ShardState[committeeID][committeePublicKey] = newChainStateFromMsgPeerState(msgPeerState, committeeID)
	}
	defer chainData.Locker.Unlock()
	return nil
}

func (chainData *ChainData) UpdatePeerStateFromHW(publisher peer.ID, data []byte, committeeID byte) error {
	//TODO check Highway signature
	msgPeerState, err := common.ParsePeerStateData(string(data))
	if err != nil {
		logger.Errorf(err.Error())
		return err
	}
	peerPublicKey := msgPeerState.SenderMiningPublicKey
	pkey, err := common.PreprocessKey(peerPublicKey)
	if err != nil {
		return err
	}
	// Store peerID of HW connected to a peer
	miningPubkey := string(pkey)
	err = chainData.CurrentNetworkState.SetHWIDOfPubKey(publisher, miningPubkey)
	if err != nil {
		logger.Errorf(err.Error())
		return err
	}
	err = chainData.CurrentNetworkStateV2.SetHWIDOfPeerID(publisher, msgPeerState.SenderID)
	if err != nil {
		logger.Errorf(err.Error())
		return err
	}
	// Store committeeID, peerID and pubkey of a peer
	pid, err := peer.IDB58Decode(msgPeerState.SenderID)
	if err != nil {
		logger.Errorf("Received invalid peerID from msg peerstate: %v %s", err, msgPeerState.SenderID)
	} else {
		// logger.Debugf("Updating committee: pkey = %v pid = %s cid = %v", pkey, pid.String(), committeeID)
		chainData.UpdateCommittee(pkey, pid, committeeID)
	}
	// Save peerstate by miningPubkey
	chainData.Locker.Lock()
	if chainData.ListMsgPeerStateOfShard[committeeID] == nil {
		chainData.ListMsgPeerStateOfShard[committeeID] = map[string][]byte{}
	}
	if !bytes.Equal(chainData.ListMsgPeerStateOfShard[committeeID][miningPubkey], data) {
		chainData.ListMsgPeerStateOfShard[committeeID][miningPubkey] = data
	}
	chainData.Locker.Unlock()
	err = chainData.UpdateStateV2WithMsgPeerState(committeeID, msgPeerState.SenderID, msgPeerState)
	if err != nil {
		logger.Errorf(err.Error())
		return err
	}
	return chainData.UpdateStateWithMsgPeerState(
		committeeID,
		miningPubkey,
		msgPeerState,
	)
	return nil
}

func newChainStateFromMsgPeerState(
	msgPeerState *wire.MessagePeerState,
	committeeID byte,
	// candidateKey string,
) ChainState {
	var blkChainState blockchain.ChainState
	if committeeID == common.BEACONID {
		blkChainState = msgPeerState.Beacon
	} else {
		blkChainState = msgPeerState.Shards[committeeID]
	}
	return ChainState{
		Height:        blkChainState.Height,
		Timestamp:     blkChainState.Timestamp,
		BestStateHash: blkChainState.BestStateHash,
		BlockHash:     blkChainState.BlockHash,
	}
}

func (chainData *ChainData) CopyNetworkState() NetworkState {
	chainData.Locker.RLock()
	defer chainData.Locker.RUnlock()
	state := NetworkState{
		BeaconState: map[string]ChainState{},
		ShardState:  map[byte]map[string]ChainState{},
	}
	for key, cs := range chainData.CurrentNetworkState.BeaconState {
		state.BeaconState[key] = cs
	}
	for cid, states := range chainData.CurrentNetworkState.ShardState {
		state.ShardState[cid] = map[string]ChainState{}
		for key, cs := range states {
			state.ShardState[cid][key] = cs
		}
	}
	return state
}

func getKeyListFromMessage(comm *incognitokey.ChainCommittee) (*common.KeyList, error) {
	// TODO(@0xbunyip): handle epoch
	keys := &common.KeyList{Sh: map[int][]common.Key{}}
	for _, k := range comm.BeaconCommittee {
		cpk, err := k.ToBase58()
		if err != nil {
			return nil, errors.Wrapf(err, "key: %+v", k)
		}
		keys.Bc = append(keys.Bc, common.Key{CommitteePubKey: cpk})
	}

	for s, vals := range comm.AllShardCommittee {
		for _, val := range vals {
			cpk, err := val.ToBase58()
			if err != nil {
				return nil, errors.Wrapf(err, "key: %+v", val)
			}
			keys.Sh[int(s)] = append(keys.Sh[int(s)], common.Key{CommitteePubKey: cpk})
		}
	}

	// Shard's pending validators
	for s, pends := range comm.AllShardPending {
		for _, pend := range pends {
			cpk, err := pend.ToBase58()
			if err != nil {
				return nil, errors.Wrapf(err, "key: %+v", pend)
			}
			keys.ShPend[int(s)] = append(keys.ShPend[int(s)], common.Key{CommitteePubKey: cpk})
		}
	}

	return keys, nil
}

func GetUserRole(role string, cid int) *proto.UserRole {
	layer := ""
	if cid == int(common.BEACONID) {
		layer = ic.BeaconRole
	} else if cid != -1 { // other than NORMAL
		layer = ic.ShardRole
	} else {
		layer = ""
		role = ""
	}
	return &proto.UserRole{
		Layer: layer,
		Role:  role,
		Shard: int32(cid),
	}
}

func MergeChainState(src, dst ChainState) ChainState {
	if src.Height > dst.Height {
		return src
	}
	return dst
}

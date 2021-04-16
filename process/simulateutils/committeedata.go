package simulateutils

import (
	"highway/common"
	"sort"
	"sync"
)

type CommitteeTable struct {
	PubKeyBySID map[byte][]string
	SIDByPubKey map[string]byte
	PubKeyIdx   map[string]int
	lock        *sync.RWMutex

	CommitteeSize map[byte]int
}

func NewCommitteeTable() *CommitteeTable {
	return &CommitteeTable{
		PubKeyBySID: map[byte][]string{},
		SIDByPubKey: map[string]byte{},
		PubKeyIdx:   map[string]int{},
		lock:        &sync.RWMutex{},

		CommitteeSize: map[byte]int{},
	}
}

func (table *CommitteeTable) AddPubKey(pubKey string, SID byte, idx int) {
	table.lock.Lock()
	defer table.lock.Unlock()
	if sID, ok := table.SIDByPubKey[pubKey]; ok {
		if sID == SID {
			return
		}
		delete(table.SIDByPubKey, pubKey)
		delete(table.PubKeyIdx, pubKey)
		table.PubKeyBySID[sID] = common.DeleteStringInList(pubKey, table.PubKeyBySID[sID])
	}
	table.SIDByPubKey[pubKey] = SID
	table.PubKeyBySID[SID] = append(table.PubKeyBySID[SID], pubKey)
	table.PubKeyIdx[pubKey] = idx
	size, ok := table.CommitteeSize[SID]
	if ok {
		if idx+1 > size {
			size = idx + 1
		}
	}
	table.CommitteeSize[SID] = size
	sort.Slice(table.PubKeyBySID[SID], func(i, j int) bool {
		return table.PubKeyIdx[table.PubKeyBySID[SID][i]] < table.PubKeyIdx[table.PubKeyBySID[SID][j]]
	})
}

func (table *CommitteeTable) RemovePubKey(pubKey string, SID byte, idx int) {
	table.lock.Lock()
	defer table.lock.Unlock()
	if sID, ok := table.SIDByPubKey[pubKey]; ok {
		delete(table.SIDByPubKey, pubKey)
		delete(table.PubKeyIdx, pubKey)
		table.PubKeyBySID[sID] = common.DeleteStringInList(pubKey, table.PubKeyBySID[sID])
	}
}

func (table *CommitteeTable) GetKeysByKey(pubKey string) []string {
	table.lock.RLock()
	defer table.lock.RUnlock()
	if sID, ok := table.SIDByPubKey[pubKey]; ok {
		return table.PubKeyBySID[sID]
	}
	return []string{}
}

func (table *CommitteeTable) GetSize(cID byte) int {
	table.lock.RLock()
	defer table.lock.RUnlock()
	if size, ok := table.CommitteeSize[cID]; ok {
		return size
	}
	return 0
}

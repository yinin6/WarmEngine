package pbft_all

import (
	"blockEmulator/consensus_shard/reconfiguration"
	"encoding/json"
	"log"
	"sync"
)

var (
	ShuffleReadyLock = sync.Mutex{}
	SyncOverLock     = sync.Mutex{}
)

func (p *PbftConsensusNode) handleShuffleReady(content []byte) {
	ShuffleReadyLock.Lock()
	defer ShuffleReadyLock.Unlock()
	shuffleReadyMessage := new(reconfiguration.ShuffleReadyMessage)
	err := json.Unmarshal(content, shuffleReadyMessage)
	if err != nil {
		log.Panic(err)
	}
	t, ok := p.Epoch2ReadyShard.Load(shuffleReadyMessage.Epoch)
	if !ok {
		p.Epoch2ReadyShard.Store(shuffleReadyMessage.Epoch, 1)
	} else {
		p.Epoch2ReadyShard.Store(shuffleReadyMessage.Epoch, t.(int)+1)
	}
	p.pl.Plog.Printf("S%dN%d : receive the shuffle ready message from shard %d for epoch %d\n", p.ShardID, p.NodeID, shuffleReadyMessage.ShardID, shuffleReadyMessage.Epoch)
}

func (p *PbftConsensusNode) handleFullSyncOver(content []byte) {
	SyncOverLock.Lock()
	defer SyncOverLock.Unlock()
	shuffleOver := new(reconfiguration.NotifyLeaderSyncOver)
	err := json.Unmarshal(content, shuffleOver)
	if err != nil {
		log.Panic(err)
	}
	p.pl.Plog.Printf("S%dN%d : receive sync over message for shard %v,Epoch %v NodeID %v from address  %v\n",
		p.ShardID, p.NodeID, shuffleOver.ShardID, shuffleOver.Epoch, shuffleOver.NodeID, shuffleOver.FromAddress)
	p.pl.Plog.Printf("new account is clear len %v \n", p.CurChain.GetAccountPoolSize())
	t, ok := p.OverSyncInShard.Load(shuffleOver.Epoch)
	if !ok {
		p.OverSyncInShard.Store(shuffleOver.Epoch, 1)
	} else {
		p.OverSyncInShard.Store(shuffleOver.Epoch, t.(int)+1)
	}

}

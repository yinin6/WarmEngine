package reconfiguration

import (
	"blockEmulator/chain"
	"blockEmulator/consensus_shard/reconfiguration/util"
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"
)

var (
	beginTime     = make(map[uint64]time.Time)
	receiveTime   = make(map[uint64]time.Time)
	overTime      = make(map[uint64]time.Time)
	stateDataSize = make(map[uint64]int)
	blockSize     = make(map[uint64]int)
)

var (
	UpdateHash         = make(map[string]int)
	RepeatedEpochRound = make(map[uint64]map[int]int)
)

var (
	waitForLossAccount = make(chan bool) // used to wait for the loss account
)

func (r *Reconfiguration) FullSync(destShard uint64) {
	s := &FullSync{
		FromAddress: r.Address,
		Epoch:       r.Epoch,
	}
	beginTime[r.Epoch] = time.Now()
	r.logger.Printf("now ip table is %v \n", r.ipNodeTable)
	r.logger.Printf("%v is selected FullSync, begin send request to %v \n", s, r.ipNodeTable[destShard][0])
	encode, err := networks.Encode(s)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.FullSync error: %v \n", err)
	}
	msgSend := message.MergeMessage(CFullSync, encode)
	target := r.tmptGetTargetAddress(destShard)
	networks.TcpDial(msgSend, target)
}

func (r *Reconfiguration) handleFullSyncData(content []byte) {
	syncData := &SyncData{}
	err := json.Unmarshal(content, syncData)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.handleFullSyncData error: %v \n", err)
	}
	r.logger.Printf("now Epoch is %v, Receive full sync data , wait to add database. Epoch %v \n", r.Epoch, syncData.Epoch)
	r.SyncDataUpdateBlockChain(r.CurChain, syncData)
}

func (r *Reconfiguration) SyncDataUpdateBlockChain(chain *chain.BlockChain, syncData *SyncData) {
	// record data
	receiveTime[syncData.Epoch] = time.Now()
	stateSize := 0
	for i := 0; i < len(syncData.Keys); i++ {
		stateSize += len(syncData.Keys[i]) + len(syncData.Values[i])
	}
	stateDataSize[syncData.Epoch] = stateSize
	blockSize[syncData.Epoch] = len(syncData.CurrBlock)

	block := core.DecodeB(syncData.CurrBlock)
	r.logger.Printf("receive block with hash %x \n", block.Header.StateRoot)
	chain.Storage.AddBlock(block)
	chain.CurrentBlock = block
	chain.ChainConfig.ShardID = r.RunningNode.ShardID
	for i := 0; i < len(syncData.Keys); i++ {
		err := r.db.Put(syncData.Keys[i], syncData.Values[i])
		if err != nil {
			r.logger.Fatalf("Reconfiguration.SyncDataUpdateBlockChain error: %v \n", err)
		}
	}
	r.logger.Printf("sync data over, len %v blob \n", len(syncData.Keys))
	overTime[syncData.Epoch] = time.Now()
	r.NotifyLeaderSyncOver()
	r.RecordData()
}

func (r *Reconfiguration) NotifyLeaderSyncOver() {
	s := &NotifyLeaderSyncOver{
		ShardID:     r.RunningNode.ShardID,
		FromAddress: r.Address,
		NodeID:      r.RunningNode.NodeID,
		Epoch:       r.Epoch,
	}
	encode, err := json.Marshal(s)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.NotifyLeaderSyncOver error: %v \n", err)
	}
	msgSend := message.MergeMessage(CFullSyncOver, encode)
	networks.TcpDial(msgSend, r.ipNodeTableOfPBFT[r.RunningNode.ShardID][0])
	r.logger.Printf("send sync confirmed message to leader %v, epoch %v \n", r.ipNodeTableOfPBFT[r.RunningNode.ShardID][0], s.Epoch)
}

func (r *Reconfiguration) RecordData() {
	name1 := "S" + strconv.Itoa(int(r.RawShardID)) + "N" + strconv.Itoa(int(r.RawNodeID))
	file, err := os.OpenFile(util.GetFilePath(name1), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	if _, err := file.WriteString(
		strconv.FormatUint(r.Epoch, 10) +
			"," + strconv.FormatUint(r.RunningNode.ShardID, 10) +
			"," + strconv.FormatUint(r.RunningNode.NodeID, 10) +
			"," + strconv.FormatUint(uint64(beginTime[r.Epoch].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(receiveTime[r.Epoch].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(overTime[r.Epoch].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(overTime[r.Epoch].UnixMilli()-beginTime[r.Epoch].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(stateDataSize[r.Epoch]), 10) +
			"," + strconv.FormatUint(uint64(blockSize[r.Epoch]), 10) +
			"\n"); err != nil {
		log.Panic(err)
	}
}

package reconfiguration

import (
	"blockEmulator/chain"
	"blockEmulator/consensus_shard/reconfiguration/util"
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

// the implementation of the tMPT
func (r *Reconfiguration) TMptSync(destShard uint64) {
	s := &TMptSync{
		FromAddress: r.Address,
		Epoch:       r.Epoch,
	}
	beginTime[r.Epoch] = time.Now()
	target := r.tmptGetTargetAddress(destShard)
	r.logger.Printf("TMP step 1, %v selected TMptSync, begin send request to %v.  \n", s, target)
	marshal, err := json.Marshal(s)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.FullSync error: %v \n", err)
	}
	msgSend := message.MergeMessage(CTmptSync, marshal)
	networks.TcpDial(msgSend, target)
}

func (r *Reconfiguration) tmptGetTargetAddress(shardID uint64) string {
	var ans string
	migrateNodeNumber := params.MigrateNodeNum
	t := int(int(r.RunningNode.NodeID-1)/params.ZoneSize) + migrateNodeNumber + 1
	ans = r.ipNodeTable[shardID][uint64(t)]
	r.logger.Printf("NodeID %v, targetID %v getSyncTargets: %v \n", r.RunningNode.NodeID, t, ans)
	return ans
}

func (r *Reconfiguration) handleTMptSyncData(content []byte) {
	tMptSyncData := &TMptSyncData{}
	err := json.Unmarshal(content, tMptSyncData)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.handleTMptSyncData error: %v \n", err)
	}
	r.logger.Printf("TMP step 2, Epoch is %v, Receive tmpt sync data. receive Epoch %v, active account %v \n", r.Epoch, tMptSyncData.Epoch, len(tMptSyncData.Accounts))
	r.SyncTMptUpdateBlockChain(r.CurChain, tMptSyncData)
}

func (r *Reconfiguration) SyncTMptUpdateBlockChain(chain *chain.BlockChain, syncData *TMptSyncData) {
	// record data
	receiveTime[syncData.Epoch] = time.Now()
	stateSize := 0
	for i := 0; i < len(syncData.Keys); i++ {
		stateSize += len(syncData.Keys[i]) + len(syncData.Values[i])
	}
	stateDataSize[syncData.Epoch] = stateSize
	blockSize[syncData.Epoch] = len(syncData.CurrBlock)

	block := core.DecodeB(syncData.CurrBlock)
	//update account
	r.CurChain.TmptAppendAccounts(syncData.Epoch, syncData.Accounts)

	r.logger.Printf("TMP step 2, receive block with hash %x, new account %v, total state data size %v \n", block.Header.StateRoot, len(syncData.Accounts), stateSize)
	chain.Storage.AddBlock(block)
	chain.CurrentBlock = block
	chain.ChainConfig.ShardID = r.RunningNode.ShardID
	for i := 0; i < len(syncData.Keys); i++ {
		err := r.db.Put(syncData.Keys[i], syncData.Values[i])
		if err != nil {
			r.logger.Fatalf("Reconfiguration error: %v \n", err)
		}
	}
	r.logger.Printf("sync data over, len %v blob \n", len(syncData.Keys))
	overTime[syncData.Epoch] = time.Now()
	r.NotifyLeaderSyncOver()
	r.RecordData()
}

// SpecificSync the implementation of the specific sync, used to sync the loss account
func (r *Reconfiguration) TMPTSpecificSync(accounts []string) {
	if _, ok := accountNumberEpochRound[r.Epoch]; !ok {
		accountNumberEpochRound[r.Epoch] = make(map[int]int)
	}
	accountNumberEpochRound[r.Epoch][r.PreSyncRound] = len(accounts)
	// record for tmpt use
	if params.SyncMod == 1 {
		if _, ok := beginTimeEpochRound[r.Epoch]; !ok {
			beginTimeEpochRound[r.Epoch] = make(map[int]time.Time)
			UpdateHash = make(map[string]int)
		}
		beginTimeEpochRound[r.Epoch][r.PreSyncRound] = time.Now()
	}
	if len(accounts) == 0 {
		r.logger.Printf("TMPTSpecificSync: no account need to sync \n")
		return
	}
	r.logger.Printf("TMP step 3, account %v need to sync \n", len(accounts))
	s := &SpecificSync{
		FromAddress: r.Address,
		Epoch:       r.Epoch,
		RootHash:    r.CurChain.CurrentBlock.Header.StateRoot,
		BlockHeight: int(r.CurChain.CurrentBlock.Header.Number),
		Accounts:    accounts,
	}
	target := r.tmptGetTargetAddress(r.RunningNode.ShardID)
	r.logger.Printf("TMP step 3, begin send request to %v. \n", target)
	marshal, err := json.Marshal(s)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.SpecificSync error: %v \n", err)
	}
	msgSend := message.MergeMessage(CTMPTSpecificSync, marshal)
	networks.TcpDial(msgSend, target)
	r.logger.Printf("TMPTSpecificSync send specific to leadr %v, wait for receive loss acount \n", target)

	// wait for the loss account
	<-waitForLossAccount
	r.logger.Printf("receive loss account, begin consensus \n")
	r.CurChain.TmptAppendAccounts(r.Epoch, accounts)
	r.logger.Printf("tmpt: record data, now Round %v\n", r.PreSyncRound)
	r.RecordEpochRound()
	r.PreSyncRound++
}

func (r *Reconfiguration) handleTMPTSpecificSyncData(content []byte) {
	syncData := &SyncSpecificData{}
	err := json.Unmarshal(content, syncData)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.handleSpecificSyncData error: %v \n", err)
	}
	r.logger.Printf("now Epoch is %v, Receive specific sync data , wait to add database. Epoch %v \n", r.Epoch, syncData.Epoch)
	datasize := 0
	keySize := 0
	valSize := 0
	repeat := 0
	for i := 0; i < len(syncData.Keys); i++ {
		keySize += len(syncData.Keys[i])
		valSize += len(syncData.Values[i])
		datasize += len(syncData.Keys[i]) + len(syncData.Values[i])
		if UpdateHash[string(syncData.Keys[i])] == 1 {
			repeat++
		} else {
			UpdateHash[string(syncData.Keys[i])] = 1
		}
		err := r.db.Put(syncData.Keys[i], syncData.Values[i])
		if err != nil {
			r.logger.Fatalf("Reconfiguration error: %v \n", err)
		}
	}
	r.logger.Printf("sync data over, len %v blob,total blob %v \n", len(syncData.Keys), len(UpdateHash))

	// record the time
	if _, ok := overTimeEpochRound[r.Epoch]; !ok {
		overTimeEpochRound[r.Epoch] = make(map[int]time.Time)
	}
	overTimeEpochRound[r.Epoch][r.PreSyncRound] = time.Now()

	if _, ok := stateDataSizeEpochRound[r.Epoch]; !ok {
		stateDataSizeEpochRound[r.Epoch] = make(map[int]int)
	}
	if _, ok := stateKeySizeEpochRound[r.Epoch]; !ok {
		stateKeySizeEpochRound[r.Epoch] = make(map[int]int)
	}
	if _, ok := stateValueSizeEpochRound[r.Epoch]; !ok {
		stateValueSizeEpochRound[r.Epoch] = make(map[int]int)
	}

	if _, ok := RepeatedEpochRound[r.Epoch]; !ok {
		RepeatedEpochRound[r.Epoch] = make(map[int]int)
	}
	stateDataSizeEpochRound[r.Epoch][r.PreSyncRound] = datasize
	stateKeySizeEpochRound[r.Epoch][r.PreSyncRound] = keySize
	stateValueSizeEpochRound[r.Epoch][r.PreSyncRound] = valSize
	RepeatedEpochRound[r.Epoch][r.PreSyncRound] = repeat

	r.logger.Printf("sync data over, ready to consensus \n")
	waitForLossAccount <- true
}

var (
	tmptRecordLock = sync.Mutex{}
)

func (r *Reconfiguration) RecordEpochRound() {
	tmptRecordLock.Lock()
	defer tmptRecordLock.Unlock()
	name2 := "S" + strconv.Itoa(int(r.RawShardID)) + "N" + strconv.Itoa(int(r.RawNodeID)) + "specific"
	file, err := os.OpenFile(util.GetFilePath(name2), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Panic(err)
		}
	}(file)

	if _, err := file.WriteString(
		strconv.FormatUint(r.Epoch, 10) +
			"," + strconv.FormatUint(uint64(r.PreSyncRound), 10) +
			"," + strconv.FormatUint(r.RunningNode.ShardID, 10) +
			"," + strconv.FormatUint(r.RunningNode.NodeID, 10) +
			"," + strconv.FormatUint(uint64(beginTimeEpochRound[r.Epoch][r.PreSyncRound].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(receiveTimeEpochRound[r.Epoch][r.PreSyncRound].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(overTimeEpochRound[r.Epoch][r.PreSyncRound].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(overTimeEpochRound[r.Epoch][r.PreSyncRound].UnixMilli()-beginTimeEpochRound[r.Epoch][r.PreSyncRound].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(stateDataSizeEpochRound[r.Epoch][r.PreSyncRound]), 10) +
			"," + strconv.FormatUint(uint64(blockSizeEpochRound[r.Epoch][r.PreSyncRound]), 10) +
			"," + strconv.FormatUint(uint64(RepeatedEpochRound[r.Epoch][r.PreSyncRound]), 10) +
			"," + strconv.FormatUint(uint64(stateKeySizeEpochRound[r.Epoch][r.PreSyncRound]), 10) +
			"," + strconv.FormatUint(uint64(stateValueSizeEpochRound[r.Epoch][r.PreSyncRound]), 10) +
			"," + strconv.FormatUint(uint64(accountNumberEpochRound[r.Epoch][r.PreSyncRound]), 10) +
			"\n"); err != nil {
		log.Panic(err)
	}
}

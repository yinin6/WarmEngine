package reconfiguration

import (
	"blockEmulator/chain"
	"blockEmulator/consensus_shard/reconfiguration/util"
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	CPreSync     message.MessageType = "CPreSync"
	CPreSyncData message.MessageType = "CPreSyncData"

	// specific sync
	CSpecificSync     message.MessageType = "CSpecificSync"
	CSpecificSyncData message.MessageType = "CSpecificSyncData"
	CErasureChunk     message.MessageType = "CErasureChunk"
)

var (
	// for monitor
	beginTimeEpochRound      = make(map[uint64]map[int]time.Time)
	receiveTimeEpochRound    = make(map[uint64]map[int]time.Time)
	overTimeEpochRound       = make(map[uint64]map[int]time.Time)
	blockSizeEpochRound      = make(map[uint64]map[int]int)
	stateDataSizeEpochRound  = make(map[uint64]map[int]int)
	stateKeySizeEpochRound   = make(map[uint64]map[int]int)
	stateValueSizeEpochRound = make(map[uint64]map[int]int)
	accountNumberEpochRound  = make(map[uint64]map[int]int)
)

var (
	ProductorReceiveTime = make(map[uint64]map[int]time.Time)
	ProductorSendTime    = make(map[uint64]map[int]time.Time)
	LocalReceiveTime     = make(map[uint64]map[int]time.Time)
	RecordLock           = sync.Mutex{}
)

var (
	dataLock      = sync.Mutex{}
	fetchDataLock = sync.Mutex{}

	erasureCodePool = make(map[ChunkIndex][]ErasureChunk, 0)
	erasueCodeLock  = sync.Mutex{}
	roundCheckLock  = sync.NewCond(&sync.Mutex{})

	chunkIndexPool = make(map[ChunkIndex]struct{}, 0)
	chunkIndexCond = sync.NewCond(&sync.Mutex{})
)

type ChunkIndex struct {
	IdxChunk int
	Epoch    int
	Round    int
}

type PreSync struct {
	FromAddress string
	Epoch       uint64
	Cnt         int
}

type PreSyncData struct {
	FromShard uint64
	Address   string
	Epoch     uint64
	Round     int
	CurrBlock []byte
	Accounts  []string
}

func (r *Reconfiguration) PreSync() {
	r.logger.SetPrefix(fmt.Sprintf("Epoch %v, Round %v, ", r.Epoch, r.PreSyncRound))
	if (r.PreSyncRound) != 0 && (r.PreSyncRound)%(params.Frequency-1) == 0 {
		r.logger.Printf("Step 0: Epch %v, Round %v, next Round begin shuffe, need not sync.\n", r.Epoch, r.PreSyncRound)
		return
	}
	s := &PreSync{
		FromAddress: r.Address,
		Epoch:       r.Epoch,
		Cnt:         r.PreSyncRound,
	}
	if _, ok := beginTimeEpochRound[r.Epoch]; !ok {
		beginTimeEpochRound[r.Epoch] = make(map[int]time.Time)
		UpdateHash = make(map[string]int)
	}
	beginTimeEpochRound[r.Epoch][r.PreSyncRound] = time.Now()
	targetAddr := r.getSyncTargets(params.NChunk, r.RunningNode.ShardID)

	r.logger.Printf("Step 0: %v PreSync, Round %v begin send request to %v.  \n", s, r.PreSyncRound, targetAddr[0])

	marshal, err := json.Marshal(s)
	if err != nil {
		r.logger.Fatalf("PreSync error: %v \n", err)
	}
	msgSend := message.MergeMessage(CPreSync, marshal)
	go networks.TcpDial(msgSend, targetAddr[0])
	r.PreSyncRound++

}

// handlePreSyncData 获取最新的区块，以及需要同步的账户状态
func (r *Reconfiguration) handlePreSyncData(content []byte) {
	r.logger.SetPrefix(fmt.Sprintf("Epoch %v, Round %v, ", r.Epoch, r.PreSyncRound))
	preSyncData := &PreSyncData{}
	err := json.Unmarshal(content, preSyncData)
	if err != nil {
		r.logger.Fatalf("handlePreSyncData error: %v \n", err)
	}
	r.logger.Printf("Step 1.1: Receive pre sync data list, wait to add database. Epoch %v, Round %v \n", preSyncData.Epoch, preSyncData.Round)
	blocksize := len(preSyncData.CurrBlock)
	if preSyncData.Round == 0 {
		block := core.DecodeB(preSyncData.CurrBlock)
		r.logger.Printf("Step 1.2: Epoch %v Round %v, begin add block to chain, new hash %x \n", preSyncData.Epoch, preSyncData.Round, block.Header.StateRoot)
		r.CurChain.Storage.AddBlock(block)
		r.CurChain.CurrentBlock = block
		r.CurChain.ChainConfig.ShardID = r.RunningNode.ShardID
	}

	if _, ok := blockSizeEpochRound[preSyncData.Epoch]; !ok {
		blockSizeEpochRound[preSyncData.Epoch] = make(map[int]int)
	}
	blockSizeEpochRound[preSyncData.Epoch][preSyncData.Round] = blocksize

	if _, ok := receiveTimeEpochRound[preSyncData.Epoch]; !ok {
		receiveTimeEpochRound[preSyncData.Epoch] = make(map[int]time.Time)
	}
	receiveTimeEpochRound[preSyncData.Epoch][preSyncData.Round] = time.Now()

	t1 := time.Now()
	filterAccount := preSyncData.Accounts
	if params.UseFilter {
		filterAccount = r.CurChain.FilterNilAccount(preSyncData.Accounts)
		header := []string{"epoch", "round", "account", "filterAccount"}
		err := writeDataToCSV(header, params.DataWrite_path+fmt.Sprintf("S%vN%vfilter.csv", r.RawShardID, r.RawNodeID), int(r.Epoch), r.PreSyncRound, len(preSyncData.Accounts), len(filterAccount))
		if err != nil {
			r.logger.Printf("writeDataToCSV error: %v \n", err)
		}
	}

	t2 := time.Now()
	r.logger.Printf("Step 1.3: Epoch %v, Round %v, rawAccount %v filterAccount %v \n",
		preSyncData.Epoch, preSyncData.Round, len(preSyncData.Accounts), len(filterAccount))

	r.SpecificSync(filterAccount, preSyncData.FromShard, preSyncData.Epoch, preSyncData.Round)
	t3 := time.Now()
	r.logger.Printf("Step 1.4: Epoch %v, Round %v, filterAccount cost %v, SpecificSync cost %v \n",
		preSyncData.Epoch, preSyncData.Round, t2.UnixMilli()-t1.UnixMilli(), t3.UnixMilli()-t2.UnixMilli())
	// 更新账户状态
	r.CurChain.UpdateAccountPool(preSyncData.Accounts)

	r.logger.Printf("Step 4: Epoch %v, Round %v sync data over, GetAccountSize %v \n", preSyncData.Epoch, preSyncData.Round, r.CurChain.GetAccountSize())
	if preSyncData.Round == 0 {
		go r.NotifyLeaderSyncOver()
	}
}

// SpecificSync the implementation of the specific sync, used to sync the loss account, 不是异步操作，会阻断
func (r *Reconfiguration) SpecificSync(accounts []string, destShard uint64, epoch uint64, round int) {
	beginSync := time.Now()

	r.logger.SetPrefix(fmt.Sprintf("Epoch %v, Round %v, ", r.Epoch, r.PreSyncRound))
	if _, ok := accountNumberEpochRound[epoch]; !ok {
		accountNumberEpochRound[epoch] = make(map[int]int)
	}
	accountNumberEpochRound[epoch][round] = len(accounts)
	if len(accounts) == 0 {
		r.logger.Printf("Step 2: SpecificSync: no account need to sync \n")
		r.ConfirmedRound.Store(int64(round))
		roundCheckLock.Signal()
		return
	}
	r.logger.Printf("Step 2: SpecificSync: chunk number is %v account %v need to sync, each account len %v \n", params.NChunk, len(accounts), len(accounts[0]))
	targetAddr := r.getSyncTargets(params.NChunk, destShard)
	r.logger.Printf("Targets address  %v \n", targetAddr)
	chunkIndexCond.L.Lock()
	for i := 0; i < params.NChunk; i++ {
		r.logger.Printf("Step 3: Epoch %v Round %v, blockHeight %v, blockHash %x wg add \n", epoch, round, r.CurChain.CurrentBlock.Header.Number, r.CurChain.CurrentBlock.Header.StateRoot)
		cidx := ChunkIndex{IdxChunk: i, Epoch: int(epoch), Round: round}

		if _, ok := chunkIndexPool[cidx]; ok {
			continue
		}

		s := &SpecificSync{
			FromAddress: r.Address,
			Epoch:       epoch,
			Round:       round,
			RootHash:    r.CurChain.CurrentBlock.Header.StateRoot,
			BlockHeight: int(r.CurChain.CurrentBlock.Header.Number),
			Accounts:    accounts,
			NChunk:      params.NChunk,
			IdxChunk:    i,
		}
		addr := targetAddr[i]
		r.logger.Printf("Step 2: Proposed begin send request to %v. \n", addr)
		marshal, err := json.Marshal(s)
		if err != nil {
			r.logger.Fatalf("Reconfiguration.SpecificSync error: %v \n", err)
		}
		msgSend := message.MergeMessage(CSpecificSync, marshal)
		go networks.TcpDial(msgSend, addr)
		r.logger.Printf("Step 2: send specific to leadr %v, wait for receive loss acount \n", addr)
	}

	check := func() bool {
		for i := 0; i < params.NChunk; i++ {
			cIdx := ChunkIndex{IdxChunk: i, Epoch: int(epoch), Round: round}
			if _, ok := chunkIndexPool[cIdx]; !ok {
				return true
			}
		}
		return false
	}
	for check() {
		chunkIndexCond.Wait()
	}
	chunkIndexCond.L.Unlock()

	overSyncTime := time.Now()
	r.logger.Printf("Step 2: Epoch %v Round %v, blockHeight %v, blockHash %x, sync over, cost %v \n",
		epoch, round, r.CurChain.CurrentBlock.Header.Number, r.CurChain.CurrentBlock.Header.StateRoot, overSyncTime.Sub(beginSync).Milliseconds())

	if round == 0 {
		r.CurChain.ClearAndUpdateAccountPool(accounts)
	} else {
		r.CurChain.UpdateAccountPool(accounts)
	}
	if r.ConfirmedRound.Load() < int64(round) || round == 0 {
		r.ConfirmedRound.Store(int64(round))
	}
	r.RecordEpochRoundNyIndx(epoch, round)
	roundCheckLock.Signal()

	r.logger.Printf("Step over: Round %v receive loss account, begin consensus  \n", round)
}

func (r *Reconfiguration) handleErasureData(content []byte) {
	r.logger.SetPrefix(fmt.Sprintf("Epoch %v, Round %v, ", r.Epoch, r.PreSyncRound))
	erasureChunk := &ErasureChunk{}
	err := json.Unmarshal(content, erasureChunk)
	if err != nil {
		r.logger.Fatalf("handleErasureData error: %v \n", err)
	}
	timeReceive := time.Now()
	r.logger.Printf("Step 3:Receive chunk, Epoch is %v, chunkN %v, chunk idx %v erasure %v chunk size %v chunk Epoch %v, timeReceive %v \n",
		r.Epoch, erasureChunk.NChunk, erasureChunk.IdxChunk, erasureChunk.ErasureId, len(erasureChunk.Keys), erasureChunk.Epoch, timeReceive.UnixMilli())

	// 转发纠删码数据
	if erasureChunk.SendTo == r.Address {
		recordProductorReceiveTime(erasureChunk.Epoch, erasureChunk.Round, erasureChunk.ProductorReceiveTime)
		recordProductorSendTime(erasureChunk.Epoch, erasureChunk.Round, erasureChunk.ProductorSendTime)
		for _, addr := range erasureChunk.Numbers {
			if addr == r.Address {
				continue
			}
			r.logger.Printf("Step 3: send chunk to %v \n", addr)
			marshal, err := json.Marshal(erasureChunk)
			if err != nil {
				r.logger.Fatalf("handleSpecificSync error: %v \n", err)
			}
			msgSend := message.MergeMessage(CErasureChunk, marshal)
			go networks.TcpDial(msgSend, addr)
		}
	}
	erasueCodeLock.Lock()
	defer erasueCodeLock.Unlock()

	cIdx := ChunkIndex{IdxChunk: erasureChunk.IdxChunk, Epoch: int(erasureChunk.Epoch), Round: erasureChunk.Round}
	erasureCodePool[cIdx] = append(erasureCodePool[cIdx], *erasureChunk)

	r.logger.Printf("Step 3: receive erasure, idx:%v len %v, over chunk ID %v \n", cIdx, len(erasureCodePool[cIdx]), erasureChunk.IdxChunk)
	if len(erasureCodePool[cIdx]) == len(erasureChunk.Numbers) || (params.UseErasure && len(erasureCodePool[cIdx]) == len(erasureChunk.Numbers)-1) {
		recordLocalReceiveTime(erasureChunk.Epoch, erasureChunk.Round)
		enough := time.Now()
		r.logger.Printf("%v \n", erasureChunk.Numbers)
		r.logger.Printf("Step 3: receive enough erasure, len %v, over chunk ID %v, time %v \n", len(erasureCodePool[cIdx]), erasureChunk.IdxChunk, enough.UnixMilli())
		allKeys := make([][]byte, 0)
		allValues := make([][]byte, 0)

		erasureData := make([][]byte, len(erasureCodePool[cIdx][0].Numbers))
		if params.UseErasure {
			r.logger.Printf("Step 3: Epoch %v Round %v, begin recover data \n", erasureCodePool[cIdx][0].Epoch, erasureCodePool[cIdx][0].Round)
			rawLen := 0
			ZoneSize := 0
			for _, erasure := range erasureCodePool[cIdx] {
				erasureData[erasure.ErasureId] = erasure.ErasureData
				rawLen = erasure.RawLen
				ZoneSize = len(erasure.Numbers)
			}
			t1 := time.Now()
			dataPair := chain.RecoverDataPair(erasureData, rawLen, ZoneSize)
			t2 := time.Now()
			r.logger.Printf("Step 3: Epoch %v Round %v, recover data over, recover cost %v \n",
				erasureCodePool[cIdx][0].Epoch, erasureCodePool[cIdx][0].Round, t2.Sub(t1).Milliseconds())
			allKeys = dataPair.Keys
			allValues = dataPair.Values
		} else {
			for _, erasure := range erasureCodePool[cIdx] {
				r.logger.Printf("erasure idx: %v", erasure.ErasureId)
				allKeys = append(allKeys, erasure.Keys...)
				allValues = append(allValues, erasure.Values...)
			}
		}

		r.UpdateTire(allKeys, allValues, erasureChunk.Epoch, erasureChunk.Round)
		r.logger.Printf("Step 3: Epoch %v Round %v, wg done, update cost %v \n", erasureCodePool[cIdx][0].Epoch, erasureCodePool[cIdx][0].Round, time.Now().Sub(enough).Milliseconds())
		//r.RecordEpochRoundNyIndx(erasureCodePool[cIdx][0].Epoch, erasureCodePool[cIdx][0].Round)
		// 清空，并继续下一轮
		erasureCodePool[cIdx] = make([]ErasureChunk, 0)
		chunkIndexCond.L.Lock()
		chunkIndexPool[cIdx] = struct{}{}
		chunkIndexCond.Broadcast()
		chunkIndexCond.L.Unlock()
	}

}

func (r *Reconfiguration) CheckSyncOver(blockHeight uint64) bool {
	if int(blockHeight)%(params.Frequency-1) == 1 {
		return true
	}
	t := int(blockHeight) % (params.Frequency - 1)
	if t == 0 {
		t = 5
	}
	roundCheckLock.L.Lock()
	r.logger.Printf("=====> checkSyncOver, block height %v, t:%v, now %v \n", blockHeight, t, r.ConfirmedRound.Load())
	for r.ConfirmedRound.Load() < int64(t)-1 {
		r.logger.Printf("===== checkSyncOver, wait for round %v, now %v \n", t-1, r.ConfirmedRound.Load())
		roundCheckLock.Wait()
	}
	roundCheckLock.L.Unlock()
	r.logger.Printf("=====> Over, block height %v, t:%v, now %v \n", blockHeight, t, r.ConfirmedRound.Load())
	return true
}

func (r *Reconfiguration) getSyncTargets(num int, shardID uint64) []string {
	var ans []string
	migrateNodeNumber := params.MigrateNodeNum
	t := int(int(r.RunningNode.NodeID-1)/params.ZoneSize) + migrateNodeNumber + 1
	ans = append(ans, r.ipNodeTable[shardID][uint64(t)])
	r.logger.Printf("NodeID %v, targetID %v getSyncTargets: %v \n", r.RunningNode.NodeID, t, r.ipNodeTable[shardID][uint64(t)])

	return ans
}

func (r *Reconfiguration) handleSpecificSyncData(content []byte) {

}

func (r *Reconfiguration) UpdateTire(keys, values [][]byte, epoch uint64, round int) {
	r.logger.SetPrefix(fmt.Sprintf("Epoch %v, Round %v, ", r.Epoch, r.PreSyncRound))
	dataLock.Lock()
	defer dataLock.Unlock()

	r.LevelDBLock.Lock()
	defer r.LevelDBLock.Unlock()

	dataSize := 0
	keySize := 0
	valSize := 0
	repeat := 0
	for i := 0; i < len(keys); i++ {
		keySize += len(keys[i])
		valSize += len(values[i])
		dataSize += len(keys[i]) + len(values[i])
		if UpdateHash[string(keys[i])] == 1 {
			repeat++
		} else {
			UpdateHash[string(keys[i])] = 1
		}
		err := r.db.Put(keys[i], values[i])
		if err != nil {
			r.logger.Fatalf("Reconfiguration error: %v \n", err)
		}
	}

	r.logger.Printf("Step 3: sync data over, len %v blob,total blob %v \n", len(keys), len(UpdateHash))

	// record the time
	if _, ok := overTimeEpochRound[epoch]; !ok {
		overTimeEpochRound[epoch] = make(map[int]time.Time)
	}
	overTimeEpochRound[epoch][round] = time.Now()

	if _, ok := stateDataSizeEpochRound[epoch]; !ok {
		stateDataSizeEpochRound[epoch] = make(map[int]int)
	}
	if _, ok := stateKeySizeEpochRound[epoch]; !ok {
		stateKeySizeEpochRound[epoch] = make(map[int]int)
	}
	if _, ok := stateValueSizeEpochRound[epoch]; !ok {
		stateValueSizeEpochRound[epoch] = make(map[int]int)
	}

	if _, ok := RepeatedEpochRound[epoch]; !ok {
		RepeatedEpochRound[epoch] = make(map[int]int)
	}
	stateDataSizeEpochRound[epoch][round] += dataSize
	stateKeySizeEpochRound[epoch][round] += keySize
	stateValueSizeEpochRound[epoch][round] += valSize
	RepeatedEpochRound[epoch][round] += repeat

	r.logger.Printf("Step 3: sync data over Epoch %v Round %v, ready to consensus \n", epoch, round)

}

func (r *Reconfiguration) RecordEpochRoundNyIndx(epoch uint64, round int) {
	RecordLock.Lock()
	defer RecordLock.Unlock()
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
		strconv.FormatUint(epoch, 10) +
			"," + strconv.FormatUint(uint64(round), 10) +
			"," + strconv.FormatUint(r.RunningNode.ShardID, 10) +
			"," + strconv.FormatUint(r.RunningNode.NodeID, 10) +
			"," + strconv.FormatUint(uint64(beginTimeEpochRound[epoch][round].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(receiveTimeEpochRound[epoch][round].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(overTimeEpochRound[epoch][round].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(overTimeEpochRound[epoch][round].UnixMilli()-beginTimeEpochRound[epoch][round].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(stateDataSizeEpochRound[epoch][round]), 10) +
			"," + strconv.FormatUint(uint64(blockSizeEpochRound[epoch][round]), 10) +
			"," + strconv.FormatUint(uint64(RepeatedEpochRound[epoch][round]), 10) +
			"," + strconv.FormatUint(uint64(stateKeySizeEpochRound[epoch][round]), 10) +
			"," + strconv.FormatUint(uint64(stateValueSizeEpochRound[epoch][round]), 10) +
			"," + strconv.FormatUint(uint64(accountNumberEpochRound[epoch][round]), 10) +
			"," + strconv.FormatUint(uint64(ProductorReceiveTime[epoch][round].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(ProductorSendTime[epoch][round].UnixMilli()), 10) +
			"," + strconv.FormatUint(uint64(LocalReceiveTime[epoch][round].UnixMilli()), 10) +
			"\n"); err != nil {
		log.Panic(err)
	}
}

func recordProductorReceiveTime(epoch uint64, round int, t time.Time) {
	RecordLock.Lock()
	defer RecordLock.Unlock()
	if _, ok := ProductorReceiveTime[epoch]; !ok {
		ProductorReceiveTime[epoch] = make(map[int]time.Time)
	}
	ProductorReceiveTime[epoch][round] = t
}

func recordProductorSendTime(epoch uint64, round int, t time.Time) {
	RecordLock.Lock()
	defer RecordLock.Unlock()
	if _, ok := ProductorSendTime[epoch]; !ok {
		ProductorSendTime[epoch] = make(map[int]time.Time)
	}
	ProductorSendTime[epoch][round] = t
}

func recordLocalReceiveTime(epoch uint64, round int) {
	RecordLock.Lock()
	defer RecordLock.Unlock()
	if _, ok := LocalReceiveTime[epoch]; !ok {
		LocalReceiveTime[epoch] = make(map[int]time.Time)
	}
	LocalReceiveTime[epoch][round] = time.Now()
}

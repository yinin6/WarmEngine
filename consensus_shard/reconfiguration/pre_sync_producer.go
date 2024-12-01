package reconfiguration

import (
	"blockEmulator/chain"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

var (
	Zones            = make(map[int][]string)
	node2Zone        = make(map[string]int)
	zoneLock         sync.RWMutex
	zoneConds        []*sync.Cond
	zoneOverChunkIdx = make(map[int]map[ChunkIndex]struct{})
)

var (
	RawDataSizeEpochRound  = make(map[uint64]map[int]int)
	slimDataSizeEpochRound = make(map[uint64]map[int]int)
	RawDataLock            sync.RWMutex
)

func RecordRawDataSize(epoch uint64, round, datasize int) {
	if _, ok := RawDataSizeEpochRound[epoch]; !ok {
		RawDataSizeEpochRound[epoch] = make(map[int]int)
	}
	RawDataSizeEpochRound[epoch][round] = datasize
}

func RecordSlimDataSize(epoch uint64, round, datasize int) {
	RawDataLock.Lock()
	defer RawDataLock.Lock()
	if _, ok := slimDataSizeEpochRound[epoch]; !ok {
		slimDataSizeEpochRound[epoch] = make(map[int]int)
	}
	slimDataSizeEpochRound[epoch][round] = datasize
}

func (r *Reconfiguration) handlePreSync(content []byte) {
	r.logger.SetPrefix(fmt.Sprintf("E%vR%v, ", r.Epoch, r.PreSyncRound))
	preSync := &PreSync{}
	err := networks.Decode(content, preSync)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.handlePreSync error: %v \n", err)
	}
	r.logger.Printf("Step 1: %v Receive %v pre-sync request, Req Epoch %v Round %v, now Epoch %v Round %v \n",
		r.RunningNode.IPaddr, preSync.FromAddress, preSync.Epoch, preSync.Cnt, r.Epoch, r.PreSyncRound)
	var account []string
	if preSync.Cnt == 0 {
		account = r.CurChain.PreSyncList(true, preSync.FromAddress, int(preSync.Epoch), preSync.Cnt)
	} else {
		account = r.CurChain.PreSyncList(false, preSync.FromAddress, int(preSync.Epoch), preSync.Cnt)
	}
	if r.Epoch != preSync.Epoch {
		r.logger.Printf("Step 1: %v Receive %v pre-sync request, Req Epoch %v Round %v, now Epoch %v Round %v \n",
			r.RunningNode.IPaddr, preSync.FromAddress, preSync.Epoch, preSync.Cnt, r.Epoch, r.PreSyncRound)
	}
	data := &PreSyncData{
		FromShard: r.RunningNode.ShardID,
		Epoch:     preSync.Epoch,
		Round:     preSync.Cnt,
		Accounts:  account,
		Address:   r.Address,
	}
	r.logger.Printf("Step 1.1: Worker %v, get pre-sync data, data len %v \n", preSync.FromAddress, len(data.Accounts))
	zoneLock.Lock()
	defer zoneLock.Unlock()

	if params.RecordRawData {
		keySize, valueSize, DataSize := r.CurChain.GetSpecificChunksSize(account, r.CurChain.CurrentBlock.Header.StateRoot)
		header := []string{"epoch", "round", "rawKeySize", "rawValueSize", "rawDataSize"}
		err := writeDataToCSV(header,
			params.DataWrite_path+fmt.Sprintf("S%vN%vfilterData.csv", r.RawShardID, r.RawNodeID),
			int(preSync.Epoch),
			preSync.Cnt,
			keySize, valueSize, DataSize)
		fmt.Println("write size: ", keySize, valueSize, DataSize)

		if err != nil {
			r.logger.Printf("writeDataToCSV error: %v \n", err)
		}
	}

	if preSync.Cnt == 0 {
		data.CurrBlock = r.CurChain.CurrentBlock.Encode()
		r.logger.Printf("接收到请求信息，来自： %v\n", preSync.FromAddress)
		if _, ok := node2Zone[preSync.FromAddress]; !ok {
			r.logger.Printf("还没有组成小组，当前小组： %v\n", node2Zone[preSync.FromAddress])
			// 找到当前分组
			groupID := len(Zones) - 1
			if groupID < 0 || len(Zones[groupID]) >= params.ZoneSize {
				groupID++
				zoneConds = append(zoneConds, sync.NewCond(&sync.Mutex{}))
			}
			// 将从节点添加到分组中
			Zones[groupID] = append(Zones[groupID], preSync.FromAddress)
			node2Zone[preSync.FromAddress] = groupID
			r.logger.Printf("Step 1.2: Worker %v 已加入 Group %v, now size %v\n", preSync.FromAddress, groupID, len(Zones[groupID]))

			// 如果该组满员
			if len(Zones[groupID]) == params.ZoneSize {
				r.logger.Printf("Step 1.3: 第 %d 组已满，通知该组从节点, 组中的成员数量 %d \n", groupID, len(Zones[groupID]))
				zoneConds[groupID].Broadcast()
			}
			r.logger.Printf("over add zone \n")
		} else {
			r.logger.Printf("have zone \n")
		}
	}

	r.logger.Printf("Step 1.4: zoneID %v, Get sync list, ready to send to %v len %v \n", node2Zone[preSync.FromAddress], preSync.FromAddress, len(data.Accounts))
	marshal, err := json.Marshal(data)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.handlePreSync error: %v \n", err)
	}
	msgSend := message.MergeMessage(CPreSyncData, marshal)
	go networks.TcpDial(msgSend, preSync.FromAddress)
}

func getZoneID(address string) int {
	zoneLock.RLock()
	defer zoneLock.RUnlock()
	return node2Zone[address]
}

func (r *Reconfiguration) handleSpecificSync(content []byte) {
	r.logger.SetPrefix(fmt.Sprintf("E%vR%v, ", r.Epoch, r.PreSyncRound))
	r.logger.Printf("Step 2: handle SpecificSync begin \n")
	specificSync := &SpecificSync{}
	err := json.Unmarshal(content, specificSync)
	if err != nil {
		r.logger.Fatalf("handleSpecificSync error: %v \n", err)
	}
	r.logger.Printf("Step 2: receive %v accounts need to get. \n", len(specificSync.Accounts))
	specificSyncReceiveTime := time.Now()

	if r.RunningNode.NodeID != 0 {
		r.logger.Printf("接收到请求信息，来自： %v\n", specificSync.FromAddress)
		if _, ok := node2Zone[specificSync.FromAddress]; !ok {
			r.logger.Printf("还没有组成小组，当前小组： %v\n", node2Zone[specificSync.FromAddress])
			// 找到当前分组
			groupID := len(Zones) - 1
			if groupID < 0 || len(Zones[groupID]) >= params.ZoneSize {
				groupID++
				zoneConds = append(zoneConds, sync.NewCond(&sync.Mutex{}))
			}
			// 将从节点添加到分组中
			Zones[groupID] = append(Zones[groupID], specificSync.FromAddress)
			node2Zone[specificSync.FromAddress] = groupID
			r.logger.Printf("Step 1.2: Worker %v 已加入 Group %v, now size %v\n", specificSync.FromAddress, groupID, len(Zones[groupID]))

			// 如果该组满员
			if len(Zones[groupID]) == params.ZoneSize {
				r.logger.Printf("Step 1.3: 第 %d 组已满，通知该组从节点, 组中的成员数量 %d \n", groupID, len(Zones[groupID]))
				zoneConds[groupID].Broadcast()
			}
			r.logger.Printf("over add zone \n")
		} else {
			r.logger.Printf("have zone \n")
		}
	}

	// wait for zone is enough
	zoneID := getZoneID(specificSync.FromAddress)
	r.logger.Printf("Step 2: zoneID: %v \n", zoneID)

	cond := zoneConds[zoneID]
	cond.L.Lock()
	defer cond.L.Unlock()
	for len(Zones[zoneID]) < params.ZoneSize {
		cond.Wait()
	}
	cidx := ChunkIndex{
		IdxChunk: specificSync.IdxChunk,
		Epoch:    int(specificSync.Epoch),
		Round:    specificSync.Round,
	}
	if _, ok := zoneOverChunkIdx[zoneID]; !ok {
		zoneOverChunkIdx[zoneID] = make(map[ChunkIndex]struct{})
	}

	if _, ok := zoneOverChunkIdx[zoneID][cidx]; ok {
		r.logger.Printf("Step 2: %v not first com %v, return  \n", specificSync.FromAddress, Zones[zoneID])
		return
	} else {
		zoneOverChunkIdx[zoneID][cidx] = struct{}{}
	}

	z := Zones[zoneID]
	r.logger.Printf("fetch data Round %v Epoch %v, Root %x, blockHeight %v \n", specificSync.Round, specificSync.Epoch, specificSync.RootHash, specificSync.BlockHeight)

	zoneOverTime := time.Now()

	Index2Chunk := chain.ChunkIndex{
		IdxChunk: specificSync.IdxChunk,
		Epoch:    int(specificSync.Epoch),
		Round:    specificSync.Round,
		ZoneID:   zoneID,
	}

	t := specificSync.RootHash
	if uint64(specificSync.BlockHeight) > r.CurChain.CurrentBlock.Header.Number {
		t = r.CurChain.CurrentBlock.Header.StateRoot
		r.logger.Printf("Step 2: block height is not enough, use my %v \n", r.CurChain.CurrentBlock.Header.Number)
	}
	chunkGetTime := time.Now()

	for idx := 0; idx < params.ZoneSize; idx++ {
		r.logger.Printf("Step 2: zone number is enough %v, send Data to %v, idx %v  \n", Zones[zoneID], specificSync.FromAddress, idx)
		chunk := r.CurChain.GetSpecificSingleChunk(specificSync.Accounts, t, params.ZoneSize, idx, Index2Chunk)
		r.logger.Printf("Step 2:specificSync.Epoch %v Round %v, get specific data, data len %v \n",
			specificSync.Epoch, specificSync.Round, len(chunk.Keys))
		r.logger.Printf("Step 2: chunk root  %x  \n", chunk.Root)

		//todo 现在还有使用纠删码，尝试将原始数据进行切分，查看结果是否正确。
		erasureChunk := &ErasureChunk{
			SendTo:               z[idx],
			Epoch:                specificSync.Epoch,
			Round:                specificSync.Round,
			Keys:                 chunk.Keys,
			Values:               chunk.Values,
			NChunk:               specificSync.NChunk,
			IdxChunk:             specificSync.IdxChunk,
			ErasureId:            idx,
			ErasureData:          chunk.ErasureData,
			Numbers:              z,
			RawLen:               chunk.RawLen,
			ProductorReceiveTime: specificSyncReceiveTime,
			ProductorSendTime:    time.Now(),
		}
		marshal, err := json.Marshal(erasureChunk)
		if err != nil {
			r.logger.Fatalf("handleSpecificSync error: %v \n", err)
		}

		msgSend := message.MergeMessage(CErasureChunk, marshal)
		networks.TcpDial(msgSend, z[idx])
	}
	chunkSendTime := time.Now()
	r.logger.Printf("receivetime %v,zoneOverTime %v, chunkGetTime %v, chunkSendTime %v  ",
		specificSyncReceiveTime.UnixMilli(), zoneOverTime.UnixMilli(), chunkGetTime.UnixMilli(), chunkSendTime.UnixMilli())

	r.logger.Printf("zone cost %v, chunk cost %v", zoneOverTime.Sub(specificSyncReceiveTime).Milliseconds(), chunkSendTime.Sub(chunkGetTime).Milliseconds())

}

// addtional module for new consensus
package pbft_all

import (
	"blockEmulator/consensus_shard/reconfiguration"
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

// simple implementation of pbftHandleModule interface ...
// only for block request and use transaction relay
type RawRelayPbftExtraHandleMod struct {
	pbftNode *PbftConsensusNode
	// pointer to pbft data
}

var (
	TxCnt2Epoch       = make(map[uint64]float64)
	beginTime2Epoch   = make(map[uint64]time.Time)
	EpochTps          = make(map[uint64]float64)
	beginTime2Shuffle = make(map[uint64]time.Time)
	overTime2Shuffle  = make(map[uint64]time.Time)
)

// propose request with different types
func (rphm *RawRelayPbftExtraHandleMod) HandleinPropose() (bool, *message.Request) {
	if rphm.pbftNode.sequenceID%uint64(params.Frequency) == 0 {
		rphm.pbftNode.reconfigStage.Store(1)
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : seqID is %v,ready to shuffle \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, rphm.pbftNode.sequenceID)
		table := rphm.pbftNode.Reconfig.ShuffleIpTable(rphm.pbftNode.ip_nodeTable)
		beginTime2Shuffle[rphm.pbftNode.Epoch+1] = time.Now()

		// make sure all shard to new view
		p := rphm.pbftNode
		epoch := uint64(rphm.pbftNode.sequenceID / uint64(params.Frequency))
		if rphm.pbftNode.NodeID == uint64(rphm.pbftNode.view.Load()) {
			p.pl.Plog.Printf("S%dN%d : is ready to shuffle (update iptable), wait other shard over, epoch %v \n", p.ShardID, p.NodeID, epoch)
			shuffleReadyMessage := reconfiguration.ShuffleReadyMessage{
				ShardID: p.ShardID,
				Epoch:   epoch,
			}
			shuffleRequestByte, err := json.Marshal(shuffleReadyMessage)
			if err != nil {
				p.pl.Plog.Panic()
			}
			msgSend := message.MergeMessage(reconfiguration.CShuffleReady, shuffleRequestByte)
			networks.Broadcast(p.RunningNode.IPaddr, p.getLeaderNodes(), msgSend)
			networks.TcpDial(msgSend, p.RunningNode.IPaddr)
			p.pl.Plog.Printf("S%dN%d : send shuffle ready for other shard, total %v \n", p.ShardID, p.NodeID, len(p.getLeaderNodes()))
		}

		value, _ := p.Epoch2ReadyShard.Load(epoch)
		for value != params.ShardNum {
			time.Sleep(time.Millisecond * 500)
			value, _ = p.Epoch2ReadyShard.Load(epoch)
			p.pl.Plog.Printf("S%dN%d : Epoch %v wait for all shard ready, now %v, need %v \n", p.ShardID, p.NodeID, epoch, value, params.ShardNum)
		}

		encode, err := networks.Encode(table)
		if err != nil {
			log.Panic("encode error")
		}
		r := &message.Request{
			RequestType: reconfiguration.ShuffleRequest,
			ReqTime:     time.Now(),
		}
		r.Msg.Content = encode
		return true, r
	}

	if rphm.pbftNode.sequenceID%uint64(params.Frequency) == 1 {
		beginTime2Epoch[rphm.pbftNode.Epoch] = time.Now() // 重组结束的时间
		overTime2Shuffle[rphm.pbftNode.Epoch] = time.Now()
		if rphm.pbftNode.Reconfig.Epoch > 0 {
			t := beginTime2Epoch[rphm.pbftNode.Epoch].Sub(beginTime2Epoch[rphm.pbftNode.Epoch-1])
			EpochTps[rphm.pbftNode.Reconfig.Epoch-1] = TxCnt2Epoch[rphm.pbftNode.Epoch-1] / t.Seconds()
			strs := []string{"epoch", "beginTime", "txs", "tps"}
			vals := []string{
				strconv.Itoa(int(rphm.pbftNode.Reconfig.Epoch - 1)),
				strconv.FormatUint(uint64(beginTime2Epoch[rphm.pbftNode.Epoch-1].UnixMilli()), 10),
				strconv.FormatFloat(TxCnt2Epoch[rphm.pbftNode.Epoch-1], 'f', -1, 64),
				strconv.FormatFloat(EpochTps[rphm.pbftNode.Epoch-1], 'f', -1, 64),
			}
			rphm.pbftNode.writeCSVlineByName("epochDatil.csv", strs, vals)

			// write shuffle time
			strs1 := []string{"epoch", "beginTime", "overTime", "cost"}
			vals1 := []string{
				strconv.Itoa(int(rphm.pbftNode.Reconfig.Epoch)),
				strconv.FormatUint(uint64(beginTime2Shuffle[rphm.pbftNode.Epoch].UnixMilli()), 10),
				strconv.FormatUint(uint64(overTime2Shuffle[rphm.pbftNode.Epoch].UnixMilli()), 10),
				strconv.FormatInt(overTime2Shuffle[rphm.pbftNode.Epoch].Sub(beginTime2Shuffle[rphm.pbftNode.Epoch]).Milliseconds(), 10),
			}
			rphm.pbftNode.writeCSVlineByName("shuffeDatil.csv", strs1, vals1)

		}
	}

	// new blocks
	block := rphm.pbftNode.CurChain.GenerateBlock(int32(rphm.pbftNode.NodeID))
	r := &message.Request{
		RequestType: message.BlockRequest,
		ReqTime:     time.Now(),
	}
	r.Msg.Content = block.Encode()

	return true, r
}

// the DIY operation in preprepare
func (rphm *RawRelayPbftExtraHandleMod) HandleinPrePrepare(ppmsg *message.PrePrepare) bool {
	if ppmsg.RequestMsg.RequestType == reconfiguration.ShuffleRequest {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : the pre-prepare message is shuffle request\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
		return true
	}

	//if params.SyncMod == 2 && rphm.pbftNode.Reconfig.IsSelectedPreSync && rphm.pbftNode.Reconfig.PreSyncRound > 0 {
	//	rphm.pbftNode.pl.Plog.Printf("S%dN%d : the pre-sync request, Round %v, hash %x height %v \n",
	//		rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, rphm.pbftNode.Reconfig.PreSyncRound,
	//		rphm.pbftNode.CurChain.CurrentBlock.Header.StateRoot, rphm.pbftNode.CurChain.CurrentBlock.Header.Number)
	//	rphm.pbftNode.Reconfig.PreSync()
	//}

	if rphm.pbftNode.CurChain.IsValidBlock(core.DecodeB(ppmsg.RequestMsg.Msg.Content)) != nil {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : not a valid block\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
		return false
	}
	rphm.pbftNode.pl.Plog.Printf("S%dN%d : the pre-prepare Seq %v message is correct, putting it into the RequestPool. \n",
		rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, ppmsg.SeqID)
	rphm.pbftNode.requestPool[string(ppmsg.Digest)] = ppmsg.RequestMsg
	// merge to be a prepare message
	return true
}

// the operation in prepare, and in pbft + tx relaying, this function does not need to do any.
func (rphm *RawRelayPbftExtraHandleMod) HandleinPrepare(pmsg *message.Prepare) bool {
	fmt.Println("No operations are performed in Extra handle mod")
	return true
}

// the operation in commit.
func (rphm *RawRelayPbftExtraHandleMod) HandleinCommit(cmsg *message.Commit) bool {
	r := rphm.pbftNode.requestPool[string(cmsg.Digest)]
	if r.RequestType == reconfiguration.ShuffleRequest {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : the commit message is shuffle request\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
		shuffleData := &reconfiguration.ShuffleData{}
		err := networks.Decode(r.Msg.Content, shuffleData)
		if err != nil {
			log.Panic("Decode error")
		}
		rphm.pbftNode.Epoch++
		isSelected, destShard := rphm.pbftNode.Shuffle(shuffleData.IpNodeTable)
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : shuffle is selected %v, destShrd %v \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, isSelected, destShard)

		if isSelected {
			switch params.SyncMod {
			case 0:
				rphm.pbftNode.Reconfig.FullSync(destShard)
			case 1:
				rphm.pbftNode.Reconfig.TMptSync(destShard)
			case 2:
				rphm.pbftNode.Reconfig.PreSync()
			case 3:
				rphm.pbftNode.Reconfig.TrieSync(destShard)

			}
		} else {
			rphm.pbftNode.Reconfig.NotifyLeaderSyncOver()
		}

		rphm.pbftNode.pl.Plog.Printf("S%dN%d :is selected %v, dest shard %v shuffle success\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, isSelected, destShard)
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : new ip_nodeTable is %v\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, rphm.pbftNode.ip_nodeTable)
		return true
	}
	// requestType block
	block := core.DecodeB(r.Msg.Content)
	rphm.pbftNode.pl.Plog.Printf("S%dN%d : adding the block %d...now height = %d \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, block.Header.Number, rphm.pbftNode.CurChain.CurrentBlock.Header.Number)

	// 提供数据的节点
	if rphm.pbftNode.NodeID > uint64(params.MigrateNodeNum) {
		rphm.pbftNode.CurChain.Txpool.RemoveTxsByHash(block.Body)

		metricName := []string{
			"Block Height",
			"TxPool Size",
			"# of all Txs in this block",
		}
		metricVal := []string{
			strconv.Itoa(int(block.Header.Number)),
			strconv.Itoa(len(rphm.pbftNode.CurChain.Txpool.TxQueue)),
			strconv.Itoa(len(block.Body)),
		}
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : removed txs in txpool\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
		rphm.pbftNode.writeCSVline(metricName, metricVal)

		if params.SyncMod == 1 {
			for _, tx := range block.Body {
				ssid := rphm.pbftNode.CurChain.Get_PartitionMap(tx.Sender)
				rsid := rphm.pbftNode.CurChain.Get_PartitionMap(tx.Recipient)
				// record account
				if ssid == rphm.pbftNode.ShardID {
					rphm.pbftNode.CurChain.TmptRecordAccount(rphm.pbftNode.Reconfig.Epoch, tx.Sender)
				}
				if rsid == rphm.pbftNode.ShardID {
					rphm.pbftNode.CurChain.TmptRecordAccount(rphm.pbftNode.Reconfig.Epoch, tx.Recipient)
				}
			}
		}

		if params.SyncMod == 3 {
			for _, tx := range block.Body {
				ssid := rphm.pbftNode.CurChain.Get_PartitionMap(tx.Sender)
				rsid := rphm.pbftNode.CurChain.Get_PartitionMap(tx.Recipient)
				// record account
				if ssid == rphm.pbftNode.ShardID {
					rphm.pbftNode.CurChain.RecordTrie(tx.Sender)
				}
				if rsid == rphm.pbftNode.ShardID {
					rphm.pbftNode.CurChain.RecordTrie(tx.Recipient)
				}
			}
		}
	}

	// specific sync to get loss account
	if params.SyncMod == 1 && rphm.pbftNode.Reconfig.IsSelectedPreSync {
		lossAcount := []string{}
		for _, tx := range block.Body {
			ssid := rphm.pbftNode.CurChain.Get_PartitionMap(tx.Sender)
			rsid := rphm.pbftNode.CurChain.Get_PartitionMap(tx.Recipient)
			if ssid == rphm.pbftNode.ShardID {
				if !rphm.pbftNode.CurChain.TmptHaveAccount(rphm.pbftNode.Reconfig.Epoch, tx.Sender) {
					lossAcount = append(lossAcount, tx.Sender)
				}
			}
			if rsid == rphm.pbftNode.ShardID {
				if !rphm.pbftNode.CurChain.TmptHaveAccount(rphm.pbftNode.Reconfig.Epoch, tx.Recipient) {
					lossAcount = append(lossAcount, tx.Recipient)
				}
			}
		}
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : now have account %v loss account %v\n",
			rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, rphm.pbftNode.CurChain.TmptAccountSize(rphm.pbftNode.Reconfig.Epoch), len(lossAcount))
		rphm.pbftNode.Reconfig.TMPTSpecificSync(lossAcount)
	}

	// pre-sync to get loss account
	if params.SyncMod == 2 && rphm.pbftNode.Reconfig.IsSelectedPreSync {

		rphm.pbftNode.pl.Plog.Printf("S%dN%d : check if the pre-sync request over, blockHeight %v \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, block.Header.Number)
		if rphm.pbftNode.Reconfig.CheckSyncOver(block.Header.Number) {
			rphm.pbftNode.pl.Plog.Printf("S%dN%d : the pre-sync is over\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
		}
		tempKV := make(map[int]int)
		lossAcount := []string{}
		for _, tx := range block.Body {
			tempKV[tx.VisitCnt] += 1
			tempKV[tx.Epoch] += 1
			tempKV[tx.Round] += 1
			ssid := rphm.pbftNode.CurChain.Get_PartitionMap(tx.Sender)
			rsid := rphm.pbftNode.CurChain.Get_PartitionMap(tx.Recipient)
			if ssid == rphm.pbftNode.ShardID {
				if !rphm.pbftNode.CurChain.HaveAccount(tx.Sender) {
					lossAcount = append(lossAcount, tx.Sender)
				}
			}
			if rsid == rphm.pbftNode.ShardID {
				if !rphm.pbftNode.CurChain.HaveAccount(tx.Recipient) {
					lossAcount = append(lossAcount, tx.Recipient)
				}
			}
		}
		for k, v := range tempKV {
			rphm.pbftNode.pl.Plog.Printf("Data of Txs : the key %v, value %v\n", k, v)
		}

		if len(lossAcount) > 0 {
			rphm.pbftNode.pl.Plog.Printf("S%dN%d : now have account %v loss account %v\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, rphm.pbftNode.CurChain.GetAccountSize(), len(lossAcount))
			rphm.pbftNode.Reconfig.SpecificSync(lossAcount, rphm.pbftNode.ShardID, rphm.pbftNode.Reconfig.Epoch, rphm.pbftNode.Reconfig.PreSyncRound*params.Frequency)
		}
	}

	rphm.pbftNode.Reconfig.LevelDBLock.Lock()
	rphm.pbftNode.CurChain.AddBlock(block)
	rphm.pbftNode.Reconfig.LevelDBLock.Unlock()

	rphm.pbftNode.pl.Plog.Printf("S%dN%d : added the block %d... \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, block.Header.Number)
	rphm.pbftNode.CurChain.PrintBlockChain()

	//todo pre sync for next Round
	if params.SyncMod == 2 && rphm.pbftNode.Reconfig.IsSelectedPreSync && rphm.pbftNode.Reconfig.PreSyncRound > 0 {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : the pre-sync request, Round %v, hash %x height %v \n",
			rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, rphm.pbftNode.Reconfig.PreSyncRound,
			rphm.pbftNode.CurChain.CurrentBlock.Header.StateRoot, rphm.pbftNode.CurChain.CurrentBlock.Header.Number)
		rphm.pbftNode.Reconfig.PreSync()
	}

	// now try to relay txs to other shards (for main nodes)
	if rphm.pbftNode.NodeID == uint64(rphm.pbftNode.view.Load()) {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : main node is trying to send relay txs at height = %d \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, block.Header.Number)
		// generate relay pool and collect txs excuted
		rphm.pbftNode.CurChain.Txpool.RelayPool = make(map[uint64][]*core.Transaction)
		interShardTxs := make([]*core.Transaction, 0)
		relay1Txs := make([]*core.Transaction, 0)
		relay2Txs := make([]*core.Transaction, 0)
		for _, tx := range block.Body {
			ssid := rphm.pbftNode.CurChain.Get_PartitionMap(tx.Sender)
			rsid := rphm.pbftNode.CurChain.Get_PartitionMap(tx.Recipient)

			if !tx.Relayed && ssid != rphm.pbftNode.ShardID {
				log.Panic("incorrect tx")
			}
			if tx.Relayed && rsid != rphm.pbftNode.ShardID {
				log.Panic("incorrect tx")
			}
			if rsid != rphm.pbftNode.ShardID {
				relay1Txs = append(relay1Txs, tx)
				tx.Relayed = true
				rphm.pbftNode.CurChain.Txpool.AddRelayTx(tx, rsid)
			} else {
				if tx.Relayed {
					relay2Txs = append(relay2Txs, tx)
				} else {
					interShardTxs = append(interShardTxs, tx)
				}
			}
		}

		TxCnt2Epoch[rphm.pbftNode.Epoch] += float64(len(interShardTxs) + (len(relay1Txs)+len(relay2Txs))/2)

		// send relay txs
		if params.RelayWithMerkleProof == 1 {
			rphm.pbftNode.RelayWithProofSend(block)
		} else {
			rphm.pbftNode.RelayMsgSend()
		}

		// send txs excuted in this block to the listener
		// add more message to measure more metrics
		bim := message.BlockInfoMsg{
			BlockBodyLength: len(block.Body),
			InnerShardTxs:   interShardTxs,
			Epoch:           0,

			Relay1Txs: relay1Txs,
			Relay2Txs: relay2Txs,

			SenderShardID: rphm.pbftNode.ShardID,
			ProposeTime:   r.ReqTime,
			CommitTime:    time.Now(),
		}
		bByte, err := json.Marshal(bim)
		if err != nil {
			log.Panic()
		}
		msg_send := message.MergeMessage(message.CBlockInfo, bByte)
		go networks.TcpDial(msg_send, rphm.pbftNode.ip_nodeTable[params.SupervisorShard][0])
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : sended excuted txs\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
		rphm.pbftNode.CurChain.Txpool.GetLocked()
		metricName := []string{
			"Block Height",
			"EpochID of this block",
			"TxPool Size",
			"# of all Txs in this block",
			"# of Relay1 Txs in this block",
			"# of Relay2 Txs in this block",
			"TimeStamp - Propose (unixMill)",
			"TimeStamp - Commit (unixMill)",

			"SUM of confirm latency (ms, All Txs)",
			"SUM of confirm latency (ms, Relay1 Txs) (Duration: Relay1 proposed -> Relay1 Commit)",
			"SUM of confirm latency (ms, Relay2 Txs) (Duration: Relay1 proposed -> Relay2 Commit)",
		}
		metricVal := []string{
			strconv.Itoa(int(block.Header.Number)),
			strconv.Itoa(bim.Epoch),
			strconv.Itoa(len(rphm.pbftNode.CurChain.Txpool.TxQueue)),
			strconv.Itoa(len(block.Body)),
			strconv.Itoa(len(relay1Txs)),
			strconv.Itoa(len(relay2Txs)),
			strconv.FormatInt(bim.ProposeTime.UnixMilli(), 10),
			strconv.FormatInt(bim.CommitTime.UnixMilli(), 10),

			strconv.FormatInt(computeTCL(block.Body, bim.CommitTime), 10),
			strconv.FormatInt(computeTCL(relay1Txs, bim.CommitTime), 10),
			strconv.FormatInt(computeTCL(relay2Txs, bim.CommitTime), 10),
		}
		rphm.pbftNode.writeCSVline(metricName, metricVal)
		rphm.pbftNode.CurChain.Txpool.GetUnlocked()
	}
	return true
}

func (rphm *RawRelayPbftExtraHandleMod) HandleReqestforOldSeq(*message.RequestOldMessage) bool {
	fmt.Println("No operations are performed in Extra handle mod")
	return true
}

// the operation for sequential requests
func (rphm *RawRelayPbftExtraHandleMod) HandleforSequentialRequest(som *message.SendOldMessage) bool {
	if int(som.SeqEndHeight-som.SeqStartHeight+1) != len(som.OldRequest) {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : the SendOldMessage message is not enough\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
	} else { // add the block into the node pbft blockchain
		for height := som.SeqStartHeight; height <= som.SeqEndHeight; height++ {
			r := som.OldRequest[height-som.SeqStartHeight]
			if r.RequestType == message.BlockRequest {
				b := core.DecodeB(r.Msg.Content)
				rphm.pbftNode.CurChain.AddBlock(b)
			}
		}
		rphm.pbftNode.sequenceID = som.SeqEndHeight + 1
		rphm.pbftNode.CurChain.PrintBlockChain()
	}
	return true
}

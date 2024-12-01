package pbft_all

import (
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"blockEmulator/shard"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// this func is only invoked by main node
func (p *PbftConsensusNode) Propose() {
	// wait other nodes to start TCPlistening, sleep 5 sec.
	time.Sleep(5 * time.Second)

	nextRoundBeginSignal := make(chan bool)

	xSeq := p.sequenceID

	go func() {
		// go into the next round
		time.Sleep(time.Duration(int64(p.pbftChainConfig.BlockInterval)) * time.Millisecond)
		nextRoundBeginSignal <- true
		for {
			time.Sleep(time.Duration(int64(p.pbftChainConfig.BlockInterval)) * time.Millisecond)
			// send a signal to another GO-Routine. It will block until a GO-Routine try to fetch data from this channel.
			for p.pbftStage.Load() != 1 {
				time.Sleep(time.Millisecond * 100)
			}

			nextRoundBeginSignal <- true
		}
	}()

	//go func() {
	//	// check whether to view change
	//	for {
	//		time.Sleep(time.Second)
	//		if time.Now().UnixMilli()-p.lastCommitTime.Load() > int64(params.PbftViewChangeTimeOut) {
	//			p.lastCommitTime.Store(time.Now().UnixMilli())
	//			go p.viewChangePropose()
	//		}
	//	}
	//}()

	for {
		select {
		case <-nextRoundBeginSignal:
			go func() {
				// if this node is not leader, do not propose.
				if uint64(p.view.Load()) != p.NodeID {
					return
				}

				p.sequenceLock.Lock()
				p.pl.Plog.Printf("S%dN%d get sequenceLock locked, now trying to propose...\n", p.ShardID, p.NodeID)

				if xSeq != 1 {
					p.ReplyListCond.L.Lock()
					for len(p.ReplyList[xSeq]) != len(p.ip_nodeTable[p.ShardID]) {
						p.pl.Plog.Printf("S%vN%v: wait for reply, xsequence %v now sequenceID %v, receive reply %v \n",
							p.ShardID, p.NodeID, xSeq, p.sequenceID, len(p.ReplyList[xSeq]))
						p.ReplyListCond.Wait()
					}
					p.pl.Plog.Printf("S%vN%v: OVER RECEIVE OF reply, xsequence %v now sequenceID %v, receive reply %v \n",
						p.ShardID, p.NodeID, xSeq, p.sequenceID, len(p.ReplyList[xSeq]))
					p.ReplyListCond.L.Unlock()
					xSeq++
				} else {
					xSeq++
				}

				if p.reconfigStage.Load() == 1 {

					p.pl.Plog.Printf("S%dN%d : all shard has shuffled ready, wait to sync data over \n", p.ShardID, p.NodeID)

					epcoh := uint64(int(p.CurChain.CurrentBlock.Header.Number) / (params.Frequency - 1))

					p.pl.Plog.Printf("curr height %v, epoch %v \n", p.CurChain.CurrentBlock.Header.Number, epcoh)

					value, _ := p.OverSyncInShard.Load(epcoh)
					// begin to sync data
					for value != params.NodesInShard {
						// 使用 Range 遍历 sync.Map 中的所有键值对
						p.OverSyncInShard.Range(func(key, value any) bool {
							p.pl.Plog.Printf("key: %v, value: %v\n", key, value)
							return true // 返回 true 表示继续遍历，返回 false 表示停止遍历
						})
						time.Sleep(time.Millisecond * 500)
						p.pl.Plog.Printf("S%dN%d : Epoch %v shard wait for all node over sync data, now %v \n", p.ShardID, p.NodeID, epcoh, value)
						value, _ = p.OverSyncInShard.Load(epcoh)
					}

					// recover
					p.reconfigStage.Store(0)

					if params.SyncMod == 1 {
						p.CurChain.ClearAccountPool()
					}
				}

				// propose
				// implement interface to generate propose
				_, r := p.ihm.HandleinPropose()

				digest := getDigest(r)
				p.requestPool[string(digest)] = r
				p.pl.Plog.Printf("S%dN%d put the request into the pool ...\n", p.ShardID, p.NodeID)

				ppmsg := message.PrePrepare{
					RequestMsg: r,
					Digest:     digest,
					SeqID:      p.sequenceID,
				}
				p.height2Digest[p.sequenceID] = string(digest)
				// marshal and broadcast
				ppbyte, err := json.Marshal(ppmsg)
				if err != nil {
					log.Panic()
				}
				msg_send := message.MergeMessage(message.CPrePrepare, ppbyte)
				networks.Broadcast(p.RunningNode.IPaddr, p.getNeighborNodes(), msg_send)
				networks.TcpDial(msg_send, p.RunningNode.IPaddr)
				p.pbftStage.Store(2)
			}()

		case <-p.pStop:
			p.pl.Plog.Printf("S%dN%d get stopSignal in Propose Routine, now stop...\n", p.ShardID, p.NodeID)
			return
		}
	}
}

// Handle pre-prepare messages here.
// If you want to do more operations in the pre-prepare stage, you can implement the interface "ExtraOpInConsensus",
// and call the function: **ExtraOpInConsensus.HandleinPrePrepare**
func (p *PbftConsensusNode) handlePrePrepare(content []byte) {
	p.RunningNode.PrintNode()

	// decode the message
	ppmsg := new(message.PrePrepare)
	err := json.Unmarshal(content, ppmsg)
	if err != nil {
		p.pl.Plog.Panic(err)
	}
	p.pl.Plog.Printf("S%dN%d : received the PrePrepare ..., message SeqID %v ,new SeqID %v \n", p.ShardID, p.NodeID, ppmsg.SeqID, p.sequenceID)
	curView := p.view.Load()
	p.pbftLock.Lock()
	defer p.pbftLock.Unlock()
	for p.pbftStage.Load() < 1 && ppmsg.SeqID >= p.sequenceID && p.view.Load() == curView {
		p.pl.Plog.Printf("error seqID %v, sequenceID %v, view %v, curView %v, stage %v \n", ppmsg.SeqID, p.sequenceID, p.view.Load(), curView, p.pbftStage.Load())
		p.conditionalVarpbftLock.Wait()
	}
	defer p.conditionalVarpbftLock.Broadcast()

	//// if this message is out of date, return.
	if ppmsg.SeqID < p.sequenceID || p.view.Load() != curView {
		return
	}

	flag := false
	if digest := getDigest(ppmsg.RequestMsg); string(digest) != string(ppmsg.Digest) {
		p.pl.Plog.Printf("S%dN%d : the digest is not consistent, so refuse to prepare. \n", p.ShardID, p.NodeID)
	} else if p.sequenceID < ppmsg.SeqID {
		p.requestPool[string(getDigest(ppmsg.RequestMsg))] = ppmsg.RequestMsg
		p.height2Digest[ppmsg.SeqID] = string(getDigest(ppmsg.RequestMsg))
		p.pl.Plog.Printf("S%dN%d : the Sequence id is not consistent, so refuse to prepare. \n", p.ShardID, p.NodeID)
	} else {
		// do your operation in this interface
		flag = p.ihm.HandleinPrePrepare(ppmsg)
		p.requestPool[string(getDigest(ppmsg.RequestMsg))] = ppmsg.RequestMsg
		p.height2Digest[ppmsg.SeqID] = string(getDigest(ppmsg.RequestMsg))
	}
	// if the message is true, broadcast the prepare message
	if flag {
		pre := message.Prepare{
			Digest:     ppmsg.Digest,
			SeqID:      ppmsg.SeqID,
			SenderNode: p.RunningNode,
		}
		prepareByte, err := json.Marshal(pre)
		if err != nil {
			log.Panic()
		}
		// broadcast
		msg_send := message.MergeMessage(message.CPrepare, prepareByte)
		networks.Broadcast(p.RunningNode.IPaddr, p.getNeighborNodes(), msg_send)
		networks.TcpDial(msg_send, p.RunningNode.IPaddr)
		p.pl.Plog.Printf("S%dN%d : has broadcast the prepare message \n", p.ShardID, p.NodeID)

		// Pbft stage add 1. It means that this round of pbft goes into the next stage, i.e., Prepare stage.

		p.pbftStage.CompareAndSwap(1, 2)
	}
}

// Handle prepare messages here.
// If you want to do more operations in the prepare stage, you can implement the interface "ExtraOpInConsensus",
// and call the function: **ExtraOpInConsensus.HandleinPrepare**
func (p *PbftConsensusNode) handlePrepare(content []byte) {

	// decode the message
	pmsg := new(message.Prepare)
	err := json.Unmarshal(content, pmsg)
	if err != nil {
		p.pl.Plog.Panic(err)
	}
	p.pl.Plog.Printf("S%dN%d : received the Prepare , message SeqID %v ,new SeqID %v \n", p.ShardID, p.NodeID, pmsg.SeqID, p.sequenceID)
	curView := p.view.Load()
	p.pbftLock.Lock()
	defer p.pbftLock.Unlock()
	for p.pbftStage.Load() < 2 && pmsg.SeqID >= p.sequenceID && p.view.Load() == curView {
		p.pl.Plog.Printf("error seqID %v, sequenceID %v, view %v, curView %v, stage %v \n", pmsg.SeqID, p.sequenceID, p.view.Load(), curView, p.pbftStage.Load())
		p.conditionalVarpbftLock.Wait()
	}
	defer p.conditionalVarpbftLock.Broadcast()

	// if this message is out of date, return.
	if pmsg.SeqID < p.sequenceID || p.view.Load() != curView {
		return
	}

	if _, ok := p.requestPool[string(pmsg.Digest)]; !ok {
		p.pl.Plog.Printf("S%dN%d : doesn't have the digest in the requst pool, refuse to commit\n", p.ShardID, p.NodeID)
	} else if p.sequenceID < pmsg.SeqID {
		p.pl.Plog.Printf("S%dN%d : inconsistent sequence ID, refuse to commit\n", p.ShardID, p.NodeID)
	} else {
		// if needed more operations, implement interfaces
		p.ihm.HandleinPrepare(pmsg)

		p.set2DMap(true, string(pmsg.Digest), pmsg.SenderNode)
		cnt := len(p.cntPrepareConfirm[string(pmsg.Digest)])

		// if the node has received 2f messages (itself included), and it haven't committed, then it commit
		p.lock.Lock()
		defer p.lock.Unlock()
		if uint64(cnt) >= 2*p.malicious_nums+1 && !p.isCommitBordcast[string(pmsg.Digest)] {
			p.pl.Plog.Printf("S%dN%d : is going to commit\n", p.ShardID, p.NodeID)
			// generate commit and broadcast
			c := message.Commit{
				Digest:     pmsg.Digest,
				SeqID:      pmsg.SeqID,
				SenderNode: p.RunningNode,
			}
			commitByte, err := json.Marshal(c)
			if err != nil {
				log.Panic()
			}
			msg_send := message.MergeMessage(message.CCommit, commitByte)
			networks.Broadcast(p.RunningNode.IPaddr, p.getNeighborNodes(), msg_send)
			networks.TcpDial(msg_send, p.RunningNode.IPaddr)
			p.isCommitBordcast[string(pmsg.Digest)] = true
			p.pl.Plog.Printf("S%dN%d : commit is broadcast\n", p.ShardID, p.NodeID)

			p.pbftStage.CompareAndSwap(2, 3)
		}
	}
}

// Handle commit messages here.
// If you want to do more operations in the commit stage, you can implement the interface "ExtraOpInConsensus",
// and call the function: **ExtraOpInConsensus.HandleinCommit**
func (p *PbftConsensusNode) handleCommit(content []byte) {
	// decode the message
	cmsg := new(message.Commit)
	err := json.Unmarshal(content, cmsg)
	if err != nil {
		log.Panic(err)
	}

	curView := p.view.Load()
	p.pbftLock.Lock()
	defer p.pbftLock.Unlock()
	for p.pbftStage.Load() < 3 && cmsg.SeqID >= p.sequenceID && p.view.Load() == curView {
		p.pl.Plog.Printf("error seqID %v, sequenceID %v, view %v, curView %v, stage %v \n", cmsg.SeqID, p.sequenceID, p.view.Load(), curView, p.pbftStage.Load())
		p.conditionalVarpbftLock.Wait()
	}
	defer p.conditionalVarpbftLock.Broadcast()

	if cmsg.SeqID < p.sequenceID || p.view.Load() != curView {
		return
	}

	p.pl.Plog.Printf("S%dN%d received the Commit from ...%d\n", p.ShardID, p.NodeID, cmsg.SenderNode.NodeID)
	p.set2DMap(false, string(cmsg.Digest), cmsg.SenderNode)
	cnt := len(p.cntCommitConfirm[string(cmsg.Digest)])

	p.lock.Lock()
	defer p.lock.Unlock()

	if uint64(cnt) >= 2*p.malicious_nums+1 && !p.isReply[string(cmsg.Digest)] {
		p.pl.Plog.Printf("S%dN%d : has received 2f + 1 commits ... \n", p.ShardID, p.NodeID)
		// if this node is left behind, so it need to requst blocks
		if _, ok := p.requestPool[string(cmsg.Digest)]; !ok {
			p.isReply[string(cmsg.Digest)] = true
			p.askForLock.Lock()
			// request the block
			sn := &shard.Node{
				NodeID:  uint64(p.view.Load()),
				ShardID: p.ShardID,
				IPaddr:  p.ip_nodeTable[p.ShardID][uint64(p.view.Load())],
			}
			orequest := message.RequestOldMessage{
				SeqStartHeight: p.sequenceID + 1,
				SeqEndHeight:   cmsg.SeqID,
				ServerNode:     sn,
				SenderNode:     p.RunningNode,
			}
			bromyte, err := json.Marshal(orequest)
			if err != nil {
				log.Panic()
			}

			p.pl.Plog.Printf("S%dN%d : is now requesting message (seq %d to %d) ... \n", p.ShardID, p.NodeID, orequest.SeqStartHeight, orequest.SeqEndHeight)
			msg_send := message.MergeMessage(message.CRequestOldrequest, bromyte)
			networks.TcpDial(msg_send, orequest.ServerNode.IPaddr)
		} else {
			leadAddr := p.ip_nodeTable[p.ShardID][uint64(p.view.Load())]
			oldNode := shard.Node{
				NodeID:  p.NodeID,
				ShardID: p.ShardID,
				IPaddr:  p.RunningNode.IPaddr,
			}

			// implement interface
			p.ihm.HandleinCommit(cmsg)
			p.isReply[string(cmsg.Digest)] = true
			p.pl.Plog.Printf("S%dN%d: this round of pbft %d is end \n", p.ShardID, p.NodeID, p.sequenceID)
			p.sequenceID += 1

			// send reply of sequenceID
			replyMsg := message.Reply{
				SeqID:      p.sequenceID,
				SenderNode: oldNode,
				Result:     true,
			}

			replyByte, err := json.Marshal(replyMsg)
			if err != nil {
				log.Panic()
			}
			fmt.Println("send reply message to  ", leadAddr)
			msg_send := message.MergeMessage(message.CReply, replyByte)
			networks.TcpDial(msg_send, leadAddr)
			p.pbftStage.CompareAndSwap(3, 1)
		}

		p.lastCommitTime.Store(time.Now().UnixMilli())

		// if this node is a main node, then unlock the sequencelock
		if p.NodeID == uint64(p.view.Load()) {
			p.pl.Plog.Printf("S%dN%d get sequenceLock unlocked...\n", p.ShardID, p.NodeID)
			p.sequenceLock.Unlock()
		}
	}
}

// main node receive reply meg, update list, as producer
func (p *PbftConsensusNode) handleReply(content []byte) {
	reply := new(message.Reply)
	err := json.Unmarshal(content, reply)
	if err != nil {
		log.Panic(err)
	}

	p.ReplyListCond.L.Lock()
	defer p.ReplyListCond.L.Unlock()
	if p.ReplyList[reply.SeqID] == nil {
		p.ReplyList[reply.SeqID] = make([]string, 0)
	}
	p.pl.Plog.Printf("receive reply message from %v seqID %v, len %v\n", reply.SenderNode.IPaddr, reply.SeqID, len(p.ReplyList[reply.SeqID]))

	p.ReplyList[reply.SeqID] = append(p.ReplyList[reply.SeqID], reply.SenderNode.IPaddr)
	p.ReplyListCond.Broadcast()
}

// this func is only invoked by the main node,
// if the request is correct, the main node will send
// block back to the message sender.
// now this function can send both block and partition
func (p *PbftConsensusNode) handleRequestOldSeq(content []byte) {
	if uint64(p.view.Load()) != p.NodeID {
		return
	}

	rom := new(message.RequestOldMessage)
	err := json.Unmarshal(content, rom)
	if err != nil {
		log.Panic()
	}
	p.pl.Plog.Printf("S%dN%d : received the old message requst from ...", p.ShardID, p.NodeID)
	rom.SenderNode.PrintNode()

	oldR := make([]*message.Request, 0)
	for height := rom.SeqStartHeight; height <= rom.SeqEndHeight; height++ {
		if _, ok := p.height2Digest[height]; !ok {
			p.pl.Plog.Printf("S%dN%d : has no this digest to this height %d\n", p.ShardID, p.NodeID, height)
			break
		}
		if r, ok := p.requestPool[p.height2Digest[height]]; !ok {
			p.pl.Plog.Printf("S%dN%d : has no this message to this digest %d\n", p.ShardID, p.NodeID, height)
			break
		} else {
			oldR = append(oldR, r)
		}
	}
	p.pl.Plog.Printf("S%dN%d : has generated the message to be sent\n", p.ShardID, p.NodeID)

	p.ihm.HandleReqestforOldSeq(rom)

	// send the block back
	sb := message.SendOldMessage{
		SeqStartHeight: rom.SeqStartHeight,
		SeqEndHeight:   rom.SeqEndHeight,
		OldRequest:     oldR,
		SenderNode:     p.RunningNode,
	}
	sbByte, err := json.Marshal(sb)
	if err != nil {
		log.Panic()
	}
	msg_send := message.MergeMessage(message.CSendOldrequest, sbByte)
	networks.TcpDial(msg_send, rom.SenderNode.IPaddr)
	p.pl.Plog.Printf("S%dN%d : send blocks\n", p.ShardID, p.NodeID)
}

// node requst blocks and receive blocks from the main node
func (p *PbftConsensusNode) handleSendOldSeq(content []byte) {
	som := new(message.SendOldMessage)
	err := json.Unmarshal(content, som)
	if err != nil {
		log.Panic()
	}
	p.pl.Plog.Printf("S%dN%d : has received the SendOldMessage message\n", p.ShardID, p.NodeID)

	// implement interface for new consensus
	p.ihm.HandleforSequentialRequest(som)
	beginSeq := som.SeqStartHeight
	for idx, r := range som.OldRequest {
		p.requestPool[string(getDigest(r))] = r
		p.height2Digest[uint64(idx)+beginSeq] = string(getDigest(r))
		p.isReply[string(getDigest(r))] = true
		p.pl.Plog.Printf("this round of pbft %d is end \n", uint64(idx)+beginSeq)
	}
	p.sequenceID = som.SeqEndHeight + 1
	if rDigest, ok1 := p.height2Digest[p.sequenceID]; ok1 {
		if r, ok2 := p.requestPool[rDigest]; ok2 {
			ppmsg := &message.PrePrepare{
				RequestMsg: r,
				SeqID:      p.sequenceID,
				Digest:     getDigest(r),
			}
			flag := false
			flag = p.ihm.HandleinPrePrepare(ppmsg)
			if flag {
				pre := message.Prepare{
					Digest:     ppmsg.Digest,
					SeqID:      ppmsg.SeqID,
					SenderNode: p.RunningNode,
				}
				prepareByte, err := json.Marshal(pre)
				if err != nil {
					log.Panic()
				}
				// broadcast
				msg_send := message.MergeMessage(message.CPrepare, prepareByte)
				networks.Broadcast(p.RunningNode.IPaddr, p.getNeighborNodes(), msg_send)
				p.pl.Plog.Printf("S%dN%d : has broadcast the prepare message \n", p.ShardID, p.NodeID)
			}
		}
	}

	p.askForLock.Unlock()
}
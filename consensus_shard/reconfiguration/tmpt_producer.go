package reconfiguration

import (
	"blockEmulator/message"
	"blockEmulator/networks"
	"encoding/json"
	"time"
)

// handle tmpt
func (r *Reconfiguration) handleTmptSync(content []byte) {
	r.logger.Printf("Reconfiguration.handleTmptSync begin \n")
	tmptSync := &TMptSync{}
	err := json.Unmarshal(content, tmptSync)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.handleTmptSync error: %v \n", err)
	}
	accounts, k, v := r.CurChain.GetTmptData(tmptSync.Epoch - 1)
	r.logger.Printf("get tmpt data, accounts %v, data len %v \n", len(accounts), len(k))
	tmpSyncData := &TMptSyncData{
		FromShard: r.RunningNode.ShardID,
		Epoch:     tmptSync.Epoch,
		Accounts:  accounts,
		Keys:      k,
		Values:    v,
		CurrBlock: r.CurChain.CurrentBlock.Encode(),
	}
	r.accountPool = make(map[string]int)
	marshal, err := json.Marshal(tmpSyncData)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.handleTmptSync error: %v \n", err)
	}
	msgSend := message.MergeMessage(CTmptSyncData, marshal)
	networks.TcpDial(msgSend, tmptSync.FromAddress)
}

func (r *Reconfiguration) handleTMPTSpecificSync(content []byte) {
	r.logger.Printf("handleSpecificSync begin \n")
	specificSync := &SpecificSync{}
	err := json.Unmarshal(content, specificSync)
	if err != nil {
		r.logger.Fatalf("handleSpecificSync error: %v \n", err)
	}
	r.logger.Printf("receive %v accounts need to get. \n", len(specificSync.Accounts))
	for specificSync.BlockHeight > int(r.CurChain.CurrentBlock.Header.Number) {
		r.logger.Printf("handleSpecificSync error: block height not match, %v, %v \n", specificSync.BlockHeight, r.CurChain.CurrentBlock.Header.Number)
		time.Sleep(500 * time.Millisecond)
	}
	k, v := r.CurChain.GetSpecificData(specificSync.Accounts, specificSync.RootHash)
	r.logger.Printf("get specific data, data len %v \n", len(k))
	syncSpecificData := &SyncSpecificData{
		FromShard: r.RunningNode.ShardID,
		Epoch:     specificSync.Epoch,
		Round:     specificSync.Round,
		Keys:      k,
		Values:    v,
	}
	marshal, err := json.Marshal(syncSpecificData)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.handleSpecificSync error: %v \n", err)
	}
	msgSend := message.MergeMessage(CTMPTSpecificSyncData, marshal)
	networks.TcpDial(msgSend, specificSync.FromAddress)
}

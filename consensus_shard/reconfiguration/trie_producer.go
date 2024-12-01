package reconfiguration

import (
	"blockEmulator/message"
	"blockEmulator/networks"
	"encoding/json"
)

// handle tmpt
func (r *Reconfiguration) handleTireSync(content []byte) {
	r.logger.Printf("Reconfiguration.handleTmptSync begin \n")
	tmptSync := &TMptSync{}
	err := json.Unmarshal(content, tmptSync)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.handleTmptSync error: %v \n", err)
	}
	accounts, k, v := r.CurChain.GetTrieData()
	accounts = make([]string, 0)
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

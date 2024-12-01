package reconfiguration

import (
	"blockEmulator/chain"
	"blockEmulator/message"
	"blockEmulator/networks"
	"encoding/json"
)

func (r *Reconfiguration) handleFullSync(content []byte) {
	fullSync := &FullSync{}
	err := networks.Decode(content, fullSync)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.handleFullSync error: %v \n", err)
	}
	r.logger.Printf("%v Receive sync request, begin get full sync data to %v \n", r.RunningNode.IPaddr, fullSync)
	syncData := r.GetAllDiskData(r.CurChain, fullSync.Epoch)
	r.logger.Printf("Get all Data, ready to send to %v len %v \n", fullSync.FromAddress, len(syncData.Keys))
	marshal, err := json.Marshal(syncData)
	if err != nil {
		r.logger.Fatalf("Reconfiguration.handleFullSync error: %v \n", err)
	}
	msgSend := message.MergeMessage(CFullSyncData, marshal)
	networks.TcpDial(msgSend, fullSync.FromAddress)
}

func (r *Reconfiguration) GetAllDiskData(chain *chain.BlockChain, epoch uint64) *SyncData {
	iterator := r.db.NewIterator(nil, nil)
	keys := make([][]byte, 0)
	values := make([][]byte, 0)
	n := 0
	for iterator.Next() {
		n++
		temp := make([]byte, len(iterator.Key()))
		copy(temp, iterator.Key())
		keys = append(keys, temp)

		temp = make([]byte, len(iterator.Value()))
		copy(temp, iterator.Value())
		values = append(values, temp)
	}
	r.logger.Printf("Reconfiguration.getAllDiskData len %v \n", len(keys))
	return &SyncData{
		FromShard: r.RunningNode.ShardID,
		Epoch:     epoch,
		Keys:      keys,
		Values:    values,
		CurrBlock: chain.CurrentBlock.Encode(),
	}
}

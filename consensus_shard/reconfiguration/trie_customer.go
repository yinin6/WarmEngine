package reconfiguration

import (
	"blockEmulator/message"
	"blockEmulator/networks"
	"encoding/json"
	"time"
)

// the implementation of the tMPT
func (r *Reconfiguration) TrieSync(destShard uint64) {
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
	msgSend := message.MergeMessage(CTrieSyncData, marshal)
	networks.TcpDial(msgSend, target)
}

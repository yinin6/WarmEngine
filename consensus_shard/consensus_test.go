package consensus_shard

import (
	"blockEmulator/build"
	"blockEmulator/consensus_shard/pbft_all"
	"blockEmulator/params"
	"testing"
)

func TestNewPbftNode(t *testing.T) {
	var (
		sid = uint64(1)
		nid = uint64(1)
		snm = uint64(2)
		nnm = uint64(4)
	)

	worker := pbft_all.NewPbftNode(sid, nid, build.GetChainConfig(nid, nnm, sid, snm), params.CommitteeMethod[3])
	worker.RunningNode.PrintNode()
	worker.GetIPTable()

	worker.GetIPTable()
	worker.RunningNode.PrintNode()

}

package consensus_shard

import (
	"blockEmulator/build"
	"blockEmulator/consensus_shard/pbft_all"
	"blockEmulator/core"
	"blockEmulator/params"
	"fmt"
	"math/big"
	"testing"
)

// test get sync data
func TestGetSyncData(t *testing.T) {
	var (
		sid  = uint64(0)
		nid  = uint64(0)
		nid1 = uint64(1)
		snm  = uint64(2)
		nnm  = uint64(4)
	)

	worker := pbft_all.NewPbftNode(sid, nid, build.GetChainConfig(nid, nnm, sid, snm), params.CommitteeMethod[3])
	worker.RunningNode.PrintNode()
	worker.GetIPTable()

	chain := worker.CurChain

	accounts := []string{"000000000001", "00000000002", "00000000003", "00000000004", "00000000005", "00000000006"}
	as := make([]*core.AccountState, 0)
	for idx := range accounts {
		as = append(as, &core.AccountState{
			AcAddress: accounts[idx],
			Balance:   big.NewInt(int64(idx)),
		})
	}

	chain.AddAccounts(accounts, as, 0)
	chain.PrintBlockChain()
	astates := worker.CurChain.FetchAccounts(accounts)
	for _, state := range astates {
		fmt.Println(state.AcAddress, state.Balance)
	}

	syncData := worker.Reconfig.GetAllDiskData(chain)
	fmt.Println("syncData len:", len(syncData.Keys))

	worker1 := pbft_all.NewPbftNode(sid, nid1, build.GetChainConfig(nid1, nnm, sid, snm), params.CommitteeMethod[3])
	worker1.RunningNode.PrintNode()
	worker1.Reconfig.SyncDataUpdateBlockChain(worker1.CurChain, syncData)
	worker1.CurChain.PrintBlockChain()

	astates = worker1.CurChain.FetchAccounts(accounts)
	for _, state := range astates {
		fmt.Println(state.AcAddress, state.Balance)
	}

	newAccounts := []string{"000000000007", "00000000008"}
	newAS := make([]*core.AccountState, 0)
	for idx := range newAccounts {
		newAS = append(newAS, &core.AccountState{
			AcAddress: newAccounts[idx],
			Balance:   big.NewInt(int64(idx)),
		})
	}

	worker1.CurChain.AddAccounts(newAccounts, newAS, 0)

	astates = worker1.CurChain.FetchAccounts(newAccounts)
	for _, state := range astates {
		fmt.Println(state.AcAddress, state.Balance)
	}

}

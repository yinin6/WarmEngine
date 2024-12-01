package chain

import "sync"

var (
	Epoch2Accounts     = make(map[uint64]map[string]int)
	Epoch2AccountsLock sync.RWMutex
)

func (b *BlockChain) TmptRecordAccount(epochId uint64, a string) {
	Epoch2AccountsLock.Lock()
	defer Epoch2AccountsLock.Unlock()
	if _, ok := Epoch2Accounts[epochId]; !ok {
		Epoch2Accounts[epochId] = make(map[string]int)
	}
	Epoch2Accounts[epochId][a] = 1
}

func (b *BlockChain) TmptGetAccountPoolSize(epochId uint64) int {
	Epoch2AccountsLock.RLock()
	defer Epoch2AccountsLock.RUnlock()
	if _, ok := Epoch2Accounts[epochId]; !ok {
		return 0
	}
	return len(Epoch2Accounts[epochId])
}

func (b *BlockChain) TmptGetEpochAccount(epochID uint64) []string {
	Epoch2AccountsLock.RLock()
	defer Epoch2AccountsLock.RUnlock()
	if _, ok := Epoch2Accounts[epochID]; !ok {
		return nil
	}
	accounts := make([]string, 0)
	for k := range Epoch2Accounts[epochID] {
		accounts = append(accounts, k)
	}
	return accounts
}

func (b *BlockChain) TmptHaveAccount(epochID uint64, address string) bool {
	Epoch2AccountsLock.RLock()
	defer Epoch2AccountsLock.RUnlock()
	if _, ok := Epoch2Accounts[epochID]; !ok {
		return false
	}
	if _, ok := Epoch2Accounts[epochID][address]; ok {
		return true
	}
	return false
}

func (b *BlockChain) TmptAppendAccounts(epochID uint64, accounts []string) {
	Epoch2AccountsLock.Lock()
	defer Epoch2AccountsLock.Unlock()
	if _, ok := Epoch2Accounts[epochID]; !ok {
		Epoch2Accounts[epochID] = make(map[string]int)
	}
	for _, account := range accounts {
		Epoch2Accounts[epochID][account] = 1
	}
}

func (b *BlockChain) TmptAccountSize(epochID uint64) int {
	Epoch2AccountsLock.Lock()
	defer Epoch2AccountsLock.Unlock()
	if _, ok := Epoch2Accounts[epochID]; !ok {
		return 0
	}
	return len(Epoch2Accounts[epochID])
}

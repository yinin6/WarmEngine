package chain

import (
	"blockEmulator/trie"
	"github.com/ethereum/go-ethereum/common"
	"sync"
)

var (
	syncN          = 0
	Epoch          = -1
	Round          = -1
	addressMapLock sync.Mutex
	addr2Hit       = make(map[string]map[string]bool)
	AccountCache   = make(map[EpochRoundPair][]string)
	AccountVisited = make(map[string]bool)
)

type EpochRoundPair struct {
	Epoch int
	Round int
}

func (b *BlockChain) PreSyncList(first bool, from string, epoch, round int) []string {
	addressMapLock.Lock()
	defer addressMapLock.Unlock()
	p := EpochRoundPair{Epoch: epoch, Round: round}

	if first {
		addr2Hit[from] = make(map[string]bool)
		AccountVisited = make(map[string]bool)
	}

	if _, ok := AccountCache[p]; ok {
		return AccountCache[p]
	} else {
		account := b.GetNext2ActiveList(from, epoch, round)
		//account := b.GetNextActiveList(first)
		AccountCache[p] = account
	}
	return AccountCache[p]
}

// 从当前交易池中获取下一个活跃账户，返回账户列表，但是不负责进行去重
func (b *BlockChain) GetNext2ActiveList(from string, epoch, round int) (account []string) {
	account = make([]string, 0)
	begin := 0
	n := 0

	for i := begin; i < begin+int(b.ChainConfig.BlockSize)*2 && i < len(b.Txpool.TxQueue); i++ {
		n++
		tx := b.Txpool.TxQueue[i]
		b.Txpool.TxQueue[i].VisitCnt++
		b.Txpool.TxQueue[i].Epoch = epoch
		b.Txpool.TxQueue[i].Round = round
		ssid := b.Get_PartitionMap(tx.Sender)
		rsid := b.Get_PartitionMap(tx.Recipient)

		if ssid == b.ChainConfig.ShardID && !addr2Hit[from][tx.Sender] {
			addr2Hit[from][tx.Sender] = true
			account = append(account, tx.Sender)
		}
		if rsid == b.ChainConfig.ShardID && !addr2Hit[from][tx.Recipient] {
			addr2Hit[from][tx.Recipient] = true
			account = append(account, tx.Recipient)
		}
	}

	b.chainLog.Printf("travel %v txs, get %v accounts \n", n, len(account))
	return account
}

// 从当前交易池中获取下一个活跃账户，返回账户列表，但是不负责进行去重
func (b *BlockChain) GetNextActiveList(isFirst bool) (account []string) {
	b.Txpool.GetLocked()
	defer b.Txpool.GetUnlocked()
	account = make([]string, 0)
	num := int(b.ChainConfig.BlockSize)
	if isFirst {
		num = int(b.ChainConfig.BlockSize)
	}
	n := 0
	for i := range b.Txpool.TxQueue {
		n++
		tx := b.Txpool.TxQueue[i]
		ssid := b.Get_PartitionMap(tx.Sender)
		rsid := b.Get_PartitionMap(tx.Recipient)

		if ssid == b.ChainConfig.ShardID && !AccountVisited[tx.Sender] {
			AccountVisited[tx.Sender] = true
			account = append(account, tx.Sender)
		}

		if rsid == b.ChainConfig.ShardID && !AccountVisited[tx.Recipient] {
			AccountVisited[tx.Recipient] = true
			account = append(account, tx.Recipient)
		}
		if len(account) >= num {
			break
		}

	}
	b.chainLog.Printf("travel %v txs, get %v accounts, totoal account of %v  \n", n, len(account))
	return account
}

func (b *BlockChain) FilterNilAccount(accounts []string) []string {
	ans := make([]string, 0)
	st, err := trie.New(trie.TrieID(common.BytesToHash(b.CurrentBlock.Header.StateRoot)), b.triedb)
	if err != nil {
		return accounts
	}

	for i := 0; i < len(accounts); i++ {
		acc := accounts[i]
		_, err := st.Get([]byte(acc))
		if err != nil {
			ans = append(ans, acc)
		}
	}
	return ans
}

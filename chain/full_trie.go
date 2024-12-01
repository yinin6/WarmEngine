package chain

import (
	"blockEmulator/trie"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"log"
)

var (
	TireAccounts = make(map[string]struct{})
)

func (b *BlockChain) RecordTrie(a string) {
	TireAccounts[a] = struct{}{}
}

func (b *BlockChain) GetTrieData() (accounts []string, k, v [][]byte) {
	tmptLock.Lock()
	defer tmptLock.Unlock()
	st, err := trie.New(trie.TrieID(common.BytesToHash(b.CurrentBlock.Header.StateRoot)), b.triedb)
	if err != nil {
		log.Panic(err)
	}
	acc := make([]string, 0)
	proof := memorydb.New()
	for key, _ := range TireAccounts {
		acc = append(acc, key)
		err := st.Prove([]byte(key), 0, proof)
		if err != nil {
			log.Fatal(err)
		}
	}
	iterator := proof.NewIterator(nil, nil)
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
	b.chainLog.Printf("get tmpt data, accounts %v, data len %v \n", len(acc), n)
	return acc, keys, values
}

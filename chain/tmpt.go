package chain

import (
	"blockEmulator/params"
	"blockEmulator/trie"
	"crypto/sha256"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"log"
	"slices"
	"strings"
	"sync"
)

var (
	AccountPool     = make(map[string]int)
	tmptLock        sync.Mutex
	accountPoolLock sync.RWMutex
)

func (b *BlockChain) GetTmptData(EpochID uint64) (accounts []string, k, v [][]byte) {
	tmptLock.Lock()
	defer tmptLock.Unlock()
	st, err := trie.New(trie.TrieID(common.BytesToHash(b.CurrentBlock.Header.StateRoot)), b.triedb)
	if err != nil {
		log.Panic(err)
	}
	acc := make([]string, 0)
	proof := memorydb.New()
	for key, _ := range Epoch2Accounts[EpochID] {
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
	return acc, keys, values
}

func (b *BlockChain) GetSpecificData(acc []string, root []byte) (k, v [][]byte) {
	st, err := trie.New(trie.TrieID(common.BytesToHash(root)), b.triedb)
	if err != nil {
		log.Panic(err)
	}
	proof := memorydb.New()
	for _, key := range acc {
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
	return keys, values
}

type chunkPair struct {
	idx string
	key []byte
	val []byte
}

func (b *BlockChain) GetSpecificChunk(acc []string, root []byte, num, idx int) (k, v [][]byte) {
	st, err := trie.New(trie.TrieID(common.BytesToHash(root)), b.triedb)
	if err != nil {
		log.Panic(err)
	}
	proof := memorydb.New()
	for _, key := range acc {
		err := st.Prove([]byte(key), 0, proof)
		if err != nil {
			log.Fatal(err)
		}
	}
	iterator := proof.NewIterator(nil, nil)
	keys := make([][]byte, 0)
	values := make([][]byte, 0)
	pairs := []chunkPair{}
	n := 0
	for iterator.Next() {
		n++
		p := chunkPair{}
		p.idx = string(iterator.Key())
		p.key = append([]byte{}, iterator.Key()...)
		p.val = append([]byte{}, iterator.Value()...)
		pairs = append(pairs, p)
	}
	slices.SortFunc(pairs, func(i, j chunkPair) int {
		return strings.Compare(i.idx, j.idx)
	})

	eachNum := n / num
	start := idx * eachNum
	end := (idx + 1) * eachNum
	if idx == num-1 {
		end = n
	}
	for i := start; i < end; i++ {
		keys = append(keys, pairs[i].key)
		values = append(values, pairs[i].val)
	}

	return keys, values
}

type Chunk struct {
	ChunkId     int
	ChunkNum    int
	Root        []byte
	Keys        [][]byte
	Values      [][]byte
	ErasureData []byte
	RawLen      int
}

var (
	ChunkPool    = make(map[ChunkIndex][]Chunk)
	chunkHitCont = make(map[ChunkIndex]int)
	trieLock     sync.Mutex
)

func hashStringSlice(slice []string) string {
	// 将所有字符串拼接成一个字符串
	joined := strings.Join(slice, ",")
	// 使用 SHA-256 生成哈希
	hash := sha256.Sum256([]byte(joined))
	return fmt.Sprintf("%x", hash)
}

type ChunkIndex struct {
	IdxChunk int
	Epoch    int
	Round    int
	ZoneID   int
}

func (b *BlockChain) GetSpecificSingleChunk(acc []string, root []byte, num, idx int, tireK ChunkIndex) Chunk {
	trieLock.Lock()

	defer releaseChunkPool(tireK)
	defer trieLock.Unlock()
	chunkHitCont[tireK]++
	if _, ok := ChunkPool[tireK]; ok {
		return ChunkPool[tireK][idx]
	}
	ChunkPool[tireK] = b.GetErasureChunks(acc, root, num)
	return ChunkPool[tireK][idx]
}

func releaseChunkPool(key ChunkIndex) {
	if chunkHitCont[key] >= params.ZoneSize {
		delete(ChunkPool, key)
		delete(chunkHitCont, key)
	}
}

func (b *BlockChain) GetSpecificChunksSize(acc []string, root []byte) (int, int, int) {
	st, err := trie.New(trie.TrieID(common.BytesToHash(root)), b.triedb)
	if err != nil {
		log.Panic(err)
	}
	proof := memorydb.New()
	for _, key := range acc {
		err := st.Prove([]byte(key), 0, proof)
		if err != nil {
			log.Fatal(err)
		}
	}
	iterator := proof.NewIterator(nil, nil)
	pairs := []chunkPair{}
	stateDataSize := 0
	stateKeySize, stateValueSize := 0, 0
	n := 0
	for iterator.Next() {
		n++
		p := chunkPair{}
		p.idx = string(iterator.Key())
		p.key = append([]byte{}, iterator.Key()...)
		p.val = append([]byte{}, iterator.Value()...)
		pairs = append(pairs, p)

		stateKeySize += len(p.key)
		stateValueSize += len(p.val)
	}
	stateDataSize = stateKeySize + stateValueSize

	return stateKeySize, stateValueSize, stateDataSize
}

func (b *BlockChain) GetSpecificChunks(acc []string, root []byte, num int) []Chunk {
	st, err := trie.New(trie.TrieID(common.BytesToHash(root)), b.triedb)
	if err != nil {
		log.Panic(err)
	}
	proof := memorydb.New()
	for _, key := range acc {
		err := st.Prove([]byte(key), 0, proof)
		if err != nil {
			log.Fatal(err)
		}
	}
	iterator := proof.NewIterator(nil, nil)
	pairs := []chunkPair{}
	n := 0
	for iterator.Next() {
		n++
		p := chunkPair{}
		p.idx = string(iterator.Key())
		p.key = append([]byte{}, iterator.Key()...)
		p.val = append([]byte{}, iterator.Value()...)
		pairs = append(pairs, p)
	}
	slices.SortFunc(pairs, func(i, j chunkPair) int {
		return strings.Compare(i.idx, j.idx)
	})

	chunks := []Chunk{}

	for idx := 0; idx < num; idx++ {
		chunk := Chunk{}
		keys := make([][]byte, 0)
		values := make([][]byte, 0)
		eachNum := n / num
		start := idx * eachNum
		end := (idx + 1) * eachNum
		if idx == num-1 {
			end = n
		}
		for i := start; i < end; i++ {
			keys = append(keys, pairs[i].key)
			values = append(values, pairs[i].val)
		}
		chunk.ChunkId = idx
		chunk.ChunkNum = num
		chunk.Keys = keys
		chunk.Values = values
		chunk.Root = root
		chunks = append(chunks, chunk)

	}

	return chunks
}

func (b *BlockChain) ClearAndUpdateAccountPool(acc []string) {
	AccountPool = make(map[string]int)
	for _, v := range acc {
		AccountPool[v] = 1
	}
}

func (b *BlockChain) UpdateAccountPool(acc []string) {
	accountPoolLock.Lock()
	defer accountPoolLock.Unlock()
	for _, v := range acc {
		AccountPool[v] = 1
	}
}

func (b *BlockChain) RecordAccount(a string) {
	AccountPool[a] = 1
}

func (b *BlockChain) GetAccountPoolSize() int {
	return len(AccountPool)
}

func (b *BlockChain) HaveAccount(s string) bool {
	accountPoolLock.RLock()
	defer accountPoolLock.RUnlock()
	if _, ok := AccountPool[s]; ok {
		return true
	}
	return false
}

func (b *BlockChain) GetAccountSize() int {
	return len(AccountPool)
}

func (b *BlockChain) ClearAccountPool() {
	AccountPool = make(map[string]int)
}

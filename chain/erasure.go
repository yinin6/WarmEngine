package chain

import (
	"blockEmulator/params"
	"blockEmulator/trie"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/klauspost/reedsolomon"
	"io"
	"log"
)

func GetErasureData(inputData []byte, totalNum int) [][]byte {
	enc, err := reedsolomon.New(totalNum-1, 1)
	if err != nil {
		fmt.Println("Failed to create encoder:", err)
	}
	splitData, err := enc.Split(inputData)
	if err != nil {
		fmt.Println(err)
	}
	err = enc.Encode(splitData)
	if err != nil {
		fmt.Println("Failed to encode data:", err)
	}
	return splitData
}

func RecoverDataPair(splitData [][]byte, rawDataSize, num int) DataPair {
	enc, err := reedsolomon.New(num-1, 1)
	if err != nil {
		fmt.Println("Failed to create encoder:", err)
	}

	// 解码并恢复数据
	err = enc.Reconstruct(splitData)
	if err != nil {
		log.Fatalf("解码失败: %v", err)
	}
	fmt.Println("解码成功，数据已恢复")
	// 验证数据是否正确
	ok, err := enc.Verify(splitData)
	if err != nil {
		log.Fatalf("验证失败: %v", err)
	}
	if ok {
		fmt.Println("验证成功，数据一致")
	} else {
		fmt.Println("验证失败，数据不一致")
	}
	// 恢复原始数据
	var buf bytes.Buffer
	writer := io.Writer(&buf)

	err = enc.Join(writer, splitData, rawDataSize)

	var decodedData DataPair
	decoder := gob.NewDecoder(&buf)
	if err := decoder.Decode(&decodedData); err != nil {
		fmt.Println("解码错误:", err)
	}
	return decodedData
}

type DataPair struct {
	Keys   [][]byte
	Values [][]byte
}

func (b *BlockChain) GetErasureChunks(acc []string, root []byte, num int) []Chunk {
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

	chunks := []Chunk{}

	rawData := DataPair{
		Keys:   keys,
		Values: values,
	}

	// 创建一个字节缓冲区用于存储编码数据
	var buffer bytes.Buffer

	// 创建 gob 编码器并将数据编码到 buffer 中
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(rawData); err != nil {
		log.Fatalf("编码错误:", err)
	}

	splitData := GetErasureData(buffer.Bytes(), num)

	for idx := 0; idx < num; idx++ {
		chunk := Chunk{}
		ks := make([][]byte, 0)
		vs := make([][]byte, 0)
		eachNum := n / num
		start := idx * eachNum
		end := (idx + 1) * eachNum
		if idx == num-1 {
			end = n
		}
		if params.UseErasure {
			chunk.RawLen = buffer.Len()
			chunk.ErasureData = splitData[idx]
		} else {
			for i := start; i < end; i++ {
				ks = append(keys, keys[i])
				vs = append(values, values[i])
			}
		}
		chunk.ChunkId = idx
		chunk.ChunkNum = num
		chunk.Keys = ks
		chunk.Values = vs
		chunk.Root = root

		chunks = append(chunks, chunk)
	}

	return chunks
}

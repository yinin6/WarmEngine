package reconfiguration

import (
	"blockEmulator/message"
	"time"
)

const (
	CShuffleTable message.MessageType = "shuffleTable"
	CShuffleReady message.MessageType = "cShuffleReady"
)

const (
	// full sync
	CFullSync     message.MessageType = "CFullSync"
	CFullSyncData message.MessageType = "CFullSyncData"
	CFullSyncOver message.MessageType = "CFullSyncOver"

	// tmpt
	CTmptSync     message.MessageType = "CTmptSync"
	CTmptSyncData message.MessageType = "CTmptSyncData"

	// trie
	CTrieSyncData message.MessageType = "CTrieSyncData"

	// tmpt specific sync
	CTMPTSpecificSync     message.MessageType = "CTMPTSpecificSync"
	CTMPTSpecificSyncData message.MessageType = "CTMPTSpecificSyncData"
)

const (
	ShuffleRequest message.RequestType = "Shuffle"
)

type ShuffleData struct {
	ShardID     uint64
	Epoch       uint64
	IpNodeTable map[uint64]map[uint64]string
}

type ShuffleReadyMessage struct {
	ShardID uint64
	Epoch   uint64
}

// FullSync is a message to request full sync data
type FullSync struct {
	FromAddress string
	Epoch       uint64
}

// SyncData is a message to send full sync data
type SyncData struct {
	FromShard uint64
	Epoch     uint64
	CurrBlock []byte
	Keys      [][]byte
	Values    [][]byte
}

type NotifyLeaderSyncOver struct {
	ShardID     uint64
	NodeID      uint64
	FromAddress string
	Epoch       uint64
}

// tmpt
// FullSync is a message to request full sync data
type TMptSync struct {
	FromAddress string
	Epoch       uint64
}

type TMptSyncData struct {
	FromShard uint64
	Epoch     uint64
	CurrBlock []byte
	Accounts  []string
	Keys      [][]byte
	Values    [][]byte
}

type SpecificSync struct {
	FromAddress string
	Epoch       uint64
	Round       int
	RootHash    []byte
	BlockHeight int
	Accounts    []string
	NChunk      int
	IdxChunk    int
}

type SyncSpecificData struct {
	FromShard uint64
	Epoch     uint64
	Round     int
	CurrBlock []byte
	NChunk    int
	IdxChunk  int
	Keys      [][]byte
	Values    [][]byte
	Numbers   []string
}

type ErasureChunk struct {
	SendTo               string
	Epoch                uint64
	Round                int
	NChunk               int
	IdxChunk             int
	ErasureId            int
	Keys                 [][]byte
	Values               [][]byte
	Data                 []byte
	Numbers              []string
	ErasureData          []byte
	RawLen               int
	ProductorReceiveTime time.Time
	ProductorSendTime    time.Time
}

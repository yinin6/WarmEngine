package reconfiguration

import (
	"blockEmulator/chain"
	"blockEmulator/consensus_shard/reconfiguration/util"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"blockEmulator/shard"
	"bufio"
	"fmt"
	"github.com/ethereum/go-ethereum/ethdb"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type Reconfiguration struct {
	db       ethdb.Database    // to save the mpt
	CurChain *chain.BlockChain // all node in the shard maintain the same blockchain

	// use for record epoch appear account
	accountPool         map[string]int
	preEpochAccountPool map[string]int

	ipNodeTableOfPBFT map[uint64]map[uint64]string // denote the ip of the specific node
	ipNodeTable       map[uint64]map[uint64]string // denote the ip of the specific node
	RunningNode       *shard.Node                  // the node information
	Address           string
	Epoch             uint64

	IsSelectedPreSync bool
	PreSyncRound      int

	// tcp control
	tcpLn       net.Listener
	tcpPoolLock sync.Mutex

	stopSignal     atomic.Bool // send stop signal
	ConfirmedRound atomic.Int64

	RawShardID uint64
	RawNodeID  uint64

	logger *log.Logger

	LevelDBLock sync.Mutex
}

func NewReconfiguration(db ethdb.Database, RunningNode *shard.Node, ipNodeTable map[uint64]map[uint64]string, chain *chain.BlockChain) *Reconfiguration {
	r := new(Reconfiguration)
	r.ipNodeTableOfPBFT = ipNodeTable
	r.LevelDBLock = sync.Mutex{}
	r.IsSelectedPreSync = false
	r.PreSyncRound = 0
	r.accountPool = make(map[string]int)
	r.preEpochAccountPool = make(map[string]int)
	r.ipNodeTable = util.UpdateIpTable(ipNodeTable)
	r.logger = NewLog(RunningNode.ShardID, RunningNode.NodeID)
	r.Address = r.ipNodeTable[RunningNode.ShardID][RunningNode.NodeID]
	r.logger.Printf("addr:%v,init ip table %v\n", r.Address, r.ipNodeTable)
	r.CurChain = chain
	r.Epoch = 0
	r.db = db
	r.RunningNode = &shard.Node{
		NodeID:  RunningNode.NodeID,
		ShardID: RunningNode.ShardID,
		IPaddr:  r.Address,
	}
	r.RawShardID = RunningNode.ShardID
	r.RawNodeID = RunningNode.NodeID
	r.stopSignal.Store(false)

	if params.SyncMod != 2 {
		column1 := []string{"epoch", "shardID", "nodeID", "beginTime", "receiveTime", "overTime", "costTime", "stateDataSize", "blockSize"}
		name1 := "S" + strconv.Itoa(int(r.RawShardID)) + "N" + strconv.Itoa(int(r.RawNodeID))
		util.MakeCsv(column1, name1)
	}

	for i := 1; i <= params.MigrateNodeNum; i++ {
		if uint64(i) == RunningNode.NodeID {
			column2 := []string{"epoch", "round", "shardID", "nodeID", "beginTime",
				"receiveTime", "overTime", "costTime", "stateDataSize", "blockSize",
				"repeat", "stateKeySize", "stateValueSize", "accountNumber", "ProductorReceiveTime", "ProductorSendTime", "LocalReceiveTime"}
			name2 := "S" + strconv.Itoa(int(r.RawShardID)) + "N" + strconv.Itoa(int(r.RawNodeID)) + "specific"
			util.MakeCsv(column2, name2)
		}

	}

	return r
}

func (r *Reconfiguration) checkValidIP(address string) bool {
	for _, v := range r.ipNodeTable {
		for _, vv := range v {
			white := strings.Split(vv, ":")[0]
			if white == address {
				return true
			}
		}
	}
	return false
}

// TcpListen A consensus node starts tcp-listen.
func (r *Reconfiguration) TcpListen() {
	ln, err := net.Listen("tcp", r.RunningNode.IPaddr)
	r.tcpLn = ln
	if err != nil {
		log.Panic(err)
	}
	for {
		conn, err := r.tcpLn.Accept()

		clientAddr := conn.RemoteAddr().String()
		remoteIp := strings.Split(clientAddr, ":")[0]
		if r.checkValidIP(remoteIp) == false {
			r.logger.Println("invalid ip address", clientAddr)
			continue
		} else {
			r.logger.Println("receive client address:", clientAddr)
		}

		if err != nil {
			return
		}
		go r.handleClientRequest(conn)
	}
}

func (r *Reconfiguration) handleClientRequest(con net.Conn) {
	defer con.Close()
	clientReader := bufio.NewReader(con)
	for {
		clientRequest, err := clientReader.ReadBytes(networks.Delimit)
		if r.stopSignal.Load() {
			return
		}
		switch err {
		case nil:
			r.tcpPoolLock.Lock()
			r.handleMessage(clientRequest)
			r.tcpPoolLock.Unlock()
		case io.EOF:
			log.Println("client closed the connection by terminating the process")
			return
		default:
			log.Printf("error: %v\n", err)
			return
		}
	}
}

// handle the raw message, send it to corresponded interfaces
func (r *Reconfiguration) handleMessage(msg []byte) {
	if len(msg) < 30 {
		r.logger.Printf("error msg len: %v\n", len(msg))
		return
	}
	msgType, content := message.SplitMessage(msg)
	r.logger.Printf("msgType: %v content len: %v\n", msgType, len(content))
	switch msgType {
	case CFullSync:
		go r.handleFullSync(content)
	case CFullSyncData:
		go r.handleFullSyncData(content)
		// tmpt
	case CTmptSync:
		go r.handleTmptSync(content)
	case CTmptSyncData:
		go r.handleTMptSyncData(content)
	case CTrieSyncData:
		go r.handleTireSync(content)
	case CTMPTSpecificSync:
		go r.handleTMPTSpecificSync(content)
	case CTMPTSpecificSyncData:
		go r.handleTMPTSpecificSyncData(content)

		//proposed
	case CPreSync:
		go r.handlePreSync(content)
	case CPreSyncData:
		go r.handlePreSyncData(content)
	case CSpecificSync:
		go r.handleSpecificSync(content)
	case CSpecificSyncData:
		go r.handleSpecificSyncData(content)
	case CErasureChunk:
		go r.handleErasureData(content)

	// handle the message from outside
	default:

	}
}

// ShuffleIpTable used by the leader to generate the new ip table
func (r *Reconfiguration) ShuffleIpTable(ipNodeTable map[uint64]map[uint64]string) *ShuffleData {

	r.RunningNode.PrintNode()
	migrateAddress := make([][]string, params.MigrateNodeNum+1)
	for i := 1; i <= params.MigrateNodeNum; i++ {
		address := make([]string, 0)
		for j := 0; j < params.ShardNum; j++ {
			address = append(address, ipNodeTable[uint64(j)][uint64(i)])
		}
		migrateAddress[i] = address
	}

	r.logger.Printf("move node: %v, list %v \n", params.MigrateNodeNum, migrateAddress)

	table := util.DeepCopyIpNodeTable(ipNodeTable)

	for i := 1; i <= params.MigrateNodeNum; i++ {
		address := migrateAddress[i]
		for j := 0; j < params.ShardNum; j++ {
			table[uint64(j)][uint64(i)] = address[(j+1)%len(address)]
		}
	}

	r.logger.Printf("shuffle result,now Epoch is %v, ip table: %v \n", r.Epoch, table)

	return &ShuffleData{
		ShardID:     r.RunningNode.ShardID,
		Epoch:       r.Epoch + 1,
		IpNodeTable: table,
	}

}

// ShuffleUpdateTable used by the follower to update the ip table
func (r *Reconfiguration) ShuffleUpdateTable(ipNodeTable map[uint64]map[uint64]string) {
	r.ipNodeTable = util.UpdateIpTable(ipNodeTable)
	r.ipNodeTableOfPBFT = ipNodeTable
	r.Epoch++
	flag := false
	for i, v := range r.ipNodeTable {
		for j, vv := range v {
			if vv == r.RunningNode.IPaddr {
				if r.RunningNode.ShardID != i {
					flag = true
				}
				r.RunningNode.ShardID = i
				r.RunningNode.NodeID = j
			}
		}
	}
	r.IsSelectedPreSync = flag
	r.PreSyncRound = 0
	r.RunningNode.PrintNode()
	r.logger.Printf("shuffle over,now Epoch is %v is selected %v \n", flag, r.Epoch)
}

func NewLog(sid, nid uint64) *log.Logger {
	pfx := fmt.Sprintf("Reconfig S%dN%d: ", sid, nid)
	writer1 := os.Stdout

	dirpath := params.LogWrite_path + "/config/S" + strconv.Itoa(int(sid))
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}
	writer2, err := os.OpenFile(dirpath+"/N"+strconv.Itoa(int(nid))+".log", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Panic(err)
	}
	pl := log.New(io.MultiWriter(writer1, writer2), pfx, log.Lshortfile|log.Ldate|log.Ltime)
	fmt.Println()

	return pl
}

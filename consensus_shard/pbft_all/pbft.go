// The pbft consensus process

package pbft_all

import (
	"blockEmulator/chain"
	"blockEmulator/consensus_shard/pbft_all/dataSupport"
	"blockEmulator/consensus_shard/pbft_all/pbft_log"
	"blockEmulator/consensus_shard/reconfiguration"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"blockEmulator/shard"
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
)

type PbftConsensusNode struct {
	// the local config about pbft
	RunningNode *shard.Node // the node information
	ShardID     uint64      // denote the ID of the shard (or pbft), only one pbft consensus in a shard
	NodeID      uint64      // denote the ID of the node in the pbft (shard)

	// the data structure for blockchain
	CurChain *chain.BlockChain // all node in the shard maintain the same blockchain
	db       ethdb.Database    // to save the mpt

	// the global config about pbft
	pbftChainConfig *params.ChainConfig          // the chain config in this pbft
	ip_nodeTable    map[uint64]map[uint64]string // denote the ip of the specific node
	node_nums       uint64                       // the number of nodes in this pfbt, denoted by N
	malicious_nums  uint64                       // f, 3f + 1 = N

	// view change
	view           atomic.Int32 // denote the view of this pbft, the main node can be inferred from this variant
	lastCommitTime atomic.Int64 // the time since last commit.
	viewChangeMap  map[ViewChangeData]map[uint64]bool
	newViewMap     map[ViewChangeData]map[uint64]bool

	// the control message and message checking utils in pbft
	sequenceID        uint64                          // the message sequence id of the pbft, begin from 1
	stopSignal        atomic.Bool                     // send stop signal
	pStop             chan uint64                     // channle for stopping consensus
	requestPool       map[string]*message.Request     // RequestHash to Request
	cntPrepareConfirm map[string]map[*shard.Node]bool // count the prepare confirm message, [messageHash][Node]bool
	cntCommitConfirm  map[string]map[*shard.Node]bool // count the commit confirm message, [messageHash][Node]bool
	isCommitBordcast  map[string]bool                 // denote whether the commit is broadcast
	isReply           map[string]bool                 // denote whether the message is reply
	height2Digest     map[uint64]string               // sequence (block height) -> request, fast read
	ReplyList         map[uint64][]string

	// pbft stage wait
	pbftStage              atomic.Int32 // 1->Preprepare, 2->Prepare, 3->Commit, 4->Done
	pbftLock               sync.Mutex
	conditionalVarpbftLock sync.Cond
	ReplyListCond          sync.Cond

	// locks about pbft
	sequenceLock sync.Mutex // the lock of sequence
	lock         sync.Mutex // lock the stage
	askForLock   sync.Mutex // lock for asking for a serise of requests

	// seqID of other Shards, to synchronize
	seqIDMap   map[uint64]uint64
	seqMapLock sync.Mutex

	//control reconfiguration
	reconfigStage    atomic.Int32
	Epoch            uint64
	Epoch2ReadyShard sync.Map
	OverSyncInShard  sync.Map // 确保分片内所有节点准备好开始共识

	// logger
	pl *pbft_log.PbftLog
	// tcp control
	tcpln       net.Listener
	tcpPoolLock sync.Mutex

	// to handle the message in the pbft
	ihm ExtraOpInConsensus

	// to handle the message outside of pbft
	ohm OpInterShards

	Reconfig *reconfiguration.Reconfiguration
}

// generate a pbft consensus for a node
func NewPbftNode(shardID, nodeID uint64, pcc *params.ChainConfig, messageHandleType string) *PbftConsensusNode {
	p := new(PbftConsensusNode)
	p.ip_nodeTable = params.IPmap_nodeTable
	p.Epoch = 0
	p.node_nums = pcc.Nodes_perShard
	p.ShardID = shardID
	p.NodeID = nodeID
	p.pbftChainConfig = pcc
	p.ReplyListCond = sync.Cond{L: &sync.Mutex{}}
	fp := params.DatabaseWrite_path + "mptDB/ldb/s" + strconv.FormatUint(shardID, 10) + "/n" + strconv.FormatUint(nodeID, 10)
	var err error
	p.db, err = rawdb.NewLevelDBDatabase(fp, 0, 1, "accountState", false)
	if err != nil {
		log.Panic(err)
	}
	p.CurChain, err = chain.NewBlockChain(pcc, p.db)
	if err != nil {
		log.Panic("cannot new a blockchain")
	}

	p.RunningNode = &shard.Node{
		NodeID:  nodeID,
		ShardID: shardID,
		IPaddr:  p.ip_nodeTable[shardID][nodeID],
	}

	p.stopSignal.Store(false)
	p.sequenceID = p.CurChain.CurrentBlock.Header.Number + 1
	p.pStop = make(chan uint64)
	p.requestPool = make(map[string]*message.Request)
	p.cntPrepareConfirm = make(map[string]map[*shard.Node]bool)
	p.cntCommitConfirm = make(map[string]map[*shard.Node]bool)
	p.isCommitBordcast = make(map[string]bool)
	p.isReply = make(map[string]bool)
	p.height2Digest = make(map[uint64]string)
	p.malicious_nums = (p.node_nums - 1) / 3
	p.ReplyList = make(map[uint64][]string)

	// init view & last commit time
	p.view.Store(0)
	p.lastCommitTime.Store(time.Now().Add(time.Second * 5).UnixMilli())
	p.viewChangeMap = make(map[ViewChangeData]map[uint64]bool)
	p.newViewMap = make(map[ViewChangeData]map[uint64]bool)

	p.seqIDMap = make(map[uint64]uint64)

	p.pl = pbft_log.NewPbftLog(shardID, nodeID)

	// choose how to handle the messages in pbft or beyond pbft
	switch string(messageHandleType) {
	case "CLPA_Broker":
		ncdm := dataSupport.NewCLPADataSupport()
		p.ihm = &CLPAPbftInsideExtraHandleMod_forBroker{
			pbftNode: p,
			cdm:      ncdm,
		}
		p.ohm = &CLPABrokerOutsideModule{
			pbftNode: p,
			cdm:      ncdm,
		}
	case "CLPA":
		ncdm := dataSupport.NewCLPADataSupport()
		p.ihm = &CLPAPbftInsideExtraHandleMod{
			pbftNode: p,
			cdm:      ncdm,
		}
		p.ohm = &CLPARelayOutsideModule{
			pbftNode: p,
			cdm:      ncdm,
		}
	case "Broker":
		p.ihm = &RawBrokerPbftExtraHandleMod{
			pbftNode: p,
		}
		p.ohm = &RawBrokerOutsideModule{
			pbftNode: p,
		}
	default:
		p.ihm = &RawRelayPbftExtraHandleMod{
			pbftNode: p,
		}
		p.ohm = &RawRelayOutsideModule{
			pbftNode: p,
		}
	}

	// set pbft stage now
	p.conditionalVarpbftLock = *sync.NewCond(&p.pbftLock)
	p.pbftStage.Store(1)

	p.Reconfig = reconfiguration.NewReconfiguration(p.db, p.RunningNode, p.ip_nodeTable, p.CurChain)

	return p
}

// handle the raw message, send it to corresponded interfaces
func (p *PbftConsensusNode) handleMessage(msg []byte) {
	if len(msg) < 30 {
		p.pl.Plog.Printf("error msg len: %v\n", len(msg))
		return
	}
	msgType, content := message.SplitMessage(msg)
	p.pl.Plog.Printf("msgType: %v content len: %v\n", msgType, len(content))
	switch msgType {
	// pbft inside message type
	case message.CPrePrepare:
		// use "go" to start a go routine to handle this message, so that a pre-arrival message will not be aborted.
		go p.handlePrePrepare(content)
	case message.CPrepare:
		// use "go" to start a go routine to handle this message, so that a pre-arrival message will not be aborted.
		go p.handlePrepare(content)
	case message.CCommit:
		// use "go" to start a go routine to handle this message, so that a pre-arrival message will not be aborted.
		go p.handleCommit(content)
	case message.CReply:
		p.handleReply(content)

	case message.ViewChangePropose:
		p.handleViewChangeMsg(content)
	case message.NewChange:
		p.handleNewViewMsg(content)

	case message.CRequestOldrequest:
		p.handleRequestOldSeq(content)
	case message.CSendOldrequest:
		p.handleSendOldSeq(content)
	case reconfiguration.CShuffleReady:
		p.handleShuffleReady(content)
	case reconfiguration.CFullSyncOver:
		p.handleFullSyncOver(content)

	case message.CStop:
		p.WaitToStop()

	// handle the message from outside
	default:
		go p.ohm.HandleMessageOutsidePBFT(msgType, content)
	}
}

func (p *PbftConsensusNode) handleClientRequest(con net.Conn) {
	defer con.Close()
	clientReader := bufio.NewReader(con)
	for {
		clientRequest, err := clientReader.ReadBytes(networks.Delimit)
		//if len(clientRequest) == 0 {
		//	continue
		//}
		//clientRequest = clientRequest[:len(clientRequest)-1]
		if p.stopSignal.Load() {
			return
		}
		switch err {
		case nil:
			p.tcpPoolLock.Lock()
			p.handleMessage(clientRequest)
			p.tcpPoolLock.Unlock()
		case io.EOF:
			log.Println("client closed the connection by terminating the process")
			return
		default:
			log.Printf("error: %v\n", err)
			return
		}
	}
}

func (p *PbftConsensusNode) checkValidIP(address string) bool {
	for _, v := range p.ip_nodeTable {
		for _, vv := range v {
			white := strings.Split(vv, ":")[0]
			if white == address {
				return true
			}
		}
	}
	return false

}

// A consensus node starts tcp-listen.
func (p *PbftConsensusNode) TcpListen() {
	fmt.Println("Pbft begins listening:", p.RunningNode.IPaddr)
	//listenAddr := "0.0.0.0:" + strings.Split(p.RunningNode.IPaddr, ":")[1]

	ln, err := net.Listen("tcp", p.RunningNode.IPaddr)
	p.tcpln = ln
	if err != nil {
		log.Panic(err)
	}
	for {
		conn, err := p.tcpln.Accept()

		clientAddr := conn.RemoteAddr().String()
		remoteIp := strings.Split(clientAddr, ":")[0]
		if p.checkValidIP(remoteIp) == false {
			p.pl.Plog.Println("invalid ip address", clientAddr)
			continue
		} else {
			p.pl.Plog.Println("receive client address:", clientAddr)
		}

		if err != nil {
			return
		}
		go p.handleClientRequest(conn)
	}
}

// When receiving a stop message, this node try to stop.
func (p *PbftConsensusNode) WaitToStop() {
	p.pl.Plog.Println("handling stop message")
	p.stopSignal.Store(true)
	networks.CloseAllConnInPool()
	p.tcpln.Close()
	p.closePbft()
	p.pl.Plog.Println("handled stop message in TCPListen Routine")
	p.pStop <- 1
}

// close the pbft
func (p *PbftConsensusNode) closePbft() {
	p.CurChain.CloseBlockChain()
}

func (p *PbftConsensusNode) Shuffle(table map[uint64]map[uint64]string) (bool, uint64) {
	p.pl.Plog.Println("shuffle ip table start")
	flag := false
	p.updateIpTable(table)
	for i, v := range p.ip_nodeTable {
		for j, vv := range v {
			if vv == p.RunningNode.IPaddr {
				if p.RunningNode.ShardID != i {
					flag = true
				}
				p.RunningNode.ShardID = i
				p.ShardID = i
				p.RunningNode.NodeID = j
				p.NodeID = j
			}
		}
	}
	p.Reconfig.ShuffleUpdateTable(p.ip_nodeTable)
	p.pl.Plog.Printf("shuffle over, result %v, \n ", flag)
	return flag, p.RunningNode.ShardID
}

func (p *PbftConsensusNode) GetIPTable() {
	fmt.Println(p.ip_nodeTable)
}

func (p *PbftConsensusNode) updateIpTable(table map[uint64]map[uint64]string) {
	for i, v := range table {
		for j, vv := range v {
			p.ip_nodeTable[i][j] = vv
		}
	}
}

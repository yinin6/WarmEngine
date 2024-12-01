package build

import (
	"blockEmulator/consensus_shard/pbft_all"
	"blockEmulator/networks"
	"blockEmulator/params"
	"blockEmulator/supervisor"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

func initConfig(nid, nnm, sid, snm uint64) *params.ChainConfig {
	path, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Current working directory:", path)
	}
	// Read the contents of ipTable.json
	file, err := os.ReadFile(path + "/ipTable.json")
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	// Create a map to store the IP addresses
	var ipMap map[uint64]map[uint64]string
	// Unmarshal the JSON data into the map
	err = json.Unmarshal(file, &ipMap)
	if err != nil {
		// handle error
		fmt.Println(err)
	}

	params.IPmap_nodeTable = ipMap
	params.SupervisorAddr = params.IPmap_nodeTable[params.SupervisorShard][0]

	// check the correctness of params
	if len(ipMap)-1 < int(snm) {
		log.Panicf("Input ShardNumber = %d, but only %d shards in ipTable.json.\n", snm, len(ipMap)-1)
	}
	for shardID := 0; shardID < len(ipMap)-1; shardID++ {
		if len(ipMap[uint64(shardID)]) < int(nnm) {
			log.Panicf("Input NodeNumber = %d, but only %d nodes in Shard %d.\n", nnm, len(ipMap[uint64(shardID)]), shardID)
		}
	}

	params.NodesInShard = int(nnm)
	params.ShardNum = int(snm)

	// if supervisor, dont limit bandwidth
	if nid == 123 {
		params.Bandwidth = -1
	}

	// init the network layer
	networks.InitNetworkTools()

	pcc := &params.ChainConfig{
		ChainID:        sid,
		NodeID:         nid,
		ShardID:        sid,
		Nodes_perShard: uint64(params.NodesInShard),
		ShardNums:      snm,
		BlockSize:      uint64(params.MaxBlockSize_global),
		BlockInterval:  uint64(params.Block_Interval),
		InjectSpeed:    uint64(params.InjectSpeed),
	}
	return pcc
}

func GetChainConfig(nid, nnm, sid, snm uint64) *params.ChainConfig {
	return initConfig(nid, nnm, sid, snm)
}

func BuildSupervisor(nnm, snm uint64) {
	methodID := params.ConsensusMethod
	var measureMod []string
	if methodID == 0 || methodID == 2 {
		measureMod = params.MeasureBrokerMod
	} else {
		measureMod = params.MeasureRelayMod
	}
	measureMod = append(measureMod, "Tx_Details")

	lsn := new(supervisor.Supervisor)
	lsn.NewSupervisor(params.SupervisorAddr, initConfig(123, nnm, 123, snm), params.CommitteeMethod[methodID], measureMod...)
	time.Sleep(3000 * time.Millisecond)
	go lsn.SupervisorTxHandling()
	lsn.TcpListen()
}

func BuildNewPbftNode(nid, nnm, sid, snm uint64) {
	methodID := params.ConsensusMethod
	worker := pbft_all.NewPbftNode(sid, nid, initConfig(nid, nnm, sid, snm), params.CommitteeMethod[methodID])
	go worker.Reconfig.TcpListen() // reconfiguration module
	go worker.TcpListen()
	worker.Propose()
}

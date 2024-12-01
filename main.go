package main

import (
	"blockEmulator/build"
	"blockEmulator/networks"
	"blockEmulator/params"
	"encoding/csv"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	"github.com/spf13/pflag"
)

var (
	// network config
	shardNum int
	nodeNum  int
	shardID  int
	nodeID   int

	// supervisor or not
	isSupervisor bool

	// batch running config
	isGen                bool
	isGenerateForExeFile bool

	//sync
	frequency      int
	syncMod        int
	migrateNodeNum int
	useFilter      int
)

func main() {

	// Generate bat files
	pflag.BoolVarP(&isGen, "gen", "g", false, "isGen is a bool value, which indicates whether to generate a batch file")
	pflag.BoolVarP(&isGenerateForExeFile, "shellForExe", "f", false, "isGenerateForExeFile is a bool value, which is effective only if 'isGen' is true; True to generate for an executable, False for 'go run'. ")

	// Start a node.
	pflag.IntVarP(&shardNum, "shardNum", "S", params.ShardNum, "shardNum is an Integer, which indicates that how many shards are deployed. ")
	pflag.IntVarP(&nodeNum, "nodeNum", "N", params.NodesInShard, "nodeNum is an Integer, which indicates how many nodes of each shard are deployed. ")
	pflag.IntVarP(&shardID, "shardID", "s", -1, "shardID is an Integer, which indicates the ID of the shard to which this node belongs. Value range: [0, shardNum). ")
	pflag.IntVarP(&nodeID, "nodeID", "n", -1, "nodeID is an Integer, which indicates the ID of this node. Value range: [0, nodeNum).")
	pflag.BoolVarP(&isSupervisor, "supervisor", "c", false, "isSupervisor is a bool value, which indicates whether this node is a supervisor.")
	pflag.IntVarP(&frequency, "frequency", "e", 50, "frequency is an Integer, which indicates the frequency of the reconfig. request. ")
	pflag.IntVarP(&syncMod, "syncMod", "m", 2, "mod of reconfiguration 0-full, 1-tmpt, 2-presync, 3.full-MPT ")
	pflag.IntVarP(&migrateNodeNum, "MigrateNodeNum", "q", 6, "num of migrate node")
	pflag.IntVarP(&useFilter, "useFilter", "u", 1, "user filter func")

	pflag.Parse()

	// Read basic configs
	params.ReadConfigFile(shardNum, nodeNum)
	if useFilter == 0 {
		params.UseFilter = false
	}

	params.Frequency = frequency + 1
	params.SyncMod = syncMod
	params.MigrateNodeNum = migrateNodeNum
	params.ZoneSize = migrateNodeNum / 2

	networks.InitNetworkSpecial()

	for i := 1; i <= migrateNodeNum/params.ZoneSize; i++ {
		params.SourceNodeID = append(params.SourceNodeID, migrateNodeNum+i)
	}
	fmt.Printf("source node id %v\n", params.SourceNodeID)

	// update result dir
	params.DataWrite_path = params.DataWrite_path + fmt.Sprintf("Mod%v_Num%v_Zone%v_frq%v_band%v/",
		params.SyncMod, params.MigrateNodeNum, params.ZoneSize, frequency, params.Bandwidth)

	go monitorResourceUsageCSV(params.DataWrite_path, shardID, nodeID)
	go monitorResourceUsageSingleProcessCSV(params.DataWrite_path, shardID, nodeID)
	if isGen {
		if isGenerateForExeFile {
			// Determine the current operating system.
			// Generate the corresponding .bat file or .sh file based on the detected operating system.
			os := runtime.GOOS
			switch os {
			case "windows":
				build.Exebat_Windows_GenerateBatFile(nodeNum, shardNum)
			case "darwin":
				build.Exebat_MacOS_GenerateShellFile(nodeNum, shardNum)
			case "linux":
				build.Exebat_Linux_GenerateShellFile(nodeNum, shardNum)
			}
		} else {
			// Without determining the operating system.
			// Generate a .bat file or .sh file for running `go run`.
			build.GenerateBatFile(nodeNum, shardNum)
			build.GenerateShellFile(nodeNum, shardNum)
		}

		return
	}

	go func() {
		if nodeID == 1 && shardID == 1 {
			fmt.Println("pprof start")
			err := http.ListenAndServe("127.0.0.1:20000", nil)
			if err != nil {
				log.Fatalf("ListenAndServe: %v", err)
			}

		}
	}()

	if isSupervisor {
		build.BuildSupervisor(uint64(nodeNum), uint64(shardNum))
	} else {
		if shardID >= shardNum || shardID < 0 {
			log.Panicf("Wrong ShardID. This ShardID is %d, but only %d shards in the current config. ", shardID, shardNum)
		}
		if nodeID >= nodeNum || nodeID < 0 {
			log.Panicf("Wrong NodeID. This NodeID is %d, but only %d nodes in the current config. ", nodeID, nodeNum)
		}
		build.BuildNewPbftNode(uint64(nodeID), uint64(nodeNum), uint64(shardID), uint64(shardNum))
	}

}

// 后台记录 CPU 和内存使用情况的函数
func monitorResourceUsageCSV(filePath string, shardID, nodeID int) {
	if nodeID != 0 {
		return

	}
	filePath = filePath + fmt.Sprintf("cpuS%v/", shardID)
	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		log.Panic(err)
	}
	// 创建或打开 CSV 文件
	file, err := os.OpenFile(filePath+fmt.Sprintf("s%vn%v_cpu_mem_usage.csv", shardID, nodeID), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Error opening CSV file:", err)
		return
	}
	defer file.Close()

	// 创建 CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 如果文件是新建的，则写入 CSV 表头
	fileInfo, _ := file.Stat()
	if fileInfo.Size() == 0 {
		err := writer.Write([]string{"Timestamp", "CPU_Usage_Percent", "Memory_Usage_Percent"})
		if err != nil {
			fmt.Println("Error writing header to CSV:", err)
			return
		}
	}

	for {
		// 获取 CPU 使用率
		cpuPercent, err := cpu.Percent(0, false)
		if err != nil {
			fmt.Println("Error getting CPU usage:", err)
			continue
		}

		// 获取内存使用率
		vmStat, err := mem.VirtualMemory()
		if err != nil {
			fmt.Println("Error getting memory usage:", err)
			continue
		}

		// 记录当前时间和 CPU、内存使用情况
		record := []string{
			fmt.Sprintf("%d", time.Now().UnixMilli()),
			fmt.Sprintf("%.2f", cpuPercent[0]),
			fmt.Sprintf("%.2f", vmStat.UsedPercent),
		}

		// 写入 CSV 文件
		err = writer.Write(record)
		if err != nil {
			fmt.Println("Error writing record to CSV:", err)
			continue
		}

		// 刷新缓冲区，确保数据写入文件
		writer.Flush()

		// 等待一段时间
		time.Sleep(1 * time.Second) // 每秒记录一次
	}
}

// 后台记录 CPU 和内存使用情况的函数
func monitorResourceUsageSingleProcessCSV(filePath string, shardID, nodeID int) {
	if shardID == -1 {
		return
	}

	// 获取当前进程的 PID
	pid := int32(os.Getpid())

	// 创建进程对象
	p, err := process.NewProcess(pid)
	if err != nil {
		fmt.Println("Error creating process instance:", err)
		return
	}
	filePath = filePath + fmt.Sprintf("cpuS%v/", shardID)
	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		log.Panic(err)
	}
	// 创建或打开 CSV 文件
	file, err := os.OpenFile(filePath+fmt.Sprintf("s%vn%v_cpu_mem_usage_single.csv", shardID, nodeID), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Error opening CSV file:", err)
		return
	}
	defer file.Close()

	// 创建 CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 如果文件是新建的，则写入 CSV 表头
	fileInfo, _ := file.Stat()
	if fileInfo.Size() == 0 {
		err := writer.Write([]string{"Timestamp", "CPU_Usage_Percent", "Memory_Usage_Percent"})
		if err != nil {
			fmt.Println("Error writing header to CSV:", err)
			return
		}
	}

	for {
		// 获取 CPU 使用率
		cpuPercent, err := p.CPUPercent()
		if err != nil {
			fmt.Println("Error getting CPU usage:", err)
			continue
		}

		// 获取内存使用情况
		memInfo, err := p.MemoryInfo()
		if err != nil {
			fmt.Println("Error getting memory usage:", err)
			continue
		}

		// 记录当前时间和 CPU、内存使用情况
		record := []string{
			fmt.Sprintf("%d", time.Now().UnixMilli()),
			fmt.Sprintf("%.2f", cpuPercent),
			fmt.Sprintf("%.2f", float64(memInfo.RSS)/1024/1024), // 将内存转换为 MB
		}

		// 写入 CSV 文件
		err = writer.Write(record)
		if err != nil {
			fmt.Println("Error writing record to CSV:", err)
			continue
		}

		// 刷新缓冲区，确保数据写入文件
		writer.Flush()

		// 等待一段时间
		time.Sleep(1 * time.Second) // 每秒记录一次
	}
}

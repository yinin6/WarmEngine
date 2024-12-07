import time
import os

import pandas as pd
import paramiko
from scp import SCPClient
from threading import Thread
import json

shard_num = 8
node_num_in_shard = 10
syncMod = 3

MigrateNodeNum = 6

frequency = 10

node_num = shard_num * node_num_in_shard

ports = list()
port_to_shardID_nodeID = dict()

exeFileName = "myapp_linux"
remotePath = "~/workspace/"

user = "root"
pwd = "i-hp35me2k9rs4oo35nben"

remote_ip_to_local = {
    "172.16.0.154": "39.104.70.77",
    "172.16.0.155": "39.104.68.197"
}


def set_ports(node_num=20):
    global ports
    global remote_ip_to_local

    ip_table_path = f"./ali_result/ecs_instance_list_cn-huhehaote_2024-11-20.csv"

    df = pd.read_csv(ip_table_path)

    ip_local = df["内网IP"].tolist()
    ip_remote = df["公网IP"].tolist()

    for i in range(len(ip_local)):
        remote_ip_to_local[ip_local[i]] = ip_remote[i]

    file_path = f"./ip/{node_num}/ipTable.json"

    with open(file_path, "r") as file:
        data = json.load(file)

    for shard_id, list_of_shard in data.items():
        for node_id, address in list_of_shard.items():
            ports.append(address)
            port_to_shardID_nodeID[address] = (shard_id, node_id)

    print(len(ports), ports)
    print(len(port_to_shardID_nodeID), port_to_shardID_nodeID)
    print(len(remote_ip_to_local), remote_ip_to_local)


def scp_files_to_remote_directly(send_bin=False):
    mkdir_exp_dir()
    temp = dict()
    for port in ports:
        if temp.get(port.split(":")[0]):
            continue
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(
            paramiko.AutoAddPolicy()
        )  # 自动添加主机名及主机密钥到本地HostKeys对象，并保存，只在代理模式下使用
        print(f"Connecting======={port}")
        t = port.split(":")[0]
        temp[t] = True
        ssh.connect(
            remote_ip_to_local[t], username=user, password=pwd, port=22
        )  # 也可以使用key_filename参数提供私钥文件路径
        with SCPClient(ssh.get_transport()) as scp:
            if send_bin:
                scp.put("./bin/" + exeFileName, recursive=True, remote_path=remotePath)
            scp.put("./bin/paramsConfig.json", recursive=True, remote_path=remotePath)
            scp.put(f"./ip/{node_num}/ipTable.json", recursive=True, remote_path=remotePath)
            scp.close()
        ssh.close()


def download_from_remote(remote_path, local_path):
    #判断本地路径是否存在，不存在则创建

    if not os.path.exists(local_path):
        os.makedirs(local_path)
    time.sleep(2)
    # 创建SSHClient对象
    temp = dict()
    for port in ports:
        if temp.get(port.split(":")[0]):
            continue
        temp[port.split(":")[0]] = True
        t = port.split(":")[0]
        print("address:", t)
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        ssh.connect(
            remote_ip_to_local[t], username=user, password=pwd, port=22
        )  # 也可以使用key_filename参数提供私钥文件路径
        print("Connected!!!!!!!!")
        with SCPClient(ssh.get_transport()) as scp:
            scp.get(remote_path, local_path, recursive=True)
            scp.close()
        ssh.close()


def clear_history_expData():
    temp = dict()
    for port in ports:
        if temp.get(port.split(":")[0]):
            continue
        temp[port.split(":")[0]] = True
        run_exe_remote(
            port=port,
            cmd=f"killall -9 {exeFileName} & rm -rf {remotePath}/expTest & rm -rf {remotePath}/result",
        )
        run_exe_remote(
            port=port,
            cmd=f"cd {remotePath} && fuser -k {exeFileName}",
        )


def run_exe_remote(port, cmd):
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(
        paramiko.AutoAddPolicy()
    )  # 自动添加主机名及主机密钥到本地HostKeys对象，并保存，只在代理模式下使用
    print(f"Connecting======={port}")

    address = port.split(":")
    print(address)

    ssh.connect(remote_ip_to_local[address[0]], username=user, password=pwd,
                port=22)  # 也可以使用key_filename参数提供私钥文件路径
    stdin, stdout, stderr = ssh.exec_command(cmd)

    # 输出错误信息
    print(f"Standard error for command '{cmd}':")
    for line in stderr:
        print(line.strip())

    # 输出命令执行结果
    print(f"Standard output for command '{cmd}':")
    for line in stdout:
        print(line.strip())

    ssh.close()


def multi_Clients_runing():
    threads = []  # 用于存储所有线程对象

    address_to_command = dict()

    for port in ports:
        address = port.split(":")

        shardID, nodeID = port_to_shardID_nodeID[port]
        print(shardID, nodeID, port)

        cmd = f"cd {remotePath} && ./{exeFileName}  -n {nodeID} -N {node_num_in_shard} -s {shardID} -S {shard_num} --syncMod {syncMod} --frequency {frequency} --MigrateNodeNum {MigrateNodeNum} > S{shardID}N{nodeID}.log"
        if int(shardID) > 100:
            cmd = f"cd {remotePath} &&  ./{exeFileName}  -c -N {node_num_in_shard} -S {shard_num} --syncMod {syncMod} --frequency {frequency}"

        if not address_to_command.get(address[0]):
            address_to_command[address[0]] = " "

        address_to_command[address[0]] += cmd + " & "

    for address, cmd in address_to_command.items():
        client = Thread(target=run_exe_remote, args=(address, cmd + " wait"))
        threads.append(client)  # 将线程对象添加到列表中
        client.start()

    # 等待所有线程执行完毕
    for thread in threads:
        thread.join()


def mkdir_exp_dir():
    temp = dict()
    for port in ports:
        if temp.get(port.split(":")[0]):
            continue
        temp[port.split(":")[0]] = True
        run_exe_remote(port=port, cmd=f"mkdir -p {remotePath}")


def chmod_executionFile():
    temp = dict()
    for port in ports:
        if temp.get(port.split(":")[0]):
            continue
        temp[port.split(":")[0]] = True
        run_exe_remote(
            port=port,
            cmd=f"cd {remotePath} && chmod +x ./{exeFileName}",
        )


if __name__ == "__main__":

    frequency = 50
    shard_num = 16
    node_num_in_shard = 10
    node_num = shard_num * node_num_in_shard
    MigrateNodeNum = 6

    set_ports(node_num)

    mods = [1, 2, 3, 0]


    for i in mods:
        syncMod = i
        clear_history_expData()
        if i == 1:
            scp_files_to_remote_directly(True)
        chmod_executionFile()

        multi_Clients_runing()
        download_from_remote(remote_path=f'{remotePath}/result', local_path='./ali_result/vary_inject')

        time.sleep(10)

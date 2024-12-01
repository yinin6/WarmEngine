import os
import random


def generate_bat_file(nodenum, shardnum):
    dir_path = f'./{nodenum*shardnum}/'

    if not os.path.exists(dir_path):
        os.makedirs(dir_path)

    file_name = dir_path + f"bat_compile_run_shardNum={shardnum}_NodeNum={nodenum}.bat"

    with open(file_name, 'w') as ofile:
        str_cmd = f' rmdir /s /q "expTest" \n\n'
        ofile.write(str_cmd)

        str_cmd = f'del /q "main.exe" \n\n'
        ofile.write(str_cmd)

        str_cmd = f'start cmd /k go build -o main.exe main.go\n\n'
        ofile.write(str_cmd)

        str_cmd = f'timeout /t 3 /nobreak >nul \n\n'
        ofile.write(str_cmd)





        for i in range(1, nodenum):
            for j in range(shardnum):
                str_cmd = f'start cmd /k main.exe -n {i} -N {nodenum} -s {j} -S {shardnum}\n\n'
                ofile.write(str_cmd)

        for j in range(shardnum):
            str_cmd = f'start cmd /k main.exe -n 0 -N {nodenum} -s {j} -S {shardnum}\n\n'
            ofile.write(str_cmd)

        str_cmd = f'start cmd /k main.exe -c -N {nodenum} -S {shardnum}\n\n'
        ofile.write(str_cmd)


def generate_ip_table(shard_size, shard_num):
    ip = '127.0.0.1'
    data = {}
    dir_path = f'./{shard_size*shard_num}/'
    if not os.path.exists(dir_path):
        os.makedirs(dir_path)

    file_path = dir_path+'ipTable.json'

    k = 0
    port = 30001
    # 遍历每个IP地址，生成随机端口
    for i in range(shard_num):
        data[str(i)]={}
        # 随机选择端口
        for j in range(shard_size):
            data[str(i)][str(j)] = f"{ip}:{port}"
            port += random.randint(1, 10)

    data["2147483647"]={}
    data["2147483647"][str(0)] = f"{ip}:38800"

    import json



    with open(file_path, 'w') as file:
        json.dump(data, file, indent=4)

    print(f"Data saved to {file_path}")


def Exebat_Linux_GenerateShellFile(nodenum, shardnum):

    dir_path = f'./{nodenum*shardnum}/'

    if not os.path.exists(dir_path):
        os.makedirs(dir_path)


    # 文件名定义
    file_name = dir_path + f"linux_shell_shardNum={shardnum}_NodeNum={nodenum}.sh"

    # 创建并打开文件
    with open(file_name, "w") as ofile:
        # 写入文件头
        ofile.write("#!/bin/bash\n\n")

        # 写入每个 shard 和节点的启动命令
        for j in range(shardnum):
            for i in range(1, nodenum):
                command = f"./myapp_linux -n {i} -N {nodenum} -s {j} -S {shardnum} &\n\n"
                ofile.write(command)

        # 写入每个 shard 的主节点的启动命令
        for j in range(shardnum):
            command = f"./myapp_linux -n 0 -N {nodenum} -s {j} -S {shardnum} &\n\n"
            ofile.write(command)

        # 写入最后的汇总命令
        final_command = f"./myapp_linux -c -N {nodenum} -S {shardnum} &\n\n"
        ofile.write(final_command)



if __name__ == '__main__':
    node_num = 10
    shard_num= 4
    generate_bat_file(node_num, shard_num)
    generate_ip_table(node_num, shard_num)
    Exebat_Linux_GenerateShellFile(node_num, shard_num)
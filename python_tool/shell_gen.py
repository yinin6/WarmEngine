import os

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

Exebat_Linux_GenerateShellFile(10, 4)
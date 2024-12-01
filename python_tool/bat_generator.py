import os

def generate_bat_file(nodenum, shardnum):
    file_name = f"bat_compile_run_shardNum={shardnum}_NodeNum={nodenum}.bat"

    with open(file_name, 'w') as ofile:
        for i in range(1, nodenum):
            for j in range(shardnum):
                str_cmd = f'start cmd /k go run main.go -n {i} -N {nodenum} -s {j} -S {shardnum}\n\n'
                ofile.write(str_cmd)

        for j in range(shardnum):
            str_cmd = f'start cmd /k go run main.go -n 0 -N {nodenum} -s {j} -S {shardnum}\n\n'
            ofile.write(str_cmd)

        str_cmd = f'start cmd /k go run main.go -c -N {nodenum} -S {shardnum}\n\n'
        ofile.write(str_cmd)


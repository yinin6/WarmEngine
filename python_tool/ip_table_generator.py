import os
import random

from python_tool.bat_generator import generate_bat_file

shard_num = 2
shard_size= 6
ip = '127.0.0.1'
data = {}


dir_path = f'./{shard_size*shard_num}/'
if not os.path.exists(dir_path):
    os.makedirs(dir_path)

file_path = dir_path+'ip.json'

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

generate_bat_file(shard_size, shard_num)
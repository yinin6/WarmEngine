import os
import random

import pandas as pd

shard_num = 16
shard_size = 10
ip = []
data = {}

ip_path = "./ali_result/ecs_instance_list_cn-huhehaote_2024-11-20.csv"

df = pd.read_csv(ip_path)

ip = df["内网IP"].tolist()

dir_path = f'./ip/{shard_size * shard_num}/'
if not os.path.exists(dir_path):
    os.makedirs(dir_path)

file_path = dir_path + 'ipTable.json'

k = 0
port = 30001
# 遍历每个IP地址，生成随机端口
for i in range(shard_num):
    data[str(i)] = {}
    # 随机选择端口
    address=ip[i%(len(ip)-1)]
    for j in range(shard_size):
        data[str(i)][str(j)] = f"{address}:{port}"
        port += random.randint(1, 10)

data["2147483647"] = {}
data["2147483647"][str(0)] = f"{ip[-1]}:38800"

import json

with open(file_path, 'w') as file:
    json.dump(data, file, indent=4)

print(f"Data saved to {file_path}")

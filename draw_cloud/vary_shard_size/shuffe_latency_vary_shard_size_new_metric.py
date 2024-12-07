import pandas as pd

import matplotlib.pyplot as plt
import numpy as np

shard_num = 0

from brokenaxes import brokenaxes

bandwidth = 2000000


def get_single_shard_sync_data_of_each_epoch(mod=1, shard_id=0, node_num=10):
    move_node_num = int(node_num / 2) + 1
    zone_size = int(move_node_num / 2)
    file_path = f"../../ali_result/result/data=3000000_inj=10000_block=1000_S16N{node_num}/Mod{mod}_Num{move_node_num}_Zone{zone_size}_frq50_band{bandwidth}/pbft_shardNum=16/shuffeDatil.csvS{shard_id}0.csv"
    df = pd.read_csv(file_path)
    df['syncTime'] = df['cost']

    return df


def get_single_shard_sync_data_of_each_epoch_of_Proposed(shard_id=0, node_num=5):
    move_node_num = int(node_num / 2) + 1
    zone_size = int(move_node_num / 2)
    file_path = f"../../ali_result/result/data=3000000_inj=10000_block=1000_S16N{node_num}/Mod2_Num{move_node_num}_Zone{zone_size}_frq50_band{bandwidth}/pbft_shardNum=16/shuffeDatil.csvS{shard_id}0.csv"
    df = pd.read_csv(file_path)
    df['syncTime'] = df['cost']
    return df


def get_average_sync_time_of_each_epoch(mod=1, shard_id=0, node_num=10):
    if mod == 2:
        data = get_single_shard_sync_data_of_each_epoch_of_Proposed(shard_id, node_num)
        return data['syncTime'].mean() / 1000
    data = get_single_shard_sync_data_of_each_epoch(mod, shard_id, node_num)
    return data['syncTime'].mean() / 1000


shard_zises = [10, 15]

mod = [3, 1, 2]
all_data = []
for m in mod:
    sync_time_list = []
    for size in shard_zises:
        sync_time = get_average_sync_time_of_each_epoch(m, shard_num, size)
        sync_time_list.append(sync_time)
    all_data.append(sync_time_list)

print(all_data)
growth_rates = []
for row in all_data:
    old_value, new_value = row
    growth_rate = ((new_value - old_value) / old_value) * 100
    growth_rates.append(growth_rate)

print(growth_rates)


# 设置柱状图的宽度
bar_width = 0.15
# 设置x轴的位置
index = np.arange(len(shard_zises))
# 绘制柱状图
# 使用 colormap 生成颜色


colors = [
    (183 / 255, 183 / 255, 235 / 255, 0.5), (157 / 255, 158 / 255, 163 / 255, 0.5),
    (240 / 255, 155 / 255, 160 / 255, 0.5)]

# hatches = ['/', '\\', '|', '+']
labels = ['ETH-fast', 'tMPT', 'Proposed']
fig = plt.figure(figsize=(7, 6))
ax = plt.gca()

for i, data in enumerate(all_data):
    plt.bar(index + i * bar_width + (i * 0.03), data, bar_width, label=labels[i], color=colors[i], edgecolor='black', )

# 设置 x 轴刻度，使其位于每组柱子的中间
ax.set_xticks(range(0, len(shard_zises)), [f'        {int(b)}' for b in shard_zises], fontsize=16)

ax.set_xlabel('Num. of nodes in per shard', fontsize=20)
ax.set_ylabel('Reconfiguration Time (Sec.)', fontsize=20)
plt.legend(fontsize=20, ncol=2)

ax.tick_params(axis='y', labelsize=20)
ax.tick_params(axis='x', labelsize=20)

plt.tight_layout()
plt.savefig(f'../pics/shuffe_latency_vary_bandwidth.pdf')
plt.show()

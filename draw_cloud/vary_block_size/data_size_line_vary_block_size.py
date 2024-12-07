import pandas as pd

import matplotlib.pyplot as plt
import numpy as np

shard_num = 4

from brokenaxes import brokenaxes

bandwidth = 1000000
font_size=24

def get_single_shard_sync_data_of_each_epoch(mod=1, shard_id=0, block_size=1000):
    dataframes = []
    node_num = 10
    move_node_num = int(node_num / 2) + 1
    zone_size = int(move_node_num / 2)
    for i in range(1, move_node_num + 1):
        file_path = f"../../ali_result/vary_blocksize/result/data=5000000_inj=10000_block={block_size}_S16N{node_num}/Mod{mod}_Num{move_node_num}_Zone{zone_size}_frq50_band{bandwidth}/S{shard_id}N{i}.csv"

        df = pd.read_csv(file_path)
        dataframes.append(df)

    merged_df = pd.concat(dataframes, axis=0)

    # 以 epoch 和 shardID 为主键分组，计算每组内最小的 beginTime 和最大的 overTime
    grouped_df = merged_df.groupby(['epoch']).agg(
        sync_DataSize=('stateDataSize', 'min'),
        min_beginTime=('beginTime', 'min'),
        max_overTime=('overTime', 'max')

    ).reset_index()

    # 计算 syncTime 列，syncTime 为每组内的 overTime 和 beginTime 的差值
    grouped_df['syncTime'] = grouped_df['max_overTime'] - grouped_df['min_beginTime']
    return grouped_df


def get_single_shard_sync_data_of_each_epoch_of_Proposed(shard_id=0, block_size=1000):
    dataframes = []
    node_num = 10
    move_node_num = int(node_num / 2) + 1
    zone_size = int(move_node_num / 2)
    for i in range(1, move_node_num + 1):
        file_path = f"../../ali_result/vary_blocksize/result/data=5000000_inj=10000_block={block_size}_S16N{node_num}/Mod2_Num{move_node_num}_Zone{zone_size}_frq50_band{bandwidth}/S{shard_id}N{i}specific.csv"
        df = pd.read_csv(file_path)
        dataframes.append(df)

    merged_df = pd.concat(dataframes, axis=0)

    # 以 epoch 和 round 为主键分组，计算每组内最小的 beginTime 和最大的 overTime, 以及每组内的 syncDataSize 均值
    grouped_df = merged_df.groupby(['epoch', 'round']).agg(
        sync_DataSize=('stateDataSize', 'min'),
        min_beginTime=('beginTime', 'min'),
        max_overTime=('overTime', 'max')
    ).reset_index()
    # 计算 syncTime 列，syncTime 为每组内的 overTime 和 beginTime 的差值
    grouped_df['syncTime'] = grouped_df['max_overTime'] - grouped_df['min_beginTime']
    # 取 round = 0 的数据
    #
    # grouped_df = grouped_df.groupby(['epoch']).agg(
    #     sync_DataSize=('sync_DataSize', 'sum'),
    # ).reset_index()

    grouped_df = grouped_df[grouped_df['round'] == 0]
    return grouped_df


def get_average_sync_data_of_each_epoch(mod=1, shard_id=0, blocksize=10):
    if mod == 2:
        data = get_single_shard_sync_data_of_each_epoch_of_Proposed(shard_id, blocksize)
        print(mod, blocksize, len(data))
        if len(data) > 4:
            data = data[:5]

        return data['sync_DataSize'].mean()
    data = get_single_shard_sync_data_of_each_epoch(mod, shard_id, blocksize)
    print(mod, blocksize, len(data))
    if len(data) > 4:
        data = data[:5]
    return data['sync_DataSize'].mean()


block_sizes = [500, 1000, 1500, 2000]

mod = [3, 1, 2]
all_data = []
for m in mod:
    sync_time_list = []
    for size in block_sizes:
        sync_time = get_average_sync_data_of_each_epoch(m, shard_num, size)
        sync_time_list.append(sync_time)
    all_data.append(sync_time_list)

# 设置柱状图的宽度
bar_width = 0.15
# 设置x轴的位置
index = np.arange(len(block_sizes))
# 绘制柱状图
# 使用 colormap 生成颜色


colors = [
    (183 / 255, 183 / 255, 235 / 255, 0.5), (157 / 255, 158 / 255, 163 / 255, 0.5),
    (240 / 255, 155 / 255, 160 / 255, 0.5)]

# hatches = ['/', '\\', '|', '+']
labels = ['ETH-fast', 'tMPT', 'Proposed']
fig = plt.figure(figsize=(7, 5))
ax = plt.gca()

for i, data in enumerate(all_data):
    plt.bar(index + i * bar_width + (i * 0.03), data, bar_width, label=labels[i], color=colors[i], edgecolor='black', )

# 设置 x 轴刻度，使其位于每组柱子的中间
ax.set_xticks(range(0, len(block_sizes)), [f'        {int(b)}' for b in block_sizes], fontsize=16)

ax.set_xlabel('Max block size (TXs)', fontsize=font_size)
ax.set_ylabel('Reconfig. data size (Bytes)', fontsize=font_size, y=0.4)
plt.legend(fontsize=22, ncol=2)

ax.tick_params(axis='y', labelsize=font_size)
ax.tick_params(axis='x', labelsize=font_size)

plt.tight_layout()
plt.savefig(f'../pics/shuffe_datasize_vary_blocksize.pdf')
plt.show()

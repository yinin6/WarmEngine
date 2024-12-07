import pandas as pd

import matplotlib.pyplot as plt
import numpy as np
shard_num = 0
move_node_num = 6
from brokenaxes import brokenaxes


def get_single_shard_sync_data_of_each_epoch(mod=1,shard_id=0, bandwidth=500000):
    dataframes = []
    for i in range(1, move_node_num + 1):
        # file_path = f"./result/data=1500000_inj=150000_block=1000/Mod{mod}_Num{move_node_num}_Zone3_frq50_band{bandwidth}_S4N4/S{shard_id}N{i}.csv"
        file_path = f"../../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod{mod}_Num{move_node_num}_Zone3_frq50_band{bandwidth}/S{shard_id}N{i}.csv"

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
    print(grouped_df)
    return grouped_df





def get_single_shard_sync_data_of_each_epoch_of_Proposed(shard_id=0, bandwidth=500000):
    dataframes = []
    for i in range(1, move_node_num + 1):
        # file_path = f"./result/data=1500000_inj=150000_block=1000/Mod2_Num{move_node_num}_Zone3_frq50_band{bandwidth}_S4N4/S{shard_id}N{i}specific.csv"
        file_path = f"../../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod2_Num{move_node_num}_Zone3_frq50_band{bandwidth}/S{shard_id}N{i}specific.csv"

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
    grouped_df = grouped_df[grouped_df['round'] == 0]
    print(grouped_df)
    return grouped_df


def get_average_sync_time_of_each_epoch(mod=1,shard_id=0, bandwidth=500000):
    if mod == 2:
        data = get_single_shard_sync_data_of_each_epoch_of_Proposed(shard_id, bandwidth)
        return data['syncTime'].mean()/1000
    data = get_single_shard_sync_data_of_each_epoch(mod,shard_id, bandwidth)
    return data['syncTime'].mean()/1000



bandwidths = [500000, 1000000, 1500000,  2000000]

mod=[0,3,1,2]
all_data=[]
for m in mod:
    sync_time_list = []
    for bandwidth in bandwidths:
        sync_time = get_average_sync_time_of_each_epoch(m, shard_num, bandwidth)
        sync_time_list.append(sync_time)
    all_data.append(sync_time_list)

print(all_data)
# 设置柱状图的宽度
bar_width = 0.15
# 设置x轴的位置
index = np.arange(len(bandwidths))
# 绘制柱状图
# 使用 colormap 生成颜色
colors = ['skyblue','#B7B7EB',  '#9D9EA3', '#F09BA0']

colors = [  (135 / 255, 206 / 255, 163 / 235, 0.5),
              (183 / 255, 183 / 255, 235 / 255, 0.5), (157 / 255, 158 / 255, 163 / 255, 0.5),
              (240 / 255, 155 / 255, 160 / 255, 0.5)]

mod_to_color = {
    0: '#7195C5',
    3: '#7262ac',
    1: '#01844F',
    2: '#E9212C'
}




# hatches = ['/', '\\', '|', '+']
labels = ['ETH-full', 'ETH-fast', 'tMPT', 'Proposed']
fig = plt.figure(figsize=(7, 5))
bax = brokenaxes(ylims=((0, 50), (80, 90)), hspace=0.05)


for i, data in enumerate(all_data):
    bax.bar(index + i * bar_width + (i*0.03), data, bar_width, label=labels[i], color=colors[i],edgecolor='black', )

# 设置 x 轴刻度，使其位于每组柱子的中间
bax.set_xticks( range(0,len(bandwidths)), [f'         {int(b/125000)} Mbps' for b in bandwidths], fontsize=16)



# bax.set_xlabel('Bandwidth')
bax.set_ylabel('Reconfig. latency (Sec.)', fontsize=24, labelpad=35)
bax.legend( fontsize=22 ,  ncol=2)


bax.tick_params(axis='y', labelsize=22)
bax.tick_params(axis='x', labelsize=22)
bax.axs[0].set_yticks([100],['100'])
# plt.tight_layout()
plt.savefig(f'../pics/shuffe_latency_vary_bandwidth.pdf')
plt.show()

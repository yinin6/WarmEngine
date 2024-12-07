# result/data=1500000_inj=150000_block=1000/Mod1_Num6_Zone3_frq50_band1000000_S4N4/S0N0.csv

import pandas as pd

import matplotlib.pyplot as plt
from mpl_toolkits.axes_grid1.inset_locator import inset_axes, mark_inset

shard_num = 0
move_node_num = 6



def get_single_shard_sync_data_of_each_epoch(mod=1,shard_id=0):
    dataframes = []
    for i in range(1, move_node_num + 1):
        file_path = f"./result/data=1500000_inj=150000_block=1000/Mod{mod}_Num{move_node_num}_Zone3_frq50_band500000_S4N4/S{shard_id}N{i}.csv"
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





def get_single_shard_sync_data_of_each_epoch_of_Proposed(shard_id):
    dataframes = []
    for i in range(1, move_node_num + 1):
        file_path = f"./result/data=1500000_inj=150000_block=1000/Mod2_Num{move_node_num}_Zone3_frq50_band500000_S4N4/S{shard_id}N{i}specific.csv"
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

df0 = get_single_shard_sync_data_of_each_epoch(mod=0,shard_id=0)
df3 = get_single_shard_sync_data_of_each_epoch(mod=3,shard_id=0)
df1 = get_single_shard_sync_data_of_each_epoch(mod=1,shard_id=0)
df2 = get_single_shard_sync_data_of_each_epoch_of_Proposed(0)


plt.figure(figsize=(7, 6))
colors = ['#7195C5', '#7262ac', '#01844F', '#E9212C']
labels = ['ETH-full', 'ETH-fast', 'tMPT', 'Proposed']
markers = ['o', 's', 'D', '^']
for i, df in enumerate([df0, df3, df1,  df2]):
    # 截取前五列
    df['syncTime'] = df['syncTime'] / 1000
    df = df[:5]


    plt.plot(df['epoch'], df['syncTime'], label=labels[i], color=colors[i], marker=markers[i], markerfacecolor='none',
             markersize=10)
    # plt.fill_between(df['epoch'], df['syncTime'], color=colors[i], alpha=0.2)
plt.xlabel('Epoch', fontsize=16)
plt.ylabel('Reconfiguration Latency (Sec)', fontsize=16)

plt.xticks(fontsize=16)
plt.yticks(fontsize=16)

# plt.ylim(0,1500)
plt.legend(fontsize=16, ncol=2)
plt.xticks([1, 2, 3, 4, 5], fontsize=16)  # 设置 x 轴刻度


# 添加局部放大图
ax = plt.gca()
ax_inset = inset_axes(ax, width="40%", height="30%",loc='center right')

for i, df in enumerate([df0, df3, df1, df2]):
    df = df[:5]
    ax_inset.plot(df['epoch'], df['syncTime'], label=labels[i], color=colors[i], marker=markers[i], markerfacecolor='none', markersize=5)


ax_inset.set_xlim(2.8, 5.1)
ax_inset.set_ylim(0, 350)
mark_inset(ax, ax_inset, loc1=2, loc2=4, fc="none", ec="0.8")

ax_inset.tick_params(axis='both', which='major', labelsize=10)



plt.tight_layout()
plt.savefig(f'./pics/reconfiguration_latency_line.pdf')
plt.show()


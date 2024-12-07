import pandas as pd

import matplotlib.pyplot as plt
import numpy as np



shard_num = 0
move_node_num = 6
def get_single_shard_sync_data_of_each_epoch_of_Proposed(shard_id=0, bandwidth=500000):
    dataframes = []
    for i in range(1, move_node_num + 1):
        file_path = f"./result/data=1500000_inj=150000_block=1000/Mod2_Num{move_node_num}_Zone3_frq50_band{bandwidth}_S4N4/S{shard_id}N{i}specific.csv"
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

    return grouped_df


def get_single_shard_sync_data_of_each_epoch(mod=1,shard_id=0, bandwidth=500000):
    dataframes = []
    for i in range(1, move_node_num + 1):
        file_path = f"./result/data=1500000_inj=150000_block=1000/Mod{mod}_Num{move_node_num}_Zone3_frq50_band{bandwidth}_S4N4/S{shard_id}N{i}.csv"
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


def draw_pic_of_proposed():
    bandwidths = [500000, 1000000]
    data = []
    for i in bandwidths:
        df = get_single_shard_sync_data_of_each_epoch_of_Proposed(shard_id=0, bandwidth=i)
        plt.figure(figsize=(10, 6))
        data.append(df['syncTime'])


    plt.figure(figsize=(10, 6))
    plt.boxplot(data, vert=True, patch_artist=True, labels=[f'{bw/125000} Mbps' for bw in bandwidths])
    plt.title('Sync Time Distribution for Different Bandwidths')
    plt.xlabel('Bandwidth')
    plt.ylabel('Sync Time')
    plt.show()



def draw_pic_of_tmpt():
    bandwidths = [500000, 1000000]
    data = []
    for i in bandwidths:
        df = get_single_shard_sync_data_of_each_epoch(mod=1,shard_id=0, bandwidth=i)
        plt.figure(figsize=(10, 6))
        data.append(df['syncTime'])

    plt.figure(figsize=(10, 6))
    plt.boxplot(data, vert=True, patch_artist=True, labels=[f'{bw / 125000} Mbps' for bw in bandwidths])
    plt.title('Sync Time Distribution for Different Bandwidths')
    plt.xlabel('Bandwidth')
    plt.ylabel('Sync Time')
    plt.show()

def draw_both():
    bandwidths = [500000, 1000000]
    data1 = []
    for i in bandwidths:
        df = get_single_shard_sync_data_of_each_epoch_of_Proposed(shard_id=0, bandwidth=i)
        data1.append(df['syncTime']/1000)

    data2 = []
    for i in bandwidths:
        df = get_single_shard_sync_data_of_each_epoch(mod=1, shard_id=0, bandwidth=i)
        data2.append(df['syncTime']/1000)

    # 创建图表
    fig, ax1 = plt.subplots(figsize=(7, 6))

    # 绘制 data1 的箱线图
    box1 = ax1.boxplot(data1, positions=[1, 2], widths=0.3, patch_artist=True,
                       boxprops=dict(facecolor="lightpink"), medianprops=dict(color= 'white'))

    # 设置 data1 的 y 轴
    ax1.set_ylabel('Reconfiguration Time (Sec) - Proposed',fontsize=16)
    ax1.set_ylim(0, 6)  # 根据实际数据调整y轴范围

    ax1.set_xticklabels(['500000', '1000000'],fontsize=16)
    ax1.set_xlabel('Bandwidth (Mbps)',fontsize=16)

    # 创建右侧y轴并绘制 data2 的箱线图
    ax2 = ax1.twinx()
    box2 = ax2.boxplot(data2, positions=[1.4, 2.4], widths=0.3, patch_artist=True,
                       boxprops=dict(facecolor="lightgray"), medianprops=dict(color= 'white'))

    data1_means = [np.mean(d) for d in data1]
    data2_means = [np.mean(d) for d in data2]
    print(data1_means[0]/ data1_means[1])
    # 添加均值差的竖线

    ax1.plot([1,1 ], [data1_means[0], data1_means[1]], color="purple", linestyle="--", linewidth=1, marker="o")
    ax1.plot([1, 2], [data1_means[1], data1_means[1]], color="purple", linestyle="--", linewidth=1, marker="o")
    # 在均值位置绘制连接线

    ax1.text(0.8, (data1_means[0] + data1_means[1]) / 2 , "3.23×",
             color="purple", ha='center', fontsize=16)

    print(data2_means[0] / data2_means[1])
    ax2.plot([1.4, 2.4], [data2_means[0], data2_means[0]], color="purple", linestyle="--", linewidth=1, marker="o")
    ax2.plot([2.4, 2.4], [data2_means[1], data2_means[0]], color="purple", linestyle="--", linewidth=1, marker="o")
    # 在均值位置绘制连接线

    ax2.text(2.6, (data2_means[0] + data2_means[1]) / 2, "1.95×",
             color="purple", ha='center', fontsize=16)

    ax2.set_ylabel('Reconfiguration Time (Sec) - tMPT',fontsize=16)
    ax2.tick_params(axis='y', labelsize=16)
    ax1.tick_params(axis='y', labelsize=16)
    ax2.set_ylim(0, 120)  # 根据实际数据调整y轴范围
    ax2.set_xticks([1.2, 2.2],['4 Mbps', '8 Mbps'])
    # 添加图例
    ax1.legend([box1["boxes"][0], box2["boxes"][0]], ['Proposed', 'tMPT'],fontsize=16)

    plt.tight_layout()
    plt.savefig(f'./pics/sync_time_vs_bandwidth.pdf')

    plt.show()


draw_both()
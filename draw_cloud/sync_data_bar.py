

import pandas as pd

import matplotlib.pyplot as plt
from mpl_toolkits.axes_grid1.inset_locator import inset_axes, mark_inset


shard_num = 0
move_node_num = 6
font_size=25


def get_single_shard_sync_data_of_each_epoch(mod=1,shard_id=0):
    dataframes = []
    for i in range(1, move_node_num + 1):
        file_path = f"../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod{mod}_Num6_Zone3_frq50_band500000/S{shard_id}N{i}.csv"
        # file_path = f"./result/data=1500000_inj=150000_block=1000/Mod{mod}_Num{move_node_num}_Zone3_frq50_band500000_S4N4/S{shard_id}N{i}.csv"
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
        file_path = f"../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod2_Num{move_node_num}_Zone3_frq50_band500000/S{shard_id}N{i}specific.csv"
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

def draw_line():
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


        plt.plot(df['epoch'], df['sync_DataSize'], label=labels[i], color=colors[i], markerfacecolor='none',
                 markersize=10)
        # plt.fill_between(df['epoch'], df['syncTime'], color=colors[i], alpha=0.2)
    plt.xlabel('Epoch', fontsize=font_size)
    plt.ylabel('Reconfiguration Data Size (Bytes)', fontsize=font_size)

    plt.xticks(fontsize=font_size)
    plt.yticks(fontsize=font_size)

    # plt.ylim(0,1500)
    plt.legend(fontsize=20, ncol=2, loc='upper left')
    plt.xticks([1, 2, 3, 4, 5], fontsize=font_size)  # 设置 x 轴刻度
    plt.tight_layout()


    # 添加局部放大图
    ax = plt.gca()
    ax_inset = inset_axes(ax, width="40%", height="30%",loc='center right')

    for i, df in enumerate([df0, df3, df1, df2]):
        df = df[:5]
        ax_inset.plot(df['epoch'], df['sync_DataSize'], label=labels[i], color=colors[i], marker=markers[i], markerfacecolor='none', markersize=5)

    ax_inset.set_xlim(1.5, 5.1)
    ax_inset.set_ylim(0, 30000000)

    plt.savefig(f'./reconfiguration_data_line.pdf')
    plt.show()

def draw_bar():
    df0 = get_single_shard_sync_data_of_each_epoch(mod=0, shard_id=0)
    df3 = get_single_shard_sync_data_of_each_epoch(mod=3, shard_id=0)
    df1 = get_single_shard_sync_data_of_each_epoch(mod=1, shard_id=0)
    df2 = get_single_shard_sync_data_of_each_epoch_of_Proposed(0)

    plt.figure(figsize=(7, 6))
    colors = ['#7195C5', '#7262ac', '#01844F', '#E9212C']
    colors = [  (135 / 255, 206 / 255, 163 / 235, 0.5),
              (183 / 255, 183 / 255, 235 / 255, 0.5), (157 / 255, 158 / 255, 163 / 255, 0.5),
              (240 / 255, 155 / 255, 160 / 255, 0.5)]

    labels = ['ETH-full', 'ETH-fast', 'tMPT', 'Proposed']
    hatches = ['/', '\\', '|', '+']
    epochs =min(len(df0['epoch'].unique()),len(df3['epoch'].unique()),len(df1['epoch'].unique()),len(df2['epoch'].unique()))
    epochs=5
    x = range(epochs)
    bar_width = 0.2
    for i, df in enumerate([df0, df3, df1, df2]):
        # 截取前五列
        df['syncTime'] = df['syncTime'] / 1000
        df = df[:epochs]
        padding = 1
        print([p + (bar_width * i) +0.1*i for p in x],[p + (bar_width * i) + 0.1 for p in x])
        plt.bar([p + bar_width * i+0.03*i for p in x], df['sync_DataSize'][:epochs], width=bar_width,
                color=colors[i],
                label=labels[i],

                 edgecolor='black',)

    plt.xlabel('Epoch', fontsize=font_size)
    plt.ylabel('Reconfig. data size (Bytes)', fontsize=font_size)
    plt.xticks(fontsize=font_size)
    plt.yticks(fontsize=font_size)
    plt.legend(fontsize=22, ncol=2, loc='upper left')
    plt.xticks([0, 1, 2, 3, 4],['1','2','3','4','5'] ,fontsize=font_size)  # 设置 x 轴刻度
    plt.tight_layout()
    plt.savefig(f'./pics/reconfiguration_data_bar.pdf')
    plt.show()
draw_bar()
# result/data=1500000_inj=150000_block=1000/Mod1_Num6_Zone3_frq50_band1000000_S4N4/S0N0.csv

import pandas as pd

import matplotlib.pyplot as plt
from mpl_toolkits.axes_grid1.inset_locator import inset_axes, mark_inset

shard_num = 0
move_node_num = 6

import seaborn as sns
import numpy as np


def get_single_shard_sync_data_of_each_epoch(mod=1, shard_id=0):
    dataframes = []
    for i in range(1, move_node_num + 1):
        file_path = f"../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod{mod}_Num6_Zone3_frq50_band500000/S{shard_id}N{i}.csv"
        df = pd.read_csv(file_path)
        dataframes.append(df)

    merged_df = pd.concat(dataframes, axis=0)

    grouped = merged_df.groupby(['epoch'])['beginTime'].transform('min')

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

    grouped = merged_df.groupby(['epoch', 'round'])['beginTime'].transform('min')

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


def get_single_shard_diff_nodes(mod=3, shard_id=0):
    dataframes = []
    for i in range(1, move_node_num + 1):
        file_path = f"../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod{mod}_Num6_Zone3_frq50_band500000/S{shard_id}N{i}.csv"
        df = pd.read_csv(file_path)
        dataframes.append(df)

    merged_df = pd.concat(dataframes, axis=0)

    grouped = merged_df.groupby(['epoch'])['beginTime'].transform('min')

    merged_df['cost'] = merged_df['overTime'] - grouped

    merged_df['cost'] = merged_df['cost'] / 1000

    plt.figure(figsize=(12, 6))

    sns.lineplot(data=merged_df, x='epoch', y='cost')

    plt.title('Cost Distribution by Epoch and Round')
    plt.xlabel('Epoch')
    plt.ylabel('Cost')
    plt.legend(title='Round')
    plt.xticks(rotation=45)
    plt.tight_layout()
    plt.savefig('xxx.png')
    plt.show()


# get_single_shard_diff_nodes(mod=3, shard_id=0)


def get_single_shard_diff_nodes_violinplot(mod=3, shard_id=0):
    dataframes = []

    for i in range(1, move_node_num + 1):
        file_path = f"../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod{mod}_Num6_Zone3_frq50_band500000/S{shard_id}N{i}.csv"
        df = pd.read_csv(file_path)
        dataframes.append(df)

    merged_df = pd.concat(dataframes, axis=0)

    grouped = merged_df.groupby(['epoch'])['beginTime'].transform('min')

    merged_df['cost'] = merged_df['overTime'] - grouped

    merged_df['cost'] = merged_df['cost'] / 1000

    epoch_0_data = merged_df[merged_df['epoch'] == 1]

    # 获取 cost 列并转化为 numpy 数组
    cost_array = epoch_0_data['cost'].to_numpy()
    plt.figure(figsize=(7, 6))
    print(cost_array)
    plt.violinplot(cost_array)
    plt.show()


def get_single_shard_diff_nodes_hot(mod=3, shard_id=0):
    dataframes = []

    if mod != 2:
        for i in range(1, move_node_num + 1):
            file_path = f"../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod{mod}_Num6_Zone3_frq50_band500000/S{shard_id}N{i}.csv"
            df = pd.read_csv(file_path)
            dataframes.append(df)
        merged_df = pd.concat(dataframes, axis=0)
        grouped = merged_df.groupby(['epoch'])['beginTime'].transform('min')
        merged_df['cost'] = merged_df['overTime'] - grouped
        merged_df['cost'] = merged_df['cost'] / 1000

    if mod == 2:
        for i in range(1, move_node_num + 1):
            file_path = f"../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod2_Num{move_node_num}_Zone3_frq50_band500000/S{shard_id}N{i}specific.csv"
            df = pd.read_csv(file_path)
            dataframes.append(df)
        merged_df = pd.concat(dataframes, axis=0)
        merged_df = merged_df[merged_df['round'] == 0]
        grouped = merged_df.groupby(['epoch'])['beginTime'].transform('min')
        merged_df['cost'] = merged_df['overTime'] - grouped
        merged_df['cost'] = merged_df['cost'] / 1000

    merged_df = merged_df[merged_df['epoch'] < 6]

    # 2. 创建透视表，将 'epoch' 和 'nodeID' 转换为行和列，'cost' 作为值
    heatmap_data = merged_df.pivot_table(index='nodeID', columns='epoch', values='cost')

    # 3. 获取热力图数值的最小值和最大值
    min_value = heatmap_data.min().min()
    max_value = heatmap_data.max().max()

    # 3. 绘制热力图
    plt.figure(figsize=(7, 3))
    if mod == 2:
        heatmap = sns.heatmap(heatmap_data, cmap="Blues", cbar=True, vmin=1, vmax=10, annot=True, )
    else:
        heatmap = sns.heatmap(heatmap_data, cmap="Blues", cbar=True, vmin=min_value, vmax=max_value, annot=True,
                              fmt=".2f", )
    # 6. 修改颜色条的刻度
    if mod == 2:
        colorbar = heatmap.collections[0].colorbar
        colorbar.set_ticks([1.5, 2, 2.5, 3, 3.5, 4])

    plt.title("Cost Heatmap")
    plt.xlabel("Epoch")
    plt.ylabel("NodeID")
    plt.show()


# get_single_shard_diff_nodes_hot(0,0)
# get_single_shard_diff_nodes_hot(3,0)
# get_single_shard_diff_nodes_hot(1,0)
# get_single_shard_diff_nodes_hot(2,0)

def get_single_shard_diff_nodes_hot_merge(shard_id=0):
    # labels = ["ETH-full", "ETH-fast", "tMPT", "Proposed"]
    # mods = [0, 3, 1, 2]
    labels = ["ETH-fast", "tMPT", "Proposed"]
    mods = [3, 1, 2]
    fig, axes = plt.subplots(nrows=len(mods), figsize=(7, 2 * len(mods)))
    for idx, mod in enumerate(mods):
        ax = axes[idx]
        dataframes = []
        if mod != 2:
            for i in range(1, move_node_num + 1):
                file_path = f"../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod{mod}_Num6_Zone3_frq50_band500000/S{shard_id}N{i}.csv"
                df = pd.read_csv(file_path)
                dataframes.append(df)
            merged_df = pd.concat(dataframes, axis=0)
            grouped = merged_df.groupby(['epoch'])['beginTime'].transform('min')
            merged_df['cost'] = merged_df['overTime'] - grouped
            merged_df['cost'] = merged_df['cost'] / 1000

        if mod == 2:
            for i in range(1, move_node_num + 1):
                file_path = f"../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod2_Num{move_node_num}_Zone3_frq50_band500000/S{shard_id}N{i}specific.csv"
                df = pd.read_csv(file_path)
                dataframes.append(df)
            merged_df = pd.concat(dataframes, axis=0)
            merged_df = merged_df[merged_df['round'] == 0]
            grouped = merged_df.groupby(['epoch'])['beginTime'].transform('min')
            merged_df['cost'] = merged_df['overTime'] - grouped
            merged_df['cost'] = merged_df['cost'] / 1000

        merged_df = merged_df[merged_df['epoch'] < 6]

        # 2. 创建透视表，将 'epoch' 和 'nodeID' 转换为行和列，'cost' 作为值
        heatmap_data = merged_df.pivot_table(index='nodeID', columns='epoch', values='cost')

        # 3. 获取热力图数值的最小值和最大值
        min_value = heatmap_data.min().min()
        max_value = heatmap_data.max().max()




        if mod == 2:
            heatmap = sns.heatmap(heatmap_data, ax=ax, cmap="Blues", cbar=True, vmin=1, vmax=10, annot=True,
                                  annot_kws={'size': 16})
            colorbar = heatmap.collections[0].colorbar
            colorbar.set_ticks([2, 4, 6, 8, 10])
            # ax.tick_params(axis='x', labelsize=20)
        else:
            if mod == 1:
                heatmap = sns.heatmap(heatmap_data, ax=ax, cmap="Blues", cbar=True, vmin=10,
                                      vmax=80, annot=True, fmt=".0f", xticklabels=False, annot_kws={'size': 16})
                ax.tick_params(axis='x', labelbottom=False)
                colorbar = heatmap.collections[0].colorbar
                colorbar.set_ticks([15, 30, 45, 60, 75])
            else:
                heatmap = sns.heatmap(heatmap_data, ax=ax, cmap="Blues", cbar=True, vmin=min_value, vmax=max_value,
                                      annot=True, fmt=".0f", xticklabels=False, annot_kws={'size': 16}, )
                ax.tick_params(axis='x', labelbottom=False)
                colorbar = heatmap.collections[0].colorbar
                colorbar.set_ticks([30,60, 90, 120,  150])

        ax.set_xlabel("")
        ax.set_ylabel(labels[idx], fontsize=25)
        ax.tick_params(axis='y', labelsize=20)
        ax.tick_params(axis='x', labelsize=20)

    plt.subplots_adjust(wspace=0.3, hspace=0.1)
    plt.xlabel("Epoch", fontsize=25)
    plt.tight_layout()
    plt.savefig("hot_vs_diff_nodes.pdf")

    plt.show()


get_single_shard_diff_nodes_hot_merge(0)

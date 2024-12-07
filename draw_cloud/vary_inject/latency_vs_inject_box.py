import pandas as pd

import matplotlib.pyplot as plt
import numpy as np

shard_num = 16

from brokenaxes import brokenaxes

bandwidth = 1000000
import seaborn as sns
font_size=24
def get_single_shard_sync_data_of_each_epoch(mod=1, shard_id=0, block_size=1000):
    dataframes = []
    node_num = 10
    move_node_num = int(node_num / 2) + 1
    zone_size = int(move_node_num / 2)
    for i in range(1, move_node_num + 1):
        file_path = f"../../ali_result/vary_inject/result/data=3000000_inj={block_size}_block=1000_S16N{node_num}/Mod{mod}_Num{move_node_num}_Zone{zone_size}_frq50_band{bandwidth}/S{shard_id}N{i}.csv"
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
    grouped_df=grouped_df[grouped_df['epoch']<=5]
    return grouped_df


def get_single_shard_sync_data_of_each_epoch_of_Proposed(shard_id=0, block_size=1000):
    dataframes = []
    node_num = 10
    move_node_num = int(node_num / 2) + 1
    zone_size = int(move_node_num / 2)
    for i in range(1, move_node_num + 1):
        file_path = f"../../ali_result/vary_inject/result/data=3000000_inj={block_size}_block=1000_S16N{node_num}/Mod2_Num{move_node_num}_Zone{zone_size}_frq50_band{bandwidth}/S{shard_id}N{i}specific.csv"
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
    grouped_df = grouped_df[grouped_df['epoch'] <= 5]
    return grouped_df


def get_average_sync_time_of_each_epoch(mod=1, shard_id=0, block_size=10):
    if mod == 2:
        data = get_single_shard_sync_data_of_each_epoch_of_Proposed(shard_id, block_size)

        if len(data) > 4:
            data = data[:5]

        return data['syncTime']
    data = get_single_shard_sync_data_of_each_epoch(mod, shard_id, block_size)

    if len(data) > 4:
        data = data[:5]
    return data['syncTime']

def get_latency_data_of_each_shard():
    data_of_all_shard = pd.DataFrame()
    mod = [1,3]
    blcokszie = [1000, 2000, 3000,4000,5000]

    for m in mod:
        for b in blcokszie:
            for i in range(shard_num):
                data = get_single_shard_sync_data_of_each_epoch(m, i, b)
                data['blocksize'] = b
                data['mod'] = m
                data_of_all_shard = pd.concat([data_of_all_shard, data], axis=0)

    for b in blcokszie:
        for i in range(shard_num):
            data = get_single_shard_sync_data_of_each_epoch_of_Proposed( i, b)
            data['blocksize'] = b
            data['mod'] = 2
            data_of_all_shard = pd.concat([data_of_all_shard, data], axis=0)

    data_of_all_shard['syncTime']=data_of_all_shard['syncTime']/1000
    # 确保数据按照 'blocksize' 和 'mod' 分组
    grouped_data = data_of_all_shard.groupby(['blocksize', 'mod'])

    # 初始化画布
    plt.figure(figsize=(7, 5))

    # 提取 unique blocksize 和 mod
    blocksizes = sorted(data_of_all_shard['blocksize'].unique())
    mods = [3,1,2]

    # 定义颜色与偏移
    colors = [
        (183 / 255, 183 / 255, 235 / 255, 0.5), (157 / 255, 158 / 255, 163 / 255, 0.5),
        (240 / 255, 155 / 255, 160 / 255, 0.5)]
    offset = 0.25
    labels = ['ETH-fast', 'tMPT', 'Proposed']

    # 开始绘图
    for i, blocksize in enumerate(blocksizes):
        for j, mod in enumerate(mods):
            x=j-1
            # 获取当前组合的 syncTime 数据
            subset = data_of_all_shard[(data_of_all_shard['blocksize'] == blocksize) & (data_of_all_shard['mod'] == mod)]['syncTime']
            # 绘制箱线图
            if not subset.empty:
                plt.boxplot(
                    subset,
                    positions=[i + x * offset],  # 偏移以区分不同 mod
                    widths=0.2,
                    patch_artist=True,
                    boxprops=dict(facecolor=colors[j]),

                )

    # 添加图例
    legend_patches = [plt.Line2D([0], [0], color=color, lw=8, label=label) for color, label in zip(colors, labels)]
    plt.legend(handles=legend_patches, fontsize=22, )

    # 添加标签和标题
    plt.xticks(ticks=range(len(blocksizes)), labels=blocksizes,fontsize=font_size)
    plt.yticks(fontsize=font_size)
    plt.xlabel("Injection rate (TXs/Sec.)", fontsize=font_size)
    plt.ylabel('Reconfig. latency (Sec.)', fontsize=font_size)

    plt.yticks(fontsize=font_size)
    plt.tight_layout()
    # 显示图形
    plt.savefig(f'../pics/shuffe_latency_vary_inject.pdf')
    plt.show()

get_latency_data_of_each_shard()


def get_latency_data_of_each_shard_error_bar():
    data_of_all_shard = pd.DataFrame()
    mod = [1,3]
    blcokszie = [500, 1000, 1500,2000]

    colors = [
        (183 / 255, 183 / 255, 235 / 255, 0.5), (157 / 255, 158 / 255, 163 / 255, 0.5),
        (240 / 255, 155 / 255, 160 / 255, 0.5)]
    mods = [3, 1, 2]

    for m in mod:
        for b in blcokszie:
            for i in range(shard_num):
                data = get_single_shard_sync_data_of_each_epoch(m, i, b)
                data['blocksize'] = b
                data['mod'] = m
                data_of_all_shard = pd.concat([data_of_all_shard, data], axis=0)

    for b in blcokszie:
        for i in range(shard_num):
            data = get_single_shard_sync_data_of_each_epoch_of_Proposed( i, b)
            data['blocksize'] = b
            data['mod'] = 2
            data_of_all_shard = pd.concat([data_of_all_shard, data], axis=0)

    # 确保数据按要求处理
    data_of_all_shard['blocksize'] = data_of_all_shard['blocksize'].astype(str)  # 确保 blocksize 为分类变量
    data_of_all_shard['mod'] = data_of_all_shard['mod'].astype(str)  # 确保 mod 为分类变量

    # 初始化图形
    plt.figure(figsize=(12, 6))
    sns.set(style="whitegrid")

    # 绘制箱线图
    sns.boxplot(
        x='blocksize',
        y='syncTime',
        hue='mod',
        data=data_of_all_shard,
        showmeans=True,
        meanprops={"marker": "o", "markerfacecolor": "red", "markeredgecolor": "black"}
    )

    # 图表增强
    plt.title("SyncTime Distribution by Blocksize and Mod")
    plt.xlabel("Blocksize")
    plt.ylabel("SyncTime")
    plt.legend(title="Mod")
    plt.xticks(rotation=45)
    plt.tight_layout()

    # 显示图形
    plt.show()







# get_latency_data_of_each_shard_error_bar()



def get_latency_data_of_each_shard_error_bar_new():
    data_of_all_shard = pd.DataFrame()
    mod = [1,3]
    blcokszie = [500, 1000, 1500,2000]

    colors = [
        (183 / 255, 183 / 255, 235 / 255, 0.5), (157 / 255, 158 / 255, 163 / 255, 0.5),
        (240 / 255, 155 / 255, 160 / 255, 0.5)]
    mods = [3, 1, 2]

    for m in mod:
        for b in blcokszie:
            for i in range(shard_num):
                data = get_single_shard_sync_data_of_each_epoch(m, i, b)
                data['blocksize'] = b
                data['mod'] = m
                data_of_all_shard = pd.concat([data_of_all_shard, data], axis=0)

    for b in blcokszie:
        for i in range(shard_num):
            data = get_single_shard_sync_data_of_each_epoch_of_Proposed( i, b)
            data['blocksize'] = b
            data['mod'] = 2
            data_of_all_shard = pd.concat([data_of_all_shard, data], axis=0)

        # 按 blocksize 和 mod 分组
    # 按 blocksize 和 mod 分组，计算 syncTime 的均值和标准差
    grouped = data_of_all_shard.groupby(['blocksize', 'mod'])['syncTime'].agg(['mean', 'std']).reset_index()

    # 获取所有的 mods 和 blocksizes
    mods = [3,1,2]
    blocksizes = grouped['blocksize'].unique()

    # 绘图
    plt.figure(figsize=(7, 6))
    for idx, mod in enumerate(mods) :
        mod_data = grouped[grouped['mod'] == mod]
        plt.errorbar(mod_data['blocksize'], mod_data['mean'], yerr=mod_data['std'], label=f'Mod: {mod}', capsize=5,
                     marker='o',colors=colors[idx])

    # 设置图例和标签
    plt.title('Sync Time vs Blocksize for Different Mods')
    plt.xlabel('Blocksize')
    plt.ylabel('Sync Time (mean ± std)')
    plt.legend()
    plt.tight_layout()

    # 显示图形
    plt.show()

# get_latency_data_of_each_shard_error_bar_new()
shard_num = 0
move_node_num = 6


import pandas as pd

import matplotlib.pyplot as plt



def get_single_shard_sync_data_of_each_epoch_of_Proposed(mod=1, shard_id=0,bandwidth=500000):
    dataframes = []
    for i in range(1, move_node_num + 1):
        file_path = f"./result/data=1500000_inj=150000_block=1000/Mod{mod}_Num{move_node_num}_Zone3_frq50_band{bandwidth}_S4N4/S{shard_id}N{i}specific.csv"
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
    return grouped_df


def get_time_vs_sync_data_size(mod, shard_id,bandwidth=500000):
    t = get_single_shard_sync_data_of_each_epoch_of_Proposed(mod, shard_id,bandwidth)
    t['adjusted_beginTime'] = t['min_beginTime'] - t['min_beginTime'].min()
    t['adjusted_beginTime'] = t['adjusted_beginTime'] / 1000
    t['cumulative_sync_DataSize'] = t['sync_DataSize'].cumsum()
    t = t[t['round'] < 50]
    t = t[t['epoch'].isin([1])]
    return t




#

def real_time_confirm_tx():
    file_path= 'result/data=1500000_inj=150000_block=1000/Mod2_Num6_Zone3_frq50_band500000_S4N4/pbft_shardNum=8/Shard00.csv'
    df = pd.read_csv(file_path)
    t = get_time_vs_sync_data_size(2, 0, 500000)
    begin_time = t['min_beginTime'].min()
    end_time = t['max_overTime'].max()

    df = df[df['TimeStamp - Commit (unixMill)'] > begin_time]
    df = df[df['TimeStamp - Propose (unixMill)'] < end_time]

    df['confirm_time']=df['TimeStamp - Commit (unixMill)']-df['TimeStamp - Commit (unixMill)'].min()



    df['confirm_time'] = df['confirm_time']/1000
    df['confirmed_tx'] = df['# of all Txs in this block'].cumsum()

    plt.plot(df['confirm_time'],df['confirmed_tx'])
    plt.show()
from matplotlib.gridspec import GridSpec
import matplotlib.ticker as ticker
def draw_time_vs_sync_data_size():
    fig = plt.figure(figsize=(7, 6))
    gs = GridSpec(2, 1, height_ratios=[1, 5])  # Adjust the height ratios as needed

    ax1 = fig.add_subplot(gs[0])
    ax2 = fig.add_subplot(gs[1])

    lines_top = []
    lines_bottom = []
    for i in range(0, 8):
        t = get_time_vs_sync_data_size(2, i, 500000)
        line, =ax2.plot(t['adjusted_beginTime'], t['cumulative_sync_DataSize'], label=f'Shard #{i}')
        if i < 4:
            lines_top.append(line)  # 前 4 条线放在左上角
        else:
            lines_bottom.append(line)  # 后 4 条线放在右下角


    file_path = 'result/data=1500000_inj=150000_block=1000/Mod2_Num6_Zone3_frq50_band500000_S4N4/pbft_shardNum=8/Shard00.csv'
    df = pd.read_csv(file_path)
    t = get_time_vs_sync_data_size(2, 0, 500000)
    begin_time = t['min_beginTime'].min()
    end_time = t['max_overTime'].max()

    df = df[df['TimeStamp - Commit (unixMill)'] > begin_time]
    df = df[df['TimeStamp - Propose (unixMill)'] < end_time]

    df['confirm_time'] = df['TimeStamp - Commit (unixMill)'] - df['TimeStamp - Commit (unixMill)'].min()

    df['confirm_time'] = df['confirm_time'] / 1000
    df['confirmed_tx'] = df['# of all Txs in this block'].cumsum()





    ax1.plot(df['confirm_time'], df['confirmed_tx'], label='Confirmed Tx', marker='o', markersize=3, markerfacecolor='none',markevery=1)

    ax1.yaxis.set_major_formatter(ticker.ScalarFormatter(useMathText=True))
    ax1.ticklabel_format(style='sci', axis='y', scilimits=(0, 0))



    ax1.tick_params(axis='both', which='major', labelsize=16)
    ax2.tick_params(axis='both', which='major', labelsize=16)




    plt.xlabel('Time (s)', fontsize=16)
    plt.ylabel('                            Sync Data Size (Bytes)         Confirmed Txs', fontsize=16)

    # 创建左上角的图例
    legend_top = ax2.legend(lines_top, [f'Shard #{i}' for i in range(0, 4)], loc="upper left", fontsize=16)

    # 创建右下角的图例
    legend_bottom = ax2.legend(lines_bottom, [f'Shard #{i}' for i in range(4, 8)], loc="lower right", fontsize=16)

    # 确保第一个图例不会被第二个覆盖
    ax2.add_artist(legend_top)

    plt.tight_layout()
    plt.savefig('./pics/real_time_data_size.pdf')
    plt.show()

draw_time_vs_sync_data_size()
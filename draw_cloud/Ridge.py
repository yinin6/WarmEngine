import joypy
import pandas as pd
import numpy as np
from matplotlib import pyplot as plt
from matplotlib import cm
from sklearn.datasets import load_iris

move_node_num = 6
mod = 1
data = []


def get_data(mod):
    merged_df = []
    dataframes = []
    for s in range(0, 16):
        for i in range(1, move_node_num + 1):
            file_path = f"../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod{mod}_Num6_Zone3_frq50_band500000/S{s}N{i}.csv"
            df = pd.read_csv(file_path)
            dataframes.append(df)
            merged_df = pd.concat(dataframes, axis=0)
    grouped = merged_df.groupby(['epoch'])['beginTime'].transform('min')
    merged_df['cost'] = merged_df['overTime'] - grouped
    merged_df['cost'] = merged_df['cost'] / 1000
    merged_df['mod'] = mod

    # 获取唯一的 epoch 值
    epochs = df['epoch'].unique()

    # 切分数据并将每个 epoch 的 cost 存入新列
    new_df = pd.DataFrame()

    for epoch in epochs:
        # 获取每个 epoch 下的 cost 值，并设置列名
        cost_column = merged_df[merged_df['epoch'] == epoch]['cost'].reset_index(drop=True)
        new_df[f'Epoch {epoch}'] = cost_column
    if mod == 1:
        new_df['mod'] = 'tMPT'
    else:
        new_df['mod'] = 'ETH-fast'

    data.append(new_df)


def get_data_proposed(mod=2):
    merged_df = []
    dataframes = []
    for s in range(0, 16):
        for i in range(1, move_node_num + 1):
            file_path = f"../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod2_Num{move_node_num}_Zone3_frq50_band500000/S{s}N{i}specific.csv"
            df = pd.read_csv(file_path)
            dataframes.append(df)
            merged_df = pd.concat(dataframes, axis=0)
    merged_df = merged_df[merged_df['round'] == 0]
    grouped = merged_df.groupby(['epoch'])['beginTime'].transform('min')
    merged_df['cost'] = merged_df['overTime'] - grouped
    merged_df['cost'] = merged_df['cost'] / 1000

    # 获取唯一的 epoch 值
    epochs = df['epoch'].unique()

    # 切分数据并将每个 epoch 的 cost 存入新列
    new_df = pd.DataFrame()

    for epoch in epochs:
        # 获取每个 epoch 下的 cost 值，并设置列名
        cost_column = merged_df[merged_df['epoch'] == epoch]['cost'].reset_index(drop=True)
        new_df[f'Epoch {epoch}'] = cost_column
    new_df['mod'] = ' Proposed'

    data.append(new_df)


get_data_proposed()
get_data(3)
get_data(1)
# get_data(0)


data = pd.concat(data, axis=0)

fig, axes = joypy.joyplot(data, by="mod",
                          column=['Epoch 1', 'Epoch 2', 'Epoch 3', 'Epoch 4', 'Epoch 5'],
                          color=[(155 / 255, 187 / 255, 225 / 255, 0.5), (157 / 255, 158 / 255, 163 / 255, 0.5),
                                 (183 / 255, 183 / 255, 235 / 255, 0.5), (234 / 255, 184 / 255, 131 / 255, 0.5),
                                 (240 / 255, 155 / 255, 160 / 255, 0.5)],
                          legend=True, ylim='own')
# 调整 x 和 y 轴的边框
for ax in axes:
    ax.spines['top'].set_visible(True)  # 如果需要显示顶部边框
    ax.spines['right'].set_visible(True)  # 如果需要显示右侧边框
    ax.spines['bottom'].set_visible(True)
    ax.spines['left'].set_visible(True)

    # 设置边框颜色、线宽
    for spine in ax.spines.values():
        spine.set_edgecolor('black')
        spine.set_linewidth(1)

# Enable y-axis tick values
for ax in axes:
    # Make y-axis visible
    ax.tick_params(axis='y', which='both', labelsize=8)  # Customize y-axis tick labels size

fig.set_size_inches(7, 6)

for ax in axes:
    ax.tick_params(axis='x', labelsize=20)  # 设置刻度的字体大小
    ax.tick_params(axis='y', labelsize=20)

for idx, ax in enumerate(axes):
    if idx==3:
        continue
    y_min, y_max = ax.get_ylim()
    ax.tick_params(pad=4)
    ax.tick_params(axis='y', which='both', direction='in', length=5, width=1, colors='black')

    print(idx,y_min, y_max)
    ax.yaxis.set_visible(True)
    ax.set_yticks([0, y_max])



    if idx == 2:
        ax.set_yticklabels(["0", "0.03"])
    if idx == 1:
        ax.set_yticklabels(["0", "0.03"])
    if idx == 0:
        ax.set_yticklabels(["0", "1"])
        ax.legend(fontsize=16, ncol=2)

fig.text(x=0.18, y=0.58, s="ETH-fast", rotation=90, va='center', ha='center', fontsize=20)
fig.text(x=0.18, y=0.3, s="tMPT", rotation=90, va='center', ha='center', fontsize=20)
fig.text(x=0.18, y=0.849, s="Proposed", rotation=90, va='center', ha='center', fontsize=20)

fig.text(x=0.03, y=0.55, s="PDF", rotation=90, va='center', ha='center', fontsize=25)

plt.xlabel('Reconfiguration latency (Sec.)', fontsize=25)  # 添加x轴名称
plt.yticks([0, 1, 2, 3, 4])
plt.tight_layout()

plt.savefig('reconfiguration_time_ridge.pdf')
plt.show()

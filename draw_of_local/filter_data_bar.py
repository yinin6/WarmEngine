


import pandas as pd
import matplotlib.pyplot as plt
from matplotlib.ticker import ScalarFormatter

# 读取 CSV 文件
file_path1= "./result/data=1500000_inj=10000_block=1000_S8N10/Mod2_Num6_Zone3_frq50_band5000000/S0N1specific.csv"

df1 = pd.read_csv(file_path1)


file_path2= "./result/data=1500000_inj=10000_block=1000_S8N10/Mod2_Num6_Zone3_frq50_band5000000/S0N7filterData.csv"
df2 = pd.read_csv(file_path2)

df2 = (
    df2.groupby(['epoch', 'round'], as_index=False)  # 分组
    .agg(raw_data_size=('rawDataSize', 'min'))     # 计算均值并重命名
)


data = pd.merge(df1, df2, on=['epoch', 'round'], how='inner')  # inner 合并方式


data=data[data['epoch']<=5]

# 数据处理：按 epoch 分组并计算 account 和 filterAccount 的总和
grouped_data = data.groupby('epoch').agg({
    'raw_data_size': 'sum',
    'stateValueSize': 'sum'
}).reset_index()
# 设置 Y 轴为科学计数法
print(grouped_data.columns)


def draw_singel_bar():
    fig, ax = plt.subplots(figsize=(7, 6))


    # 创建柱状图
    bars=ax.bar(grouped_data['epoch'], grouped_data['raw_data_size'],bottom=grouped_data['stateValueSize'],  label='Reduced repeat retrieved data', color='#d8f0f2' , edgecolor='black' ,hatch='//')
    for bar in bars:
        bar.set_edgecolor('black')  # 边框颜色
        bar.set_linestyle((0, (5, 5)))  # 虚线样式 (长度为5的实线 + 长度为5的空白)
        bar.set_linewidth(1)  # 边框宽度
    ax.bar(grouped_data['epoch'], grouped_data['stateValueSize'], label='The actual retrieved data', color='skyblue' , edgecolor='black')

    # 添加图例和标签
    ax.set_xlabel('Epoch', fontsize=20)
    ax.set_ylabel('The total of retrieved data (bytes)', fontsize=20)
    ax.legend(fontsize=20)
    ax.yaxis.set_major_formatter(ScalarFormatter(useMathText=True))
    ax.ticklabel_format(axis='y', style='sci', scilimits=(0, 0))
    # 绘制堆叠柱状图
    ax.tick_params(axis='x', labelsize=20)
    ax.tick_params(axis='y', labelsize=20)
    # 显示图形
    plt.xticks(grouped_data['epoch'])
    plt.tight_layout()
    plt.savefig('./pics/preload_filter_data.pdf')
    plt.show()


def draw_duble_bar(df):
    x = df['epoch']
    width = 0.4  # 设置柱状图宽度

    # 创建图表
    fig, ax = plt.subplots(figsize=(7, 6))

    # 绘制两组柱状图
    bars1 = ax.bar(x - width / 2, df['raw_data_size']/1024/1024, width-0.05, label='tMPT', color='orange', edgecolor='black')
    bars2 = ax.bar(x + width / 2, df['stateValueSize']/1024/1024, width-0.05, label='Proposed', color='skyblue', edgecolor='black')


    # 添加具体的值到柱状图上
    for bar in bars1:
        height = bar.get_height()
        ax.text(bar.get_x() + bar.get_width() / 2, height, f'{height:.1f}', ha='center', va='bottom', fontsize=19)

    for bar in bars2:
        height = bar.get_height()
        ax.text(bar.get_x() + bar.get_width() / 2, height, f'{height:.1f}', ha='center', va='bottom', fontsize=19)


    ax.legend(fontsize=25)
    # 设置图表标题和轴标签
    ax.set_xlabel('Epoch')
    ax.set_ylabel('Size')
    ax.set_xticks(x)
    ax.set_xticklabels(x)
    ax.tick_params(axis='x', labelsize=25)
    ax.tick_params(axis='y', labelsize=25)
    ax.set_xlabel('Epoch', fontsize=25)
    ax.set_ylabel('Retrieved data size (MBs)', fontsize=25)
    plt.tight_layout()
    plt.savefig('./pics/preload_filter_data.pdf')



    # 显示图表
    plt.show()

draw_duble_bar(grouped_data)
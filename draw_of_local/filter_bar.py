


import pandas as pd
import matplotlib.pyplot as plt
from matplotlib.ticker import ScalarFormatter

# 读取 CSV 文件
file_path = './result/Mod2_Num6_Zone3_frq50_band5000000/S0N1filter.csv'  # 替换为您的 CSV 文件路径
data = pd.read_csv(file_path)

data=data[data['epoch']<=5]

# 数据处理：按 epoch 分组并计算 account 和 filterAccount 的总和
grouped_data = data.groupby('epoch').agg({
    'account': 'sum',
    'filterAccount': 'sum'
}).reset_index()
# 设置 Y 轴为科学计数法

fig, ax = plt.subplots(figsize=(7, 6))


# 创建柱状图
bars=ax.bar(grouped_data['epoch'], grouped_data['account'],bottom=grouped_data['filterAccount'],  label='New created accounts', color='#d8f0f2' , edgecolor='black' ,hatch='//')
for bar in bars:
    bar.set_edgecolor('black')  # 边框颜色
    bar.set_linestyle((0, (5, 5)))  # 虚线样式 (长度为5的实线 + 长度为5的空白)
    bar.set_linewidth(1)  # 边框宽度
ax.bar(grouped_data['epoch'], grouped_data['filterAccount'], label='Accounts to be retrieved', color='skyblue' , edgecolor='black')

# 添加图例和标签
ax.set_xlabel('Epoch', fontsize=20)
ax.set_ylabel('Num. of accounts to be retrieved', fontsize=20)
ax.legend(fontsize=20)
ax.yaxis.set_major_formatter(ScalarFormatter(useMathText=True))
ax.ticklabel_format(axis='y', style='sci', scilimits=(0, 0))
# 绘制堆叠柱状图
ax.tick_params(axis='x', labelsize=20)
ax.tick_params(axis='y', labelsize=20)
# 显示图形
plt.xticks(grouped_data['epoch'])
plt.tight_layout()
plt.savefig('./pics/preload_filter.pdf')
plt.show()

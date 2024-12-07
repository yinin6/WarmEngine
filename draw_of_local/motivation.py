

path='result/export-PendingQueue.csv'

from matplotlib.ticker import ScalarFormatter

import pandas as pd
import matplotlib.pyplot as plt


# 读取 CSV 文件
file_path = "result/export-PendingQueue.csv"
df = pd.read_csv(file_path)


# 11/12/2024 3:06:00 AM
# 将 "Date(UTC)" 列转换为日期时间格式
df["Date(UTC)"] = pd.to_datetime(df["Date(UTC)"], errors="coerce")

# 筛选出 2024/11/9 的数据
df_filtered = df[df["Date(UTC)"].dt.date == pd.to_datetime("2024-11-09").date()]

# 绘制图表
fig, ax = plt.subplots(figsize=(7, 6))
plt.plot(df_filtered["Date(UTC)"], df_filtered["Value"], linestyle='-', color='b')
plt.fill_between(df_filtered["Date(UTC)"], df_filtered["Value"], color='b', alpha=0.05)

# 将 y 轴调整为科学计数法
ax.yaxis.set_major_formatter(ScalarFormatter(useMathText=True))
ax.ticklabel_format(style='sci', axis='y', scilimits=(0,0))
# 设置 x 和 y 轴刻度字体大小为 16
ax.tick_params(axis='x', labelsize=20)
ax.tick_params(axis='y', labelsize=20)


plt.ylabel("Ethereum Pending Transactions",fontsize=20)
plt.xticks(rotation=30)
plt.ylim(0,200000)

plt.tight_layout()

plt.savefig(f'./pics/PendingTx.pdf')

# 显示图形
plt.show()

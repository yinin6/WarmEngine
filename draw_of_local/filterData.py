import pandas as pd
import matplotlib.pyplot as plt

file_path1= "./result/data=1500000_inj=10000_block=1000_S8N10/Mod2_Num6_Zone3_frq50_band5000000/S0N1specific.csv"

df1 = pd.read_csv(file_path1)


file_path2= "./result/data=1500000_inj=10000_block=1000_S8N10/Mod2_Num6_Zone3_frq50_band5000000/S0N7filterData.csv"
df2 = pd.read_csv(file_path2)

df2 = (
    df2.groupby(['epoch', 'round'], as_index=False)  # 分组
    .agg(raw_data_size=('rawDataSize', 'min'))     # 计算均值并重命名
)


merged_df = pd.merge(df1, df2, on=['epoch', 'round'], how='inner')  # inner 合并方式


# 绘制折线图
plt.figure(figsize=(7, 6))

# 折线图1：raw_data_size
plt.plot(merged_df.index, merged_df['raw_data_size'], marker='o', label='tMPT',color='orange')

# 折线图2：stateValueSize
plt.plot(merged_df.index, merged_df['stateValueSize'], marker='s', label='Proposed',color='skyblue')



merged_df['reduction_percentage'] = (1 - merged_df['stateValueSize'] / merged_df['raw_data_size']) * 100
average_reduction = merged_df['reduction_percentage'].mean()
print(f"Average Reduction Percentage: {average_reduction:.2f}%")



from matplotlib.ticker import ScalarFormatter
# 图形美化
plt.gca().yaxis.set_major_formatter(ScalarFormatter(useMathText=True))
plt.ticklabel_format(axis='y', style='sci', scilimits=(0, 0))


plt.xlabel('Index of round', fontsize=25)
plt.ylabel('Fetched data size (Bytes)', fontsize=25)
plt.legend(fontsize=25)

plt.xticks(fontsize=25)
plt.yticks(fontsize=25)
plt.tight_layout()
plt.savefig('./pics/preload_filter_data_each_round.pdf')
# 显示图形
plt.show()
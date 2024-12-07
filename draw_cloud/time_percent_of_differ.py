import pandas as pd
import matplotlib.pyplot as plt

file_path = "../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod2_Num6_Zone3_frq50_band500000/S0N1specific.csv"

blcok_path = '../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod2_Num6_Zone3_frq50_band500000/pbft_shardNum=16/Shard00.csv'

df = pd.read_csv(file_path)
print(df.head)

df = df[df['round'] < 50]

df = df[df['epoch'] == 3]

# 找到所有记录中最小的beginTime
min_begin_time = df['beginTime'].min()
df['begin_diff'] = df['beginTime'] - min_begin_time
df['over_diff'] = df['overTime'] - min_begin_time


block_df = pd.read_csv(blcok_path)
block_df = block_df[block_df['TimeStamp - Propose (unixMill)'] > min_begin_time]
block_df = block_df[block_df['Block Height'] <200]
block_df['begin_diff'] = block_df['TimeStamp - Propose (unixMill)'] - min_begin_time
block_df['over_diff'] = block_df['TimeStamp - Commit (unixMill)'] - min_begin_time

block_df['begin_diff']=block_df['begin_diff']/1000
block_df['over_diff']=block_df['over_diff']/1000

df['begin_diff']=df['begin_diff']/1000
df['over_diff']=df['over_diff']/1000

# 按照epoch和round排序
df.sort_values(by=['epoch', 'round'], inplace=True)

# 获取唯一的epoch列表
epochs = df['epoch'].unique()

# 创建一个新的图形
plt.figure(figsize=(10, 6))

# 遍历每个epoch
for i, epoch in enumerate(epochs):
    # 过滤出当前epoch的数据
    epoch_data = df[df['epoch'] == epoch]

    # 初始化当前epoch的开始位置
    start_pos = 0

    # 遍历每个round
    for _, row in epoch_data.iterrows():
        # 计算每个部分的宽度
        width = row['over_diff'] - row['begin_diff']
        print(i, row['begin_diff'],row['over_diff'])
        # 绘制条形图的一段
        plt.barh(i, width, left=row['begin_diff'], color='red')


for _, row in block_df.iterrows():
    # 计算每个部分的宽度
    width = row['over_diff'] - row['begin_diff']
    # 绘制条形图的一段
    plt.barh(1, width, left=row['begin_diff'], color='blue', alpha=0.7)





# 设置y轴标签
plt.yticks(range(len(epochs)), epochs)

# 添加标题和标签
plt.title('Epochs and Rounds Timeline')
plt.xlabel('Time Difference from Minimum Begin Time')
plt.ylabel('Epoch')

# 显示图表
plt.show()

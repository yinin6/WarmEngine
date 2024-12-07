import pandas as pd
import matplotlib.pyplot as plt

shard_num=16

mod=0


data=[]
# result/data=1500000_inj=150000_block=1000/Mod1_Num6_Zone3_frq50_band500000_S4N4/pbft_shardNum=8/epochDatil.csvS00.csv
for mod in [0,3,1,2]:
    tps=None
    for i in range(shard_num):
        file_path=f"../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod{mod}_Num6_Zone3_frq50_band500000/pbft_shardNum={shard_num}/epochDatil.csvS{i}0.csv"
        df = pd.read_csv(file_path)
        if tps is None:
            tps = df['tps'].tolist()
        else:
            tps = [x + y for x, y in zip(tps, df['tps'].tolist())]
    tps = tps[:5]
    mean = sum(tps) / len(tps)
    print(mean)
    data.append(tps)



    print(data)

plt.figure(figsize=(7, 6))
box = plt.boxplot(data, patch_artist=True)
# plt.xlabel('TPS')
plt.ylabel('Throughput (TXs/Sec.)',fontsize=25)
colors = ['skyblue','#B7B7EB',  '#9D9EA3', '#F09BA0']
colors = [  (135 / 255, 206 / 255, 163 / 235, 0.5),
              (183 / 255, 183 / 255, 235 / 255, 0.5), (157 / 255, 158 / 255, 163 / 255, 0.5),
              (240 / 255, 155 / 255, 160 / 255, 0.5)]
hatches = ['/', '\\', '|', '+']

hatches = [' ', ' ', ' ', ' ']


for patch, color, hatch in zip(box['boxes'], colors, hatches):
    patch.set_facecolor(color)
    patch.set_hatch(hatch)
# plt.title('Box Plot of TPS',fontsize=16)

import matplotlib.ticker as mticker
# 设置y轴为科学计数法
plt.gca().yaxis.set_major_formatter(mticker.ScalarFormatter(useMathText=True))
plt.gca().yaxis.get_major_formatter().set_scientific(True)
plt.gca().yaxis.get_major_formatter().set_powerlimits((0, 0))

plt.xticks([1, 2, 3, 4], ['ETH-full','  ETH-fast','tMPT' ,'Proposed'],fontsize=24)
plt.yticks(fontsize=25)
plt.tight_layout()
plt.savefig(f'./pics/tps_vs_epoch_box.pdf')
plt.show()



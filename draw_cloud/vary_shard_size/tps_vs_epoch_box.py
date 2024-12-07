import pandas as pd
import matplotlib.pyplot as plt

shard_num=16
node_num_in_shard = 10
bandwidth = 2000000
move_node_num = 6
zone=int(move_node_num/2)

mod=0


data=[]
for mod in [3,1,2]:
    tps=None
    for i in range(shard_num):
        file_path=f"../../ali_result/result/data=3000000_inj=10000_block=1000_S16N{node_num_in_shard}/Mod{mod}_Num{move_node_num}_Zone{zone}_frq50_band{bandwidth}/pbft_shardNum={shard_num}/epochDatil.csvS{i}0.csv"
        df = pd.read_csv(file_path)
        if tps is None:
            tps = df['tps'].tolist()
        else:
            tps = [x + y for x, y in zip(tps, df['tps'].tolist())]
    tps = tps[:5]
    data.append(tps)



    print(data)

plt.figure(figsize=(7, 6))
box = plt.boxplot(data, patch_artist=True)
# plt.xlabel('TPS')
plt.ylabel('TPS (TXs/sec.)',fontsize=20)
colors = ['skyblue', 'lightgreen', 'lightgray', 'lightpink']
hatches = ['/', '\\', '|', '+']



for patch, color, hatch in zip(box['boxes'], colors, hatches):
    patch.set_facecolor(color)
    patch.set_hatch(hatch)
# plt.title('Box Plot of TPS',fontsize=16)

import matplotlib.ticker as mticker
# 设置y轴为科学计数法
plt.gca().yaxis.set_major_formatter(mticker.ScalarFormatter(useMathText=True))
plt.gca().yaxis.get_major_formatter().set_scientific(True)
plt.gca().yaxis.get_major_formatter().set_powerlimits((0, 0))

plt.xticks([1, 2, 3, 4], ['ETH-full','ETH-fast','tMPT' ,'Proposed'],fontsize=20)
plt.yticks(fontsize=20)
plt.tight_layout()
plt.savefig(f'./tps_vs_epoch_box.pdf')
plt.show()



import pandas as pd
import matplotlib.pyplot as plt

shard_num=8

mod=0


data=[]
# result/data=1500000_inj=150000_block=1000/Mod1_Num6_Zone3_frq50_band500000_S4N4/pbft_shardNum=8/epochDatil.csvS00.csv
for mod in [0,3,1,2]:
    tps=None
    for i in range(shard_num):
        file_path=f"./result/data=1500000_inj=150000_block=1000/Mod{mod}_Num6_Zone3_frq50_band500000_S4N4/pbft_shardNum={shard_num}/epochDatil.csvS{i}0.csv"
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
plt.ylabel('TPS (Txs/sec)',fontsize=16)
colors = ['skyblue', 'lightgreen', 'lightgray', 'lightpink']
hatches = ['/', '\\', '|', '+']

for patch, color, hatch in zip(box['boxes'], colors, hatches):
    patch.set_facecolor(color)
    patch.set_hatch(hatch)
# plt.title('Box Plot of TPS',fontsize=16)

plt.xticks([1, 2, 3, 4], ['ETH-full','ETH-fast','tMPT' ,'Proposed'],fontsize=16)
plt.yticks(fontsize=16)
plt.tight_layout()
plt.savefig(f'./pics/tps_vs_epoch_box.pdf')
plt.show()



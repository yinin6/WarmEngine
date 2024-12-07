import pandas as pd
import matplotlib.pyplot as plt

shard_num = 16

mod = 0


data = []

for mod in [0, 3, 1, 2]:
    tps = None
    for i in range(shard_num):
        file_path=f"../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod{mod}_Num6_Zone3_frq50_band500000/pbft_shardNum={shard_num}/epochDatil.csvS{i}0.csv"
        df = pd.read_csv(file_path)
        if tps is None:
            tps = df['tps'].tolist()
        else:
            tps = [x + y for x, y in zip(tps, df['tps'].tolist())]
    tps = tps[:5]
    print(len(tps))
    data.append(tps)


# Plotting the line chart
plt.figure(figsize=(7, 6))
colors = ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728']  # Scientific color palette
labels = ['ETH-full','ETH-fast','tMPT' ,'Proposed']
markers = ['o', 's', 'D', '^']
for i, tps in enumerate(data):
    plt.plot(tps, label=labels[i], color=colors[i],marker=markers[i], markerfacecolor='none', markersize=10)

plt.xlabel('Epoch', fontsize=16)
plt.ylabel('TPS (Txs/sec)', fontsize=16)
# plt.title('TPS vs Epoch Line Chart', fontsize=16)
plt.xticks([0, 1, 2, 3, 4], ['1','2','3' ,'4','5'],fontsize=16)
plt.yticks(fontsize=16)
plt.legend(fontsize=16, ncol=2,loc='upper left', bbox_to_anchor=(0.2, 0.45))
plt.tight_layout()
plt.savefig(f'./pics/tps_vs_epoch_line.pdf')
plt.show()
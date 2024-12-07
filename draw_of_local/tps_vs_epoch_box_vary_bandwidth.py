import pandas as pd
import matplotlib.pyplot as plt
shard_num=8


def get_diff_data(bandwidth):
    data = []
    for mod in [0, 3, 1, 2]:
        tps = None
        for i in range(shard_num):
            file_path = f"./result/data=1500000_inj=150000_block=1000/Mod{mod}_Num6_Zone3_frq50_band{bandwidth}_S4N4/pbft_shardNum={shard_num}/epochDatil.csvS{i}0.csv"
            df = pd.read_csv(file_path)
            if tps is None:
                tps = df['tps'].tolist()
            else:
                tps = [x + y for x, y in zip(tps, df['tps'].tolist())]
        tps = tps[:5]
        data.append(tps)
    return data


bandwidths = [500000, 1000000,1500000,  2000000]
all_data = []

for b in bandwidths:
    all_data.extend(get_diff_data(b))


labels = ['ETH-full', 'ETH-fast', 'tMPT', 'Proposed']
colors = ['skyblue', 'lightgreen', 'lightgray', 'lightpink']
hatches = [' ', ' ', ' ', ' ']

plt.figure(figsize=(7, 6))
positions = []
for i in range(len(bandwidths)):
    positions.extend([x + i * (len(labels) + 1.3) for x in range(1, len(labels) + 1)])

box = plt.boxplot(all_data, positions=positions, patch_artist=True, widths=1)
for patch, color, hatch in zip(box['boxes'], colors * len(bandwidths), hatches * len(bandwidths)):
    patch.set_facecolor(color)
    patch.set_hatch(hatch)

plt.ylabel('TPS (Txs/sec)', fontsize=16)
plt.xlim(-0.5, len(bandwidths) * (len(labels) + 1)+1.5 )
plt.xticks([i * (len(labels) + 1) + 2.5 for i in range(len(bandwidths))], [f'{b/125000} Mbps' for b in bandwidths], fontsize=16)

# 添加图例
legend_patches = [plt.Line2D([0], [0], color=color, lw=8, label=label) for color, label in zip(colors, labels)]
plt.legend(handles=legend_patches, fontsize=16, loc='upper left')
plt.yticks(fontsize=16)
plt.tight_layout()
plt.savefig(f'./pics/tps_vs_epoch_box_vary_bandwidth.pdf')
plt.show()

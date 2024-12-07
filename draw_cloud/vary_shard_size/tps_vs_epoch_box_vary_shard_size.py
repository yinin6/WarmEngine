import pandas as pd
import matplotlib.pyplot as plt

shard_num = 16

bandwidth = 2000000


def get_diff_data(node_num):
    data = []
    move_node=int(node_num/2)+1
    zone_size=int(move_node/2)
    for mod in [3, 1, 2]:
        tps = None
        for i in range(shard_num):
            file_path = f"../../ali_result/result/data=3000000_inj=10000_block=1000_S16N{node_num}/Mod{mod}_Num{move_node}_Zone{zone_size}_frq50_band{bandwidth}/pbft_shardNum={shard_num}/epochDatil.csvS{i}0.csv"

            df = pd.read_csv(file_path)
            if tps is None:
                tps = df['tps'].tolist()
            else:
                tps = [x + y for x, y in zip(tps, df['tps'].tolist())]
        tps = tps[:5]
        data.append(tps)
    return data


node_nums=[10,15]

all_data = []

for b in node_nums:
    all_data.extend(get_diff_data(b))

labels = [ 'ETH-fast', 'tMPT', 'Proposed']
colors = [ '#B7B7EB', '#9D9EA3', '#F09BA0']
colors = [
          (183 / 255, 183 / 255, 235 / 255, 0.5), (157 / 255, 158 / 255, 163 / 255, 0.5),
          (240 / 255, 155 / 255, 160 / 255, 0.5)]
hatches = [' ', ' ', ' ', ' ']

plt.figure(figsize=(7, 6))
positions = []
for i in range(len(node_nums)):
    positions.extend([x + i * (len(labels) + 1.4) for x in range(1, len(labels) + 1)])

box = plt.boxplot(all_data, positions=positions, patch_artist=True, widths=0.9)
for patch, color, hatch in zip(box['boxes'], colors * len(node_nums), hatches * len(node_nums)):
    patch.set_facecolor(color)
    patch.set_hatch(hatch)

plt.ylabel('TPS (TXs / Sec.)', fontsize=20)
plt.xlim(-0.5, len(node_nums) * (len(labels) + 1) + 1.5)
plt.xticks([i * (len(labels) + 1)+2  for i in range(len(node_nums))],
           [f'  {int(b)} ' for b in node_nums], fontsize=20)

import matplotlib.ticker as mticker

# 设置y轴为科学计数法
plt.gca().yaxis.set_major_formatter(mticker.ScalarFormatter(useMathText=True))
plt.gca().yaxis.get_major_formatter().set_scientific(True)
plt.gca().yaxis.get_major_formatter().set_powerlimits((0, 0))

# 添加图例
legend_patches = [plt.Line2D([0], [0], color=color, lw=8, label=label) for color, label in zip(colors, labels)]
plt.legend(handles=legend_patches, fontsize=20, loc='upper right')
plt.yticks(fontsize=20)
plt.xlabel('Num. of nodes in per shard', fontsize=20)
plt.tight_layout()
plt.savefig(f'../pics/tps_vs_epoch_box_vary_bandwidth.pdf')
plt.show()

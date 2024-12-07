import pandas as pd
import pandas as pd
import matplotlib.pyplot as plt


mods=[3,1,2]

font_size=25

mod_to_label = {
    0: 'ETH-full',
    3: 'ETH-fast',
    1: 'tMPT',
    2: 'Proposed'
}

mod_to_color = {
    0: '#7195C5',
    3: '#7262ac',
    1: '#01844F',
    2: '#E9212C'
}

import numpy as np
fig, axs = plt.subplots(len(mods), 1, figsize=(7, 2 * len(mods)))  # 创建子图

for idx, mod in enumerate(mods):
    file_path=f'../ali_result/result/data=3000000_inj=10000_block=1000_S16N10/Mod{mod}_Num6_Zone3_frq50_band500000/cpuS3/s3n7_cpu_mem_usage_single.csv'

    df = pd.read_csv(file_path)

    # 初始化变量
    start_index = 0
    diffs = [np.nan] * len(df)  # 新增一列来存储每个递增段的差值

    for i in range(1, len(df)):
        if df['CPU_Usage_Percent'][i] <= df['CPU_Usage_Percent'][i - 1]:  # 检测递增段的结束
            # 计算当前递增段的差值
            segment_diff = df['CPU_Usage_Percent'][start_index:i].max() - df['CPU_Usage_Percent'][start_index:i].min()

            # 将差值填入递增段的每一行
            for j in range(start_index, i):
                diffs[j] = segment_diff

            # 更新递增段的开始索引
            start_index = i

    # 最后一段递增段处理
    segment_diff = df['CPU_Usage_Percent'][start_index:].max() - df['CPU_Usage_Percent'][start_index:].min()
    for j in range(start_index, len(df)):
        diffs[j] = segment_diff

    # 将结果添加为新的列
    df['Increase_Segment_Diff'] = diffs

    print(df)

    axs[idx].set_ylim(0, 4.5)
    axs[idx].set_xlim(-100, 2900)
    axs[idx].tick_params(axis='x', labelsize=15)  # 设置 x 轴刻度字体大小
    axs[idx].tick_params(axis='y', labelsize=15)  # 设置 y 轴刻度字体大小



    if idx==0:
        # axs[idx].set_title(f'Data provider node CPU usage over Time')
        print()
    if idx==1:
        axs[idx].set_ylabel('CPU load (%)',fontsize=font_size)
    else:
        axs[idx].set_ylabel(' ')


    cpu_usage = df['CPU_Usage_Percent']


    print(mod_to_label)
    axs[idx].plot(cpu_usage, label=mod_to_label[mod], color=mod_to_color[mod], linestyle='-')
    if idx==2:
        axs[idx].set_xlabel('Running time (Sec.)',fontsize=font_size)

    axs[idx].legend(fontsize=16, loc='upper right')
    # axs[idx].grid(True)

    # Identify points with sudden increase

    max_value = df['Increase_Segment_Diff'][500:1000].max()
    print(max_value)


    if idx==0:
        j = 730

        axs[idx].annotate(
            f'  ',
            xy=(j, cpu_usage[j]-0.2),
            xytext=(j, cpu_usage[j] - 1.8),
            arrowprops=dict(facecolor='red', shrink=1),
            fontsize=16,
            color='red'
        )
        axs[idx].text(j + 900, cpu_usage[j] - 1.8,
                      f' {max_value:.2f}% increase, during reconfiguration',
                      fontsize=17, color='red', ha='center', va='center')
        axs[idx].text(550, cpu_usage[j] + 1.8,
                      f'Severe fluctuations',
                      fontsize=17, color='red', ha='center', va='center')

    if idx==1:
        j = 730
        axs[idx].annotate(
            f' ',
            xy=(j, cpu_usage[j]-0.35),
            xytext=(j, cpu_usage[j] - 2),
            arrowprops=dict(facecolor='red', shrink=1),
            fontsize=16,
            color='red'
        )
        axs[idx].text(j + 600, cpu_usage[j] - 1.8,
                      f' {max_value:.2f}% increase, during reconfig.',
                      fontsize=17, color='red', ha='center', va='center')
        axs[idx].text(530, 4 ,
                      f'Severe fluctuations',
                      fontsize=17, color='red', ha='center', va='center')
    if idx==2:
        j = 700
        axs[idx].annotate(
            ' ',
            xy=(j, cpu_usage[j]-0.2),
            xytext=(j, cpu_usage[j] - 1.7),
            arrowprops=dict(facecolor='red', shrink=1),
            fontsize=18,
            color='red'
        )
        axs[idx].text(j+600, cpu_usage[j] - 1.5,
                      f' {max_value:.2f}% increase, during reconfig.',
                fontsize=17, color='green', ha='center', va='center')

        # axs[idx].text(600, 4,
        #               f'Minimal and stable CPU load',
        #               fontsize=16, color='green', ha='center', va='center')


plt.tight_layout()
plt.savefig(f'./pics/cpu_vs_time.pdf')
plt.show()

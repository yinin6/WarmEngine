import matplotlib.pyplot as plt
import seaborn as sns
import numpy as np
import pandas as pd

# 生成示例数据
np.random.seed(42)
categories = ['A', 'B', 'C', 'D', 'E']
data = {category: np.random.normal(loc=i * 10, scale=5, size=100) for i, category in enumerate(categories)}
df = pd.DataFrame(data)

# 计算每组的均值
means = df.mean()

# 设置图形
plt.figure(figsize=(10, 6))

# 绘制箱线图
sns.boxplot(data=df, width=0.6, palette='Set3', showfliers=False)

# 绘制折线图（显示均值）
plt.plot(range(len(categories)), means, marker='o', color='red', linestyle='-', label='Mean')

# 美化图形
plt.title('Boxplot with Line Plot', fontsize=16)
plt.xlabel('Categories', fontsize=12)
plt.ylabel('Values', fontsize=12)
plt.xticks(ticks=range(len(categories)), labels=categories)
plt.legend()
plt.grid(axis='y', linestyle='--', alpha=0.7)

# 显示图形
plt.show()

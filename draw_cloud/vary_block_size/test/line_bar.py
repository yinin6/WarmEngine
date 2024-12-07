import seaborn as sns
import matplotlib.pyplot as plt
import numpy as np
import pandas as pd

# 创建示例数据
np.random.seed(10)
x = np.arange(0, 10)
y = np.sin(x) + np.random.normal(0, 0.2, size=10)  # 生成带噪声的y数据

# 将数据转换为DataFrame格式
data = pd.DataFrame({'x': x, 'y': y})

# 使用Seaborn绘制带误差条的折线图
sns.lineplot(x='x', y='y', data=data, ci='sd', err_style='bars', color='blue')

# 设置标题和显示图形
plt.title('Line plot with Error Bars')
plt.show()

import numpy as np

path='../20250000to20499999_BlockTransaction.csv'





import pandas as pd

# 读取 CSV 文件
df = pd.read_csv(path)  # 请将 your_file.csv 替换为实际文件名

print(df.shape)

filtered_df = df.head(5000000)

# 将 'from', 'to', 'value' 以外的列值设置为 NaN
columns_to_clear = [col for col in filtered_df.columns if col not in ['from', 'to', 'value']]
filtered_df[columns_to_clear] = np.nan

# 截取前500万行


# 保存到新的 CSV 文件
filtered_df.to_csv('filtered_output_5M.csv', index=False)

print("处理完成，前500万行数据已保存到 'filtered_output_5M.csv'")


# 将筛选后的结果保存到新的 CSV 文件
# filtered_df.to_csv('filtered_output.csv', index=False)

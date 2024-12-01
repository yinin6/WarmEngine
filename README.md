

## Project

- **main** branch: This branch contains the core code and main functionality of the project. All major development work is focused on this branch.
- **experiment** branch: This branch is dedicated to storing experimental code, data, and results. It includes content related to experimental setup and data processing.

## Data Source for the Experiment

* The experiment uses on-chain data from the blockchain, with the data sourced from XBlock. 
* The data can be downloaded from the following link: [https://zhengpeilin.com/download.php?file=20250000to20499999_BlockTransaction.zip](https://zhengpeilin.com/download.php?file=20250000to20499999_BlockTransaction.zip).
* Due to the large size of the dataset, only the first 100K records are extracted as a sample and saved in the file `small_dataset_100K.csv`.


## System Execution Script

* `bat_compile_run_shardNum=8_NodeNum=10.bat` is designed to run on a Windows system. It launches 8 shards, with 10 nodes in each shard.
* `compile_for_linux.bat` is designed to run on a Linux system. It compiles the code for the experiment.
* `ipTable.json` contains the IP addresses of the nodes in the network.
* `paramsConfig.json` contains the configuration parameters for the experiment.



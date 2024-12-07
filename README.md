

## Experiment

## file
- `ip_table_generator.py`: generates the IP address table.
- `control_cloud.py`: the control program of the cloud server.

## directory
- `bin`: contains the binary files and configuration.
- `ip`: contains the IP address of the server under different size.
- `draw_cloud`: contains the script to draw figures.
- `draw_of_local`: is a script for plotting experimental graphs based on the results obtained from local experiments.
- `ali_result`: contains experimental data. Since there are many files, they are packed into a compressed archive. URL: [ali_result.zip](https://drive.google.com/file/d/1PuqvrviE9m8Kz-y5-H5CNMKUNfuPzhFl/view?usp=sharing)



## Usage


1. **Purchase Cloud Servers**  
   Export the public and private IP addresses of each cloud server. The public IPs are used for running the control nodes, while the private IPs are used for communication between nodes.

2. **Run `ip_table_generator.py`**  
   Set the system scale and run the script to generate the node IP table. This table is used for node-to-node communication and for distributing control nodes across the cloud servers.

3. **Upload the Dataset**  
   Upload the dataset to the cloud servers.

4. **Contents of the `bin` Directory**  
   The `bin` directory contains:
    - The compiled binary executable file, `myapp_linux`.
    - The system parameter configuration file, `paramsConfig.json`.

5. **Run `ip_table_generator.py`**  
   Use the control script to manage the startup of nodes on the cloud servers, execute experiments, and download the results.




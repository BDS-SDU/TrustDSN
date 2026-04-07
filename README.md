# TrustDSN

## 目录
- [1. 项目概述](#1-项目概述)
- [2. 部署指南](#2-部署指南)
- [3. 运行示例](#3-运行示例)
- [4. 系统前端部署](#4-系统前端部署)
- [5. 退出系统](#5-退出系统)

## 1. 项目概述

TrustDSN 是论文《BFT-DSN: A Byzantine Fault-Tolerant Decentralized Storage Network》(IEEE TC 2024) 的代码实现。该项目实现了一个具备拜占庭容错机制的去中心化存储网络，通过结合存储加权的 BFT 共识、纠删码、同态指纹和权阈值签名技术，在确保高安全性和强一致性的同时，实现高效、低成本的数据存储。

Copyright (c) 2024-2025, Guo Hechuan, MIT License

## 2. 部署指南

### 2.1 系统要求

1. 硬件要求
- CPU：2 核及以上
- 内存：4GB 及以上
- 存储：支持 8MiB sectors 的存储空间
- 网络：稳定的网络连接

2. 软件要求
- 操作系统：Linux 或 macOS
- Go：1.18.1 或更高版本
- Rust：建议通过 `rustup` 安装
- 其他依赖：`git`、`jq`、`pkg-config`、`clang`、`hwloc` 等

### 2.2 环境准备与编译

1. 安装系统依赖

Ubuntu / Debian:

```bash
sudo apt install mesa-opencl-icd ocl-icd-opencl-dev gcc git bzr jq pkg-config curl clang build-essential hwloc libhwloc-dev wget -y && sudo apt upgrade -y
```

macOS:

```bash
brew install go bzr jq pkg-config rustup hwloc coreutils
```

2. 安装 Rust

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

3. 安装 Go

```bash
wget -c https://golang.org/dl/go1.18.1.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
```

4. 配置环境变量

```bash
export LOTUS_PATH=~/.lotus-local-net
export LOTUS_MINER_PATH=~/.lotus-miner-local-net
export LOTUS_SKIP_GENESIS_CHECK=_yes_
export CGO_CFLAGS_ALLOW="-D__BLST_PORTABLE__"
export CGO_CFLAGS="-D__BLST_PORTABLE__"
export IPFS_GATEWAY=https://proof-parameters.s3.cn-south-1.jdcloud-oss.com/ipfs/
```

5. 获取源码并构建

```bash
git clone https://github.com/BDS-SDU/TrustDSN.git
cd TrustDSN
make debug
```

### 2.3 启动创世节点

TrustDSN 当前推荐通过脚本一键启动第一个节点：

```bash
cd /path/to/TrustDSN
bash scripts/genesis_node_start.sh
```

这个脚本会自动完成以下工作：

- 清理旧的本地网络数据和日志
- 拉取 8MiB 参数文件
- 预密封 2 个创世扇区
- 生成 `localnet.json` 和 `devgen.car`
- 启动创世 `lotus daemon`
- 导入创世矿工密钥
- 初始化创世矿工 `t01000`
- 启动 `lotus-miner`
- 启动 `listen_and_send.sh`
- 启动 `trustdsn-api`

脚本末尾会输出两条很重要的信息：

```bash
./lotus net listen
./lotus-miner net listen
```

这两条命令分别给出：

- 创世节点 daemon 的 multiaddr
- 创世节点 miner 的 multiaddr

后续添加更多节点时，需要把这两个地址记录下来。

### 2.4 添加更多节点

新增节点时，推荐统一使用：

```bash
bash scripts/node_start.sh <daemon_multiaddr> <miner_multiaddr> <genesis_ip> [local_ip]
```

在运行脚本前，还需要先把创世节点生成的：

```text
devgen.car
```

复制到新节点的仓库根目录。

其中：

- `daemon_multiaddr`
  - 来自创世节点执行 `./lotus net listen` 的输出
- `miner_multiaddr`
  - 来自创世节点执行 `./lotus-miner net listen` 的输出
- `genesis_ip`
  - 创世节点所在机器的 IP
  - 新节点会通过这个地址把钱包地址和本机 IP 发给创世节点上的 `listen_and_send.sh`
- `local_ip`
  - 可选参数
  - 如果不传，脚本会自动检测本机 IP
  - 如果本机同时有多个网卡或多个地址，建议显式指定

`scripts/node_start.sh` 会自动完成“普通节点接入网络并成为存储提供者”的流程。脚本内部包括：

- 启动本地 `lotus daemon`
- 连接到创世节点 daemon 和 miner
- 创建钱包地址
- 将钱包地址和本机 IP 发给创世节点
- 等待创世节点打款
- 初始化 `lotus-miner`
- 启动 `lotus-miner`
- 配置 `sealed` 与 `unseal` 存储目录

一个典型示例如下：

```bash
bash scripts/node_start.sh \
  /ip4/192.168.1.10/tcp/1234/p2p/12D3KooW... \
  /ip4/192.168.1.10/tcp/2345/p2p/12D3KooW... \
  192.168.1.10 \
  192.168.1.11
```

如果最后一个 `local_ip` 不传，脚本会优先自动检测公网 IPv4；没有公网时，再回退到私网 IPv4。

### 2.5 常见问题

1. 创世节点启动失败
- 检查 `go`、`rustup` 和系统依赖是否安装完整
- 检查 `make debug` 是否成功完成
- 检查端口是否被旧进程占用

2. 新节点无法加入网络
- 确认新节点目录下已经复制了 `devgen.car`
- 确认 `daemon_multiaddr` 和 `miner_multiaddr` 复制完整
- 确认创世节点 `9999` 端口和节点通信端口在内网可达

3. 新节点没有收到创世节点打款
- 确认创世节点上的 `listen_and_send.sh` 已启动
- 确认 `genesis_ip` 填写正确
- 确认新节点向创世节点发送的钱包地址和本机 IP 没有被防火墙拦截

4. 多网卡环境 IP 不正确
- 创世节点和普通节点现在都优先选择公网 IPv4
- 如果你的部署环境更适合使用私网地址，建议在 `node_start.sh` 中显式传入 `local_ip`

## 3. 运行示例

### 3.1 创建存储交易：`bftdsn deal`

TrustDSN 主要通过 `lotus bftdsn deal` 发起文件存储。最常用命令如下：

```bash
./lotus bftdsn deal <inputPath>
```

该命令会完成：

- 对输入文件进行纠删码切片
- 将切片导入本地 client
- 自动向网络中的 miner 发起存储交易
- 记录原始文件名到 `filenames.log`
- 生成 `<fileName>_meta`，用于后续检索

演示环境下更推荐显式指定参数，例如：

```bash
./lotus bftdsn deal -k 3 -m 1 --keep-chunks file1
```

参数说明：

- `-k`
  - 数据分片数
- `-m`
  - 校验分片数
- `--keep-chunks`
  - 保留生成的切片文件，便于调试

命令执行完成后，终端会提示：

- 所有交易是否已发送
- 总耗时
- 可用于后续检索的文件名

### 3.2 文件检索：`bftdsn retrieve`

使用文件名从 TrustDSN 网络检索文件：

```bash
./lotus bftdsn retrieve <fileName> <outPath>
```

例如：

```bash
./lotus bftdsn retrieve -k 3 -m 1 file1 retrieve_f1
```

这个命令会：

- 读取 `<fileName>_meta`
- 逐个向 miner 发起 retrieval deal
- 下载所有分片
- 自动解码并恢复原始文件到 `<outPath>`

如果不希望保留下载后的分片，可使用默认行为；如果需要调试，也可以加：

```bash
--keep-chunks
```

### 3.3 查看可检索文件：`bftdsn list-files`

要查看当前工作区中已经记录、可直接检索的文件名：

```bash
./lotus bftdsn list-files
```

该命令会读取：

```text
filenames.log
```

并列出去重后的文件名列表。  
这些名字就是 `bftdsn retrieve` 的第一个参数。

### 3.4 交易管理

1. 查看交易状态

```bash
./lotus client list-deals
```

2. 查看交易详情

```bash
./lotus client get-deal <DealCID>
```

### 3.5 常见问题

1. `bftdsn deal` 失败
- 检查 miner 是否已全部上线
- 检查创世节点和普通节点是否都已完成初始化
- 检查文件路径是否正确

2. `bftdsn retrieve` 失败
- 检查输入的是否是 `list-files` 返回的文件名
- 检查 `<fileName>_meta` 是否存在
- 检查 `k` 和 `m` 参数是否与交易创建时一致

3. `list-files` 没有内容
- 说明当前工作区里还没有成功执行过 `bftdsn deal`
- 或者 `filenames.log` 已被清理

## 4. 系统前端部署

### 4.1 依赖安装

前端页面依赖：

- Node.js
- npm
- nginx

先检查当前机器是否已经安装：

```bash
node -v
npm -v
nginx -v
```

如果 `node -v` 和 `npm -v` 发现没有安装，或者版本过低导致前端无法编译，可以使用 `nvm` 安装较新的 Node.js 版本，例如：

```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
source ~/.bashrc

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

nvm install 22
nvm use 22
nvm alias default 22

node -v
npm -v
which node
which npm
```

如果 `nginx` 没有安装，可以在 Ubuntu 上执行：

```bash
sudo apt update
sudo apt install -y nginx
```

### 4.2 前端页面编译

在 `demo-web` 目录下执行：

```bash
cd /path/to/TrustDSN/demo-web
npm install
npm run build
```

构建完成后会生成：

```text
demo-web/dist
```

### 4.3 nginx 托管

仓库里已经提供了 nginx 配置样板：

```text
deploy/nginx/sites-available/trustdsn
```

需要注意，这个样板文件中的 `root` 当前是按仓库默认路径写死的。例如：

```nginx
root /home/jiahao/go/src/TrustDSN/demo-web/dist;
```

如果你的 TrustDSN 实际部署路径不同，拷贝到系统 nginx 目录之前，需要先把它改成你自己的仓库绝对路径。

当前样板默认监听8081端口，这是为了方便在本地测试时与其他系统并行打开。如果你是单独正式部署 TrustDSN，可以把样板中的：

```nginx
listen 8081;
```

改成：

```nginx
listen 80;
```

Ubuntu 上的典型托管步骤如下：

1. 拷贝配置

```bash
sudo cp /path/to/TrustDSN/deploy/nginx/sites-available/trustdsn /etc/nginx/sites-available/trustdsn
```

2. 启用站点

```bash
sudo ln -sf /etc/nginx/sites-available/trustdsn /etc/nginx/sites-enabled/trustdsn
```

3. 如有需要，移除默认站点

```bash
sudo rm -f /etc/nginx/sites-enabled/default
```

4. 检查配置

```bash
sudo nginx -t
```

5. 重载 nginx

```bash
sudo systemctl reload nginx
```

### 4.4 查看前端页面

1. 先启动后端和链服务：

```bash
cd /path/to/TrustDSN
bash scripts/genesis_node_start.sh
```

2. 再通过浏览器访问页面：

- 如果 nginx 样板保持 `8081`：

```text
http://<server-ip>:8081
```

- 如果你已经改成 `80`：

```text
http://<server-ip>
```

3. 可以用下面的命令检查后端和 nginx 是否正常：

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8081/api/miners
```

如果页面能打开但数据是空的，优先检查：

- 是否重新执行过 `npm run build`
- 是否已经重载 nginx
- 浏览器是否缓存了旧资源
- 当前运行的是否真的是 `trustdsn-api`

## 5. 退出系统

TrustDSN 仓库提供了统一的退出脚本：

```bash
bash scripts/exit.sh
```

该脚本会尝试停止：

- `listen_and_send.sh`
- 监听 `9999` 端口的 `nc` 进程
- `trustdsn-api`
- `lotus-miner`
- `lotus daemon`

如果你仍然在开发模式下运行前端 `npm run dev`，该脚本也会尝试停止前端 dev server。

需要注意：

- 如果前端是通过 `nginx + dist` 托管的，`exit.sh` 不会关闭 nginx
- 如需停止 nginx，需要单独执行：

```bash
sudo systemctl stop nginx
```

脚本执行完成后，终端会输出：

```text
Exit script finished.
```

表示后端仓库内管理的相关进程已经基本退出。

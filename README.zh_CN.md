<div align="center">
  <h1>TrustDSN: A Byzantine Fault-Tolerant Decentralized Storage Network</h1>
  <p>
    <img src="https://img.shields.io/badge/Go-%E2%89%A51.18.1-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
    <img src="https://img.shields.io/badge/Rust-rustup-black?style=flat&logo=rust&logoColor=white" alt="Rust">
    <img src="https://img.shields.io/badge/Sector-8MiB-blue" alt="Sector">
    <img src="https://img.shields.io/badge/Frontend-React%20%2B%20Vite-61DAFB?style=flat&logo=react&logoColor=white" alt="Frontend">
    <img src="https://img.shields.io/badge/License-MIT-green" alt="License">
  </p>
  <p>
    <a href="./README.md">English</a> · <a href="./README.zh_CN.md">简体中文</a>
  </p>
</div>

TrustDSN 是一个面向多节点部署的去中心化存储网络系统，提供文件编码、分片存储、跨节点检索、矿工状态采集、Proof 信息展示以及前后端一体化的演示与运维能力。

它既可以作为本地多节点实验环境使用，也可以部署到云服务器中，通过脚本快速拉起创世节点、普通存储节点、后端 API 与前端页面。

## 🎬 系统演示

<p align="center">
  <a href="./assets/readme/TrustDSN_video.mp4"><strong>点击查看演示视频</strong></a>
</p>

<p align="center">
  <video src="./assets/readme/TrustDSN_video.mp4" controls width="880"></video>
</p>

> 如果当前代码托管平台不支持直接预览视频，可点击上方链接下载或单独打开 `assets/readme/TrustDSN_video.mp4`。

## ✨ 系统特性

- **多节点部署**：支持创世节点与多个普通节点快速组网。
- **可靠存储**：基于纠删码切片进行文件存储与恢复。
- **可视化运维**：前端页面可展示存储节点信息、最新 Proof 信息和文件操作结果。
- **脚本化管理**：提供启动脚本、退出脚本和 nginx 托管样板，便于本地实验与云部署。
- **演示友好**：支持前后端分离部署，适合实验演示、课程展示和系统验证。

## 目录
- [1. 项目概述](#1-项目概述)
- [2. 部署指南](#2-部署指南)
- [3. 运行示例](#3-运行示例)
- [4. 系统前端部署](#4-系统前端部署)
- [5. 退出系统](#5-退出系统)

## 1. 项目概述

TrustDSN 由以下几个核心部分组成：

- **链与存储节点**
  - `lotus daemon` 与 `lotus-miner` 负责链服务与存储节点运行。
- **节点启动脚本**
  - `scripts/genesis_node_start.sh` 用于启动创世节点。
  - `scripts/node_start.sh` 用于让普通节点接入现有网络。
- **后端 API**
  - `cmd/trustdsn-api` 提供前端调用接口，负责聚合 miner 信息、proof 信息以及文件操作。
- **前端页面**
  - `demo-web` 提供系统展示页面，可通过 `nginx` 托管。
- **退出脚本**
  - `scripts/exit.sh` 用于统一关闭仓库内启动的核心进程。

TrustDSN 当前默认围绕 `8MiB` sector、脚本化启动流程和演示型前端进行组织，便于在实验室内网、云服务器或课程展示环境中快速复现一套完整的 DSN 系统。

Copyright (c) 2024-2025, Guo Hechuan, MIT License

## 2. 部署指南

### 2.1 系统要求

<table>
  <tr>
    <th align="left">类别</th>
    <th align="left">要求</th>
  </tr>
  <tr>
    <td>CPU</td>
    <td>2 核及以上</td>
  </tr>
  <tr>
    <td>内存</td>
    <td>4GB 及以上</td>
  </tr>
  <tr>
    <td>存储</td>
    <td>支持 8MiB sectors 的存储空间</td>
  </tr>
  <tr>
    <td>网络</td>
    <td>稳定的网络连接</td>
  </tr>
  <tr>
    <td>操作系统</td>
    <td>Linux 或 macOS</td>
  </tr>
  <tr>
    <td>Go</td>
    <td>1.18.1 或更高版本</td>
  </tr>
  <tr>
    <td>Rust</td>
    <td>建议通过 <code>rustup</code> 安装</td>
  </tr>
  <tr>
    <td>其他依赖</td>
    <td><code>git</code>、<code>jq</code>、<code>pkg-config</code>、<code>clang</code>、<code>hwloc</code> 等</td>
  </tr>
</table>

> [!TIP]
> 如果你计划在云服务器上部署多节点网络，建议优先准备同一内网中的多台机器，并提前确认节点之间的私网互通和安全组规则。

### 2.2 环境准备与编译

#### 2.2.1 安装系统依赖

Ubuntu / Debian:

```bash
sudo apt install mesa-opencl-icd ocl-icd-opencl-dev gcc git bzr jq pkg-config curl clang build-essential hwloc libhwloc-dev wget -y && sudo apt upgrade -y
```

macOS:

```bash
brew install go bzr jq pkg-config rustup hwloc coreutils
```

#### 2.2.2 安装 Rust

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

#### 2.2.3 安装 Go

```bash
wget -c https://golang.org/dl/go1.18.1.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
```

#### 2.2.4 配置环境变量

```bash
export LOTUS_PATH=~/.lotus-local-net
export LOTUS_MINER_PATH=~/.lotus-miner-local-net
export LOTUS_SKIP_GENESIS_CHECK=_yes_
export CGO_CFLAGS_ALLOW="-D__BLST_PORTABLE__"
export CGO_CFLAGS="-D__BLST_PORTABLE__"
export IPFS_GATEWAY=https://proof-parameters.s3.cn-south-1.jdcloud-oss.com/ipfs/
```

#### 2.2.5 获取源码并构建

```bash
git clone https://github.com/BDS-SDU/TrustDSN.git
cd TrustDSN
make debug
```

> [!IMPORTANT]
> 构建完成后，请确认仓库根目录下已经生成 `lotus`、`lotus-miner`、`lotus-seed` 等可执行文件，再继续后续步骤。

### 2.3 启动创世节点

TrustDSN 当前推荐通过脚本一键启动第一个节点：

```bash
cd /path/to/TrustDSN
bash scripts/genesis_node_start.sh
```

#### 脚本会自动完成的工作

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

#### 脚本执行后需要记录的信息

```bash
./lotus net listen
./lotus-miner net listen
```

这两条命令分别给出：

- 创世节点 daemon 的 multiaddr
- 创世节点 miner 的 multiaddr

后续添加更多节点时，需要把这两个地址记录下来。

> [!TIP]
> `genesis_node_start.sh` 已经集成了创世节点、监听脚本和后端 API 的启动逻辑。部署时推荐优先使用该脚本，而不是手动逐条执行命令。

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

#### 参数说明

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

#### `node_start.sh` 自动完成的工作

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

> [!IMPORTANT]
> 新节点在运行前必须已经拿到创世节点生成的 `devgen.car`，并且能够访问创世节点的 `9999` 端口以及 lotus daemon 和 lotus miner 的通信地址。

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

> [!TIP]
> 演示环境下推荐固定使用 `-k 3 -m 1 --keep-chunks`，这样更方便观察分片结果和后续检索过程。

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

> [!TIP]
> 如果你在云服务器上遇到前端构建失败，优先检查 `node -v`。较旧的 Node 版本很容易导致 `npm run build` 失败。

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

> [!IMPORTANT]
> 正式部署时，请使用 `npm run build` 生成静态页面并交给 `nginx` 托管，不要依赖 `npm run dev` 作为长期运行方案。

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

#### Ubuntu 上的典型托管步骤

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

> [!TIP]
> 如果页面静态内容正常、但数据为空，通常优先检查：
> 1. 后端 API 是否已经启动
> 2. nginx 是否已重载
> 3. 前端是否重新执行过 `npm run build`
> 4. 浏览器是否缓存了旧资源

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

#### 注意事项

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

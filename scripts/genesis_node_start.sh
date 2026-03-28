#!/bin/bash

rm -rf devgen.car
rm -rf localnet.json
rm -rf ~/.genesis-sectors
rm -rf ~/.lotus
rm -rf ~/.lotusminer
rm -rf DB
rm -rf SST
rm -rf client_upload
rm -rf client_download
rm -rf acc
rm -rf leaf
rm -rf paths
rm -rf zk_output
rm lotus.log
rm miner.log
rm fund.log
rm api.log
rm demo-web.log
rm aggregate.log
rm filenames.log


./lotus fetch-params 8MiB
./lotus-seed pre-seal --sector-size 8MiB --num-sectors 2
./lotus-seed genesis new localnet.json
./lotus-seed genesis add-miner localnet.json ~/.genesis-sectors/pre-seal-t01000.json
#tmux new-session -s "lotus-daemon" -d "./lotus daemon --lotus-make-genesis=devgen.car --genesis-template=localnet.json --bootstrap=false"
nohup ./lotus daemon --lotus-make-genesis=devgen.car --genesis-template=localnet.json --bootstrap=false > lotus.log 2>&1 &

ps -ef | grep lotus

echo "sleep 30s" && sleep 30s
./lotus wallet import --as-default ~/.genesis-sectors/pre-seal-t01000.key
./lotus-miner init --genesis-miner --actor=t01000 --sector-size=8MiB --pre-sealed-sectors=~/.genesis-sectors --pre-sealed-metadata=~/.genesis-sectors/pre-seal-t01000.json --nosync
#tmux new-session -s "lotus-miner" -d "./lotus-miner run --nosync"
nohup ./lotus-miner run --nosync > miner.log 2>&1 &

sleep 30

./lotus net listen
sleep 3
./lotus-miner net listen
sleep 3
./lotus-miner info

nohup bash scripts/listen_and_send.sh > fund.log 2>&1 &

sleep 3

nohup go run ./cmd/trustdsn-api > api.log 2>&1 &
nohup bash -lc 'cd demo-web && npm run dev' > demo-web.log 2>&1 &

#nohup bash scripts_FileDES/server_listen_aggregate.sh > aggregate.log 2>&1 &


#./main bls.txt

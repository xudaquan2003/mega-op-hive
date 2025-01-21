#!/bin/sh
set -exu

op-deployer init --intent-config-type custom --l1-chain-id 3151908 --l2-chain-ids 2151908 --workdir /network-data

export PRIVATE_KEY=eaba42282ad33c8ef2524f07277c03a776d98ae19f581990ce75becb7cfa1c23
export FUND_VALUE=10ether
export L1_NETWORK=local
export L1_RPC_URL=$HIVE_L1_RPC_URL
export ETH_RPC_URL=$L1_RPC_URL
/fund-script/fund.sh 3151908 

cp /intent.toml /network-data/intent.toml

op-deployer apply --l1-rpc-url $L1_RPC_URL --private-key $PRIVATE_KEY --workdir /network-data
op-deployer inspect genesis --workdir /network-data --outfile /network-data/genesis-2151908.json 2151908
op-deployer inspect rollup --workdir /network-data --outfile /network-data/rollup-2151908.json 2151908


jq --from-file /fund-script/gen2spec.jq < "/network-data/genesis-2151908.json" > "/network-data/chainspec-2151908.json"


# while true; do
#   echo "Container is running..."
#   sleep 60
# done

reth node \
    --chain /network-data/genesis-2151908.json \
    --auto-mine \
    --dev.block-time 1s \
    --http \
    --http.addr=0.0.0.0 \
    --http.api=admin,debug,eth,net,web3 \
    --ws \
    --ws.addr=0.0.0.0 \
    --ws.api=admin,debug,eth,net,web3 \
    --nat none



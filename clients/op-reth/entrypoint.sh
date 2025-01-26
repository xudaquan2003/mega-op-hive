#!/bin/sh
set -exu

if [ -f /genesis_account.json ]; then
    jq -s '.[0] * .[1]' /genesis_account.json /genesis.json  > /genesis_tmp.json
    mv /genesis_tmp.json /genesis.json
fi

op-reth node \
    --datadir=/data/op-reth/execution-data \
    --chain=/genesis.json \
    --http \
    --http.port=$HIVE_L2_HTTP_PORT \
    --http.addr=0.0.0.0 \
    --http.corsdomain=* \
    --http.api=admin,net,eth,web3,debug,trace,miner \
    --ws \
    --ws.addr=0.0.0.0 \
    --ws.port=$HIVE_L2_WS_PORT \
    --ws.api=admin,net,eth,web3,debug,trace,miner \
    --ws.origins=* \
    --authrpc.port=$HIVE_L2_AUTH_PORT \
    --authrpc.jwtsecret=/jwtsecret \
    --authrpc.addr=0.0.0.0 \
    --discovery.port=30303 \
    --port=30303 \
    --rpc.eth-proof-window=302400 \
    --metrics=0.0.0.0:9001


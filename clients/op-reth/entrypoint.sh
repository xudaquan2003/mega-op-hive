#!/bin/sh
set -exu

op-reth node \
    --datadir=/data/op-reth/execution-data \
    --chain=/genesis-2151908.json \
    --http \
    --http.port=9545 \
    --http.addr=0.0.0.0 \
    --http.corsdomain=* \
    --http.api=admin,net,eth,web3,debug,trace,miner \
    --ws \
    --ws.addr=0.0.0.0 \
    --ws.port=9546 \
    --ws.api=admin,net,eth,web3,debug,trace,miner \
    --ws.origins=* \
    --authrpc.port=9551 \
    --authrpc.jwtsecret=/jwtsecret \
    --authrpc.addr=0.0.0.0 \
    --discovery.port=30303 \
    --port=30303 \
    --rpc.eth-proof-window=302400 \
    --metrics=0.0.0.0:9001


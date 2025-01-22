#!/bin/sh
set -exu

devnet \
    --datadir=/data/op-reth/execution-data \
    --chain=/genesis-2151908.json \
    --http \
    --http.port=$HIVE_L2_HTTP_PORT \
    --http.addr=0.0.0.0 \
    --http.corsdomain=* \
    --http.api=admin,eth,net,web3,debug,trace,txpool \
    --ws \
    --ws.addr=0.0.0.0 \
    --ws.port=$HIVE_L2_WS_PORT \
    --ws.api=admin,eth,net,web3,debug,trace,txpool \
    --ws.origins=* \
    --authrpc.port=$HIVE_L2_AUTH_PORT \
    --authrpc.jwtsecret=/jwtsecret \
    --authrpc.addr=0.0.0.0 \
    --rpc.max-request-size=150 \
    --rpc.max-response-size=1600 \
    --rpc.max-subscriptions-per-connection=1024 \
    --rpc.max-connections=50000 \
    --rpc.max-tracing-requests=200 \
    --rpc.eth-proof-window=210000 \
    --metrics=0.0.0.0:9001


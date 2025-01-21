#!/bin/sh
set -exu

cat /wallets.json

private_key=$(cat /wallets.json | jq -r '."3151908".sequencerPrivateKey')

echo "private_key: $private_key"

op-node \
    --l2=$HIVE_L2_URL \
    --l2.jwt-secret=/jwtsecret \
    --verifier.l1-confs=4 \
    --rollup.config=/rollup-2151908.json \
    --rpc.addr=0.0.0.0 \
    --rpc.port=8547 \
    --rpc.enable-admin \
    --l1=$HIVE_L1_URL \
    --l1.rpckind=standard \
    --l1.beacon.ignore=true \
    --l1.trustrpc \
    --p2p.listen.ip=0.0.0.0 \
    --p2p.listen.tcp=9003 \
    --p2p.listen.udp=9003 \
    --safedb.path=/data/op-node/op-node-beacon-data \
    --metrics.enabled=true \
    --metrics.addr=0.0.0.0 \
    --metrics.port=9001 \
    --p2p.sequencer.key=$private_key \
    --sequencer.enabled \
    --sequencer.l1-confs=5



# op-node \
#     --l2=$HIVE_L2_URL \
#     --l2.jwt-secret=/jwtsecret \
#     --verifier.l1-confs=4 \
#     --rollup.config=/rollup-2151908.json \
#     --rpc.addr=0.0.0.0 \
#     --rpc.port=8547 \
#     --rpc.enable-admin \
#     --l1=$HIVE_L1_URL \
#     --l1.rpckind=standard \
#     --l1.beacon=http://172.16.0.11:4000 \
#     --l1.trustrpc \
#     --p2p.advertise.ip=172.16.0.22 \
#     --p2p.advertise.tcp=9003 \
#     --p2p.advertise.udp=9003 \
#     --p2p.listen.ip=0.0.0.0 \
#     --p2p.listen.tcp=9003 \
#     --p2p.listen.udp=9003 \
#     --safedb.path=/data/op-node/op-node-beacon-data \
#     --metrics.enabled=true \
#     --metrics.addr=0.0.0.0 \
#     --metrics.port=9001 \
#     --p2p.sequencer.key=$private_key \
#     --sequencer.enabled \
#     --sequencer.l1-confs=5
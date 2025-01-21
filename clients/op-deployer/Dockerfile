FROM mslipper/deployment-utils:latest AS deployment-utils

FROM us-docker.pkg.dev/oplabs-tools-artifacts/images/op-deployer:v0.0.8 AS base

RUN apt-get update && apt-get install -y jq && rm -rf /var/lib/apt/lists/*

RUN mkdir /network-data

ADD static_files/scripts /fund-script
RUN chmod +x /fund-script/fund.sh 

# Inject the startup script
ADD entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ADD intent.toml /intent.toml

# Helper scripts; called from devnet.go
ADD cat.sh /hive-bin/cat.sh
RUN chmod +x /hive-bin/cat.sh

COPY --from=deployment-utils /usr/local/bin/cast /usr/local/bin/cast
COPY --from=deployment-utils /usr/local/bin/dasel /usr/local/bin/dasel

ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]


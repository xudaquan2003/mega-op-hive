FROM ghcr.io/paradigmxyz/op-reth:latest AS base

RUN apt-get update && apt-get install -y jq

# Inject the startup script
ADD entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Create version.txt
RUN /usr/local/bin/op-reth --version | sed -e 's/reth-optimism-cli \(.*\)/\1/' > /version.txt

VOLUME ["/db"]

EXPOSE 8545/tcp
EXPOSE 8546/tcp
EXPOSE 8551/tcp

ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]


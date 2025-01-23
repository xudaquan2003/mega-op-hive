FROM megaeth-labs/testnet-all/op-reth:latest AS base

RUN apt-get update && apt-get install -y jq

# Inject the startup script
ADD entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Create version.txt
RUN /usr/local/bin/op-reth --version > /version.txt
# RUN echo "v0.0.0" > /version.txt

VOLUME ["/db"]

EXPOSE 9545/tcp
EXPOSE 9546/tcp
EXPOSE 9551/tcp

ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]


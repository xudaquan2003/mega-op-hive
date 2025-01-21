FROM us-docker.pkg.dev/oplabs-tools-artifacts/images/op-node:develop AS base

RUN apt-get update && apt-get install -y jq && rm -rf /var/lib/apt/lists/*

# Inject the startup script
ADD entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh


ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]
FROM us-docker.pkg.dev/oplabs-tools-artifacts/images/op-node:develop AS base

RUN apk add --no-cache jq 

RUN /usr/local/bin/op-node --version  > /version.txt

# Inject the startup script
ADD entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh


ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]
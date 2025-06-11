FROM alpine AS certs

RUN apk update && apk add --no-cache ca-certificates

FROM busybox:stable-musl

ARG TARGETOS
ARG TARGETARCH

COPY --from=certs /etc/ssl/certs /etc/ssl/certs

WORKDIR /data

COPY dist/torrentremover-${TARGETOS}-${TARGETARCH} ./torrentremover

VOLUME ["/data/config.yaml"]
ARG TZ=Etc/UTC
ENV TZ=$TZ
ENV DRY_RUN=false

ENTRYPOINT ["/bin/sh", "-c", "\
  CMD=\"/data/torrentremover -c /data/config.yaml\"; \
  if [ \"$DRY_RUN\" = \"true\" ]; then \
    CMD=\"$CMD -n\"; \
  fi; \
  exec $CMD \
"]

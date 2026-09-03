FROM alpine:3.22

RUN apk add --no-cache \
      ca-certificates \
      jq \
      openbao \
      postgresql17-client \
      restic \
      supercronic \
    && addgroup -S -g 10001 backup \
    && adduser -S -D -H -u 10001 -G backup backup

COPY --chmod=0755 bin/entrypoint /usr/local/bin/backup-entrypoint
COPY --chmod=0755 bin/backup-app-store /usr/local/bin/backup-app-store
COPY --chmod=0755 bin/backup-key-store /usr/local/bin/backup-key-store
COPY --chmod=0755 bin/backup-openbao /usr/local/bin/backup-openbao
COPY --chmod=0644 bin/backup-common /usr/local/lib/backup-common
COPY --chmod=0755 bin/healthcheck /usr/local/bin/backup-healthcheck
COPY backup.crontab /etc/supercronic/backup.crontab

ENV HOME=/tmp \
    RESTIC_CACHE_DIR=/tmp/restic-cache \
    TZ=UTC

USER backup

HEALTHCHECK --interval=5m --timeout=5s --start-period=30m --retries=2 \
  CMD ["/usr/local/bin/backup-healthcheck"]

ENTRYPOINT ["/usr/local/bin/backup-entrypoint"]

FROM debian:bookworm-slim AS builder

COPY _source /grt/

RUN tar -czf /grt/grt-data.tar.gz -C /grt GordonsReloadingTool.cfg GordonsReloadingTool.db loads || true

FROM ghcr.io/linuxserver/baseimage-kasmvnc:debianbookworm-version-876361b9@sha256:c6129530811450448ab760064b27e111fb3351fc3222af652f605a48eb518ed7

# Set environment variables
ENV CUSTOM_USER=gordon \
    DATA_DIR=/data \
    NO_DECOR=false \
    START_DOCKER=false \
    TITLE="Gordon's Reloading Tool"

RUN dpkg --add-architecture i386 && \
    apt-get update && \
    apt-get install --no-install-recommends -y \
    gnome-themes-extra:i386 \
    gtk2-engines-murrine:i386 \
    gtk2-engines-pixbuf:i386 \
    libc6:i386 \
    libcurl4:i386 \
    libgtk2.0-0:i386 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /tmp/* && \
    rm -rf /var/tmp/* && \
    rm -rf /usr/lib/locale/locale-archive

COPY root/ /

# We MUST copy the entire archive due to licensing
COPY _source /grt/

# Copy the pre-built tarball from builder stage
COPY --from=builder /grt/grt-data.tar.gz /grt/grt-data.tar.gz

RUN chmod +x /grt/GordonsReloadingTool && \
    chmod -R 777 /grt && \
    mkdir -p /data

# Set working directory
WORKDIR /data

# OCI annotations
LABEL org.opencontainers.image.title="docker-grt" \
    org.opencontainers.image.source="https://github.com/hugginsio/docker-grt" \
    org.opencontainers.image.licenses="BSD-3-Clause"

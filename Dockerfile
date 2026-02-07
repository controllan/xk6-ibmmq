FROM debian:13-slim

ENV DEBIAN_FRONTEND=noninteractive

# Minimal base tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    build-essential \
    pkg-config \
    libssl-dev \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

# Go 1.24.x (for k6 1.5+)
ENV GO_VERSION=1.24.11
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz && \
    tar -C /usr/local -xzf /tmp/go.tgz && \
    rm /tmp/go.tgz
ENV PATH=/usr/local/go/bin:${PATH}
ENV GOPATH=/go
ENV PATH=${GOPATH}/bin:${PATH}

ARG MQ_CLIENT_VERSION="9.4.4.1"
ARG MQ_CLIENT_TGZ_URL="https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqdev/redist/${MQ_CLIENT_VERSION}-IBM-MQC-Redist-LinuxX64.tar.gz"
ARG MQ_INSTALL_DIR=/opt/mqm

RUN mkdir -p /tmp/mq && \
    curl -fsSL "${MQ_CLIENT_TGZ_URL}" -o /tmp/mq/mqclient.tgz

RUN tar -C /tmp/mq -xzf /tmp/mq/mqclient.tgz && \
    rm /tmp/mq/mqclient.tgz && \
    mkdir -p "${MQ_INSTALL_DIR}"

RUN cp -R /tmp/mq/* "${MQ_INSTALL_DIR}/" && \
    rm -rf /tmp/mq

# MQ environment
ENV MQ_INSTALLATION_PATH=${MQ_INSTALL_DIR}
ENV LD_LIBRARY_PATH=${MQ_INSTALL_DIR}/lib64:${MQ_INSTALL_DIR}/lib:${LD_LIBRARY_PATH}
ENV LIBRARY_PATH=${MQ_INSTALL_DIR}/lib64:${MQ_INSTALL_DIR}/lib:${LIBRARY_PATH}
ENV C_INCLUDE_PATH=${MQ_INSTALL_DIR}/inc:${C_INCLUDE_PATH}
ENV PATH=${MQ_INSTALL_DIR}/bin:${PATH}

# CGO for mq-golang
ENV CGO_ENABLED=1

# xk6
RUN go install go.k6.io/xk6/cmd/xk6@latest

WORKDIR /workspace
CMD ["bash"]
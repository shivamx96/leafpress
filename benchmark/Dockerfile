# Multi-runtime Dockerfile for SSG benchmarking
# Includes: Go, Rust, Ruby, Node.js and all SSGs

FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# Install base dependencies
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    build-essential \
    time \
    && rm -rf /var/lib/apt/lists/*

# Install Go (for Hugo and Leafpress)
ENV GO_VERSION=1.21.5
ARG TARGETARCH
RUN if [ "$TARGETARCH" = "arm64" ]; then \
      curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-arm64.tar.gz | tar -xz -C /usr/local; \
    else \
      curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz | tar -xz -C /usr/local; \
    fi
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/root/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# Install Rust (for Zola)
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"

# Install Ruby (for Jekyll)
RUN apt-get update && apt-get install -y ruby-full ruby-bundler && rm -rf /var/lib/apt/lists/*

# Install Node.js (for Eleventy, Astro)
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install Hugo
ENV HUGO_VERSION=0.121.1
RUN if [ "$TARGETARCH" = "arm64" ]; then \
      wget -qO- https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_extended_${HUGO_VERSION}_linux-arm64.tar.gz | tar -xz -C /usr/local/bin hugo; \
    else \
      wget -qO- https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_extended_${HUGO_VERSION}_linux-amd64.tar.gz | tar -xz -C /usr/local/bin hugo; \
    fi

# Install Zola
ENV ZOLA_VERSION=0.21.0
RUN if [ "$TARGETARCH" = "arm64" ]; then \
      wget -qO- https://github.com/getzola/zola/releases/download/v${ZOLA_VERSION}/zola-v${ZOLA_VERSION}-aarch64-unknown-linux-gnu.tar.gz | tar -xz -C /usr/local/bin; \
    else \
      wget -qO- https://github.com/getzola/zola/releases/download/v${ZOLA_VERSION}/zola-v${ZOLA_VERSION}-x86_64-unknown-linux-gnu.tar.gz | tar -xz -C /usr/local/bin; \
    fi

# Install Jekyll
RUN gem install jekyll bundler --no-document

# Install Node.js SSGs globally
RUN npm install -g @11ty/eleventy

# Create working directory
WORKDIR /benchmark

# Copy benchmark files
COPY . /benchmark/

# Make scripts executable
RUN chmod +x /benchmark/*.sh /benchmark/generators/*/*.sh 2>/dev/null || true

# Copy and build Leafpress from source (mounted at runtime, but we need a startup script)
COPY docker-entrypoint.sh /benchmark/docker-entrypoint.sh
RUN chmod +x /benchmark/docker-entrypoint.sh

# Default command
ENTRYPOINT ["/benchmark/docker-entrypoint.sh"]
CMD ["/benchmark/run-all.sh"]

# Multi-runtime Dockerfile for SSG benchmarking
# Includes: Go, Ruby, Node.js and all benchmarked SSGs

FROM ubuntu:22.04@sha256:3b06811b2afd352be909dd088a004166d665dc76d38b13eada33522a9d915c6f

ENV DEBIAN_FRONTEND=noninteractive

# Install base dependencies
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    git \
    build-essential \
    python3 \
    time \
    xz-utils \
    && rm -rf /var/lib/apt/lists/*

# Install Go (for Hugo and Leafpress)
ENV GO_VERSION=1.25.5
ARG TARGETARCH=amd64
RUN case "$TARGETARCH" in \
      arm64) archive="go${GO_VERSION}.linux-arm64.tar.gz"; checksum="b00b694903d126c588c378e72d3545549935d3982635ba3f7a964c9fa23fe3b9" ;; \
      amd64) archive="go${GO_VERSION}.linux-amd64.tar.gz"; checksum="9e9b755d63b36acf30c12a9a3fc379243714c1c6d3dd72861da637f336ebb35b" ;; \
      *) echo "unsupported architecture: $TARGETARCH" >&2; exit 1 ;; \
    esac \
    && curl -fsSL "https://go.dev/dl/${archive}" -o "/tmp/${archive}" \
    && echo "${checksum}  /tmp/${archive}" | sha256sum -c - \
    && tar -xzf "/tmp/${archive}" -C /usr/local \
    && rm "/tmp/${archive}"
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/root/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# Install Ruby (for Jekyll)
RUN apt-get update && apt-get install -y ruby-full && rm -rf /var/lib/apt/lists/*

# Install Node.js (for Eleventy, Astro)
ENV NODE_VERSION=20.19.5
RUN case "$TARGETARCH" in \
      arm64) node_arch="arm64"; checksum="d462267863ae8ee556039ebdf559055a8ec562c633889ef1403f3adb449ba1dd" ;; \
      amd64) node_arch="x64"; checksum="315046739a513a70e03a4a55a8afda8cf979f30852e576075c340084e3f8ac0f" ;; \
      *) echo "unsupported architecture: $TARGETARCH" >&2; exit 1 ;; \
    esac \
    && archive="node-v${NODE_VERSION}-linux-${node_arch}.tar.xz" \
    && curl -fsSL "https://nodejs.org/dist/v${NODE_VERSION}/${archive}" -o "/tmp/${archive}" \
    && echo "${checksum}  /tmp/${archive}" | sha256sum -c - \
    && tar -xJf "/tmp/${archive}" -C /usr/local --strip-components=1 \
    && rm "/tmp/${archive}"

# Install Hugo
ENV HUGO_VERSION=0.121.1
RUN case "$TARGETARCH" in \
      arm64) hugo_arch="arm64"; checksum="286dde6af51c21db6d196fa5ccbc5956c4d02984fd7ff83bc1cc1a1fc9acfad0" ;; \
      amd64) hugo_arch="amd64"; checksum="762ddf0b3507284f8a046299fe0dea8fe5fe8b8ccd0c9e68b1604625822501fd" ;; \
      *) echo "unsupported architecture: $TARGETARCH" >&2; exit 1 ;; \
    esac \
    && archive="hugo_extended_${HUGO_VERSION}_linux-${hugo_arch}.tar.gz" \
    && curl -fsSL "https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/${archive}" -o "/tmp/${archive}" \
    && echo "${checksum}  /tmp/${archive}" | sha256sum -c - \
    && tar -xzf "/tmp/${archive}" -C /usr/local/bin hugo \
    && rm "/tmp/${archive}"

# Install Zola
ENV ZOLA_VERSION=0.21.0
RUN case "$TARGETARCH" in \
      arm64) zola_arch="aarch64"; checksum="da7e9def7b9acf0c9b78fb99ce72624504cfee1066ab7a161f09a4126dd7ffe6" ;; \
      amd64) zola_arch="x86_64"; checksum="5c37a8f706567d6cad3f0dbc0eaebe3b9591cc301bd67089e5ddc0d0401732d6" ;; \
      *) echo "unsupported architecture: $TARGETARCH" >&2; exit 1 ;; \
    esac \
    && archive="zola-v${ZOLA_VERSION}-${zola_arch}-unknown-linux-gnu.tar.gz" \
    && curl -fsSL "https://github.com/getzola/zola/releases/download/v${ZOLA_VERSION}/${archive}" -o "/tmp/${archive}" \
    && echo "${checksum}  /tmp/${archive}" | sha256sum -c - \
    && tar -xzf "/tmp/${archive}" -C /usr/local/bin zola \
    && rm "/tmp/${archive}"

# Install Jekyll
RUN gem install jekyll -v 4.4.1 --no-document \
    && jekyll --version

# Install Node.js SSGs globally
RUN npm install -g @11ty/eleventy@3.1.2 \
    && eleventy --version

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

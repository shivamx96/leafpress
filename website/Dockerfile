FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy cli source and build leafpress
COPY cli/ ./cli/
RUN cd cli && go build -o /usr/local/bin/leafpress ./cmd/leafpress

# Copy website source and build
COPY website/ ./website/
RUN cd website && leafpress build

FROM nginx:alpine

COPY --from=builder /build/website/_site /usr/share/nginx/html
COPY website/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

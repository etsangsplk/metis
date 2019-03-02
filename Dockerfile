FROM golang:alpine as builder
ENV GOFLAGS=-mod=vendor GO111MODULE=on CGO_ENABLED=0 GOOS=linux 
RUN mkdir -p /build
ADD . /build/
WORKDIR /build 
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o metis-store ./cmd/metis-store/*.go

FROM alpine
RUN mkdir -p /var/metis
RUN wget -O /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_amd64 \
	&& chmod +x /usr/local/bin/dumb-init
COPY --from=builder /build/metis-store /usr/bin/
WORKDIR /var/metis
VOLUME /var/metis/data
ENTRYPOINT [ "/usr/local/bin/dumb-init", "--", "/usr/bin/metis-store" ]
CMD [ "-data-dir", "/var/metis/data" ]
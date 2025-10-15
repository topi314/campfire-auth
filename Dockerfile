FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build

WORKDIR /build

COPY go.mod ./

RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build -o campfire-auth github.com/topi314/campfire-auth

FROM alpine

COPY --from=build /build/campfire-auth /bin/campfire-auth

ENTRYPOINT ["/bin/campfire-auth"]

CMD ["-config", "/var/lib/config.toml"]

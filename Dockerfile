FROM golang:alpine as build

WORKDIR /opt

# Copy dependencies list
COPY go.mod go.sum ./

#install WebP
RUN apk add --no-cache libwebp libwebp-dev

# Build
COPY main.go .
RUN go build -o main main.go

# Copy artifacts to a clean image
FROM alpine:latest
COPY --from=build /opt/main /main

RUN apk add --no-cache libwebp git
RUN git clone https://github.com/acvigue/TidbytCommunity

ENTRYPOINT [ "/main" ]
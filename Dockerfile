FROM golang:latest as build

WORKDIR /opt

# Copy dependencies list
COPY go.mod go.sum ./

#install WebP
RUN apt update && \
    apt upgrade && \
    apt install -y libwebp-dev git

#download community apps
RUN git clone https://github.com/acvigue/TidbytCommunity

RUN find . -name "*.gif" -type f -delete
RUN find . -name "*.jpg" -type f -delete
RUN find . -name "*.jpeg" -type f -delete
RUN find . -name "*.png" -type f -delete
RUN find . -name "*.svg" -type f -delete
RUN find . -name "*.webp" -type f -delete

RUN go mod download

# Build
COPY *.go ./
RUN CGO_LDFLAGS="-Wl,-Bstatic -lwebp -lwebpdemux -lwebpmux -Wl,-Bdynamic" CGO_ENABLED=1 go build -o main

# Copy artifacts to a clean image
FROM gcr.io/distroless/base

COPY --from=build /opt/main /main
COPY --from=build /opt/TidbytCommunity/apps /apps

ENV GIN_MODE=release
ENV APPS_PATH=/apps/

ENTRYPOINT [ "/main" ]
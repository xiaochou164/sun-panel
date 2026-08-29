# build frontend
FROM node:18-alpine AS web_image

# 国内 npm 源
RUN npm config set registry https://registry.npmmirror.com

# pnpm 9（兼容项目的 pnpm-lock.yaml v6，且不强制 approve-builds）
RUN npm install -g pnpm@9

WORKDIR /build

COPY ./package.json /build
COPY ./pnpm-lock.yaml /build

RUN pnpm install --registry=https://registry.npmmirror.com

COPY . /build

RUN pnpm run build

# build backend
FROM golang:1.24-alpine3.21 as server_image

WORKDIR /build

COPY ./service .

# 国内源
RUN sed -i "s@dl-cdn.alpinelinux.org@mirrors.aliyun.com@g" /etc/apk/repositories \
    && apk add --no-cache bash curl gcc git musl-dev \
    && go env -w GOPROXY=https://goproxy.cn,direct

RUN go env -w GO111MODULE=on \
    && export PATH=$PATH:$(go env GOPATH)/bin \
    && go install github.com/go-bindata/go-bindata/...@latest \
    && go install github.com/elazarl/go-bindata-assetfs/...@v1.0.1 \
    && go-bindata-assetfs -o=assets/bindata.go -pkg=assets assets/... \
    && go build -o sun-panel --ldflags="-X sun-panel/global.RUNCODE=release -X sun-panel/global.ISDOCKER=docker" main.go

# run_image
FROM alpine

WORKDIR /app

COPY --from=web_image /build/dist /app/web

COPY --from=server_image /build/sun-panel /app/sun-panel

EXPOSE 3002

RUN sed -i "s@dl-cdn.alpinelinux.org@mirrors.aliyun.com@g" /etc/apk/repositories \
    && apk add --no-cache bash ca-certificates su-exec tzdata \
    && chmod +x ./sun-panel \
    && ./sun-panel -config

CMD ./sun-panel

# server builder

FROM golang:1.23 AS server_builder

ENV APP_HOME=/code/bbs-go/server
WORKDIR "$APP_HOME"

COPY ./server ./
RUN go env -w GOPROXY=https://goproxy.cn,direct \
    && go mod download
RUN go mod download

RUN CGO_ENABLED=0 go build -v -o bbs-go main.go && chmod +x bbs-go


# site builder
FROM node:20 AS site_builder

ENV APP_HOME=/code/bbs-go/site
WORKDIR "$APP_HOME"

COPY ./site ./
# RUN npm install -g pnpm --registry=https://registry.npmmirror.com
# RUN pnpm install --registry=https://registry.npmmirror.com
RUN npm install -g pnpm
RUN pnpm install
RUN pnpm build:docker

# admin builder
FROM node:20-alpine AS admin_builder

ENV APP_HOME=/code/bbs-go/admin
WORKDIR "$APP_HOME"

# 安装 jpegtran 所需库
# RUN apk add --no-cache libjpeg-turbo-utils

# 安装autoconf依赖
RUN apt update && apt add -y autoconf automake libtool

COPY ./admin ./
# RUN npm install -g pnpm --registry=https://registry.npmmirror.com
# RUN pnpm install --registry=https://registry.npmmirror.com
RUN npm install -g pnpm
RUN pnpm install
RUN pnpm build:docker

# run
FROM node:20-alpine

ENV APP_HOME=/app/bbs-go
WORKDIR "$APP_HOME"


COPY --from=server_builder /code/bbs-go/server/bbs-go ./server/bbs-go
COPY --from=server_builder /code/bbs-go/server/*.yaml ./server/
COPY --from=server_builder /code/bbs-go/server/*.yml ./server/
COPY --from=site_builder /code/bbs-go/site/.output ./site/.output
COPY --from=site_builder /code/bbs-go/site/node_modules ./site/node_modules
COPY --from=admin_builder /code/bbs-go/admin/dist ./admin

COPY ./start.sh ${APP_HOME}/start.sh
RUN chmod +x ${APP_HOME}/start.sh

EXPOSE 8082
EXPOSE 3000

ENV ENV=docker
CMD ["./start.sh"]

FROM golang:alpine AS golang-builder
ENV GO111MODULE=on

COPY ./api-server /yatai/src/api-server
COPY ./common /yatai/src/common
COPY ./schemas /yatai/src/schemas

COPY go.mod go.sum /yatai/src

WORKDIR /yatai/src

RUN go build -o /go/bin/yatai-api-server ./api-server/main.go


FROM node:14.16.1-alpine as node-builder

WORKDIR /app

ENV PATH /app/node_modules/.bin:$PATH

COPY dashboard/package.json ./

COPY dashboard ./

RUN yarn

RUN yarn build


FROM scratch

COPY ./statics /statics
COPY ./yatai-config.production.yaml /yatai-config.production.yaml

COPY --from=golang-builder /etc/passwd /etc/passwd
COPY --from=golang-builder /go/bin/yatai-api-server /bin/yatai-api-server

COPY --from=node-builder /app/build /dashboard

EXPOSE 3000 7777

ENTRYPOINT ["/bin/yatai-api-server serve -d -c /yatai-config.production.yaml"]

# tags to serve front end build
# RUN ["/bin/yatai-api-server", "serve", "--frontend", "/dashboard"]

FROM golang:1.17

RUN mkdir /app

RUN chmod a+x /app/api-server
RUN mkdir -p /app/dashboard
RUN mkdir -p /app/scripts
COPY ./statics /app/statics
COPY ./dashboard/build /app/dashboard/build
COPY ./api-server/db /app/db
COPY ./bin/api-server /app/api-server

WORKDIR /app

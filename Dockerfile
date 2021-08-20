FROM golang:1.17

RUN mkdir /app

COPY ./bin/api-server /app/api-server
RUN chmod a+x /app/api-server
RUN mkdir -p /app/dashboard
COPY ./dashboard/build /app/dashboard/build
COPY ./api-server/db /app/db
ENV MIGRATION_DIR=/app/db/migrations

WORKDIR /app

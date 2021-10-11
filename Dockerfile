FROM golang:1.17

RUN mkdir /app

COPY ./bin/api-server /app/api-server
RUN chmod a+x /app/api-server
RUN mkdir -p /app/dashboard
RUN mkdir -p /app/scripts
COPY ./scripts/helm-charts /app/scripts/helm-charts
COPY ./dashboard/build /app/dashboard/build
COPY ./api-server/db /app/db
COPY ./statics /app/statics

WORKDIR /app

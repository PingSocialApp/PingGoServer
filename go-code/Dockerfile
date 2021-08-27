ARG fbadmin
ARG neo4jpass

FROM golang:alpine AS build

WORKDIR /tmp/app

COPY src/ .
RUN go mod download

RUN GOOS=linux go build -o ./out/api .

FROM alpine:latest
COPY --from=build /tmp/app/out/api /app/pingserver
COPY src/public/ /app/public/

WORKDIR /app

ARG fbadmin
ARG neo4jpass

ENV CLOUD_DEV_URL=neo4j+s://57b3b322.databases.neo4j.io
ENV CLOUD_DEV_USER=neo4j
ENV ADMIN_SDK=$fbadmin
ENV CLOUD_DEV_PASS=$neo4jpass

EXPOSE $PORT

CMD [ "./pingserver",  "-cloud", "-localDev=false"]



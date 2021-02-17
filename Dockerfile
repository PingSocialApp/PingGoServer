FROM golang:alpine

MAINTAINER sreegrandhe@gmail.com

ENV HOST="localhost"
ENV PORT=8080
ENV DB_HOST="bolt://localhost:7687"
ENV DB_USER="neo4j"
ENV DB_PASS="pingdev"
ENV FIREBASE_ADMIN="./circles-4d081-firebase-adminsdk-rtjsi-51616d71b7.json"

RUN mkdir /app
COPY src/ app/
COPY go.mod ./
COPY go.sum ./
COPY $FIREBASE_ADMIN ./

RUN go mod download
RUN go build

EXPOSE $PORT
ENTRYPOINT ["./app"]
FROM golang:latest

LABEL CREATOR="sreegrandhe@gmail.com"

ENV HOST="localhost"
ENV PORT=8080
ENV GO111MODULE=on
ENV DB_HOST="bolt://0.0.0.0:7687"
ENV DB_USER="neo4j"
ENV DB_PASS="pingdev"
ENV FIREBASE_ADMIN="./circles-4d081-firebase-adminsdk-rtjsi-51616d71b7.json"

WORKDIR /app
COPY src/go.mod .
COPY src/go.sum .
COPY $FIREBASE_ADMIN .

RUN go mod download

COPY src/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

EXPOSE $PORT
ENTRYPOINT go run .
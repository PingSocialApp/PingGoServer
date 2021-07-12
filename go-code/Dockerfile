FROM golang:alpine AS build

WORKDIR /tmp/app

COPY src/ .
RUN go mod download

RUN GOOS=linux go build -o ./out/api .

FROM alpine:latest
 
COPY --from=build /tmp/app/out/api /app/pingserver
COPY src/.env /app/
COPY src/public/ /app/public/

WORKDIR /app

EXPOSE 8080

CMD [ "./pingserver",  "-cloud"]



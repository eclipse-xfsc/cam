FROM golang:1.19 as build

WORKDIR /usr/src/app

COPY . . 

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server cmd/cam-api-gateway/cam-api-gateway.go

FROM scratch as run
COPY --from=build /server .
CMD ["./server"]

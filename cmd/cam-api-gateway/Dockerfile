FROM node:20 AS frontend
WORKDIR /app
COPY ./dashboard ./dashboard
RUN bash -c "pushd dashboard && npm install && npm run build && popd"

FROM golang:1.19 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/dashboard/dist ./dashboard/dist
RUN go build -o server cmd/cam-api-gateway/cam-api-gateway.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
ENTRYPOINT ["./server"]
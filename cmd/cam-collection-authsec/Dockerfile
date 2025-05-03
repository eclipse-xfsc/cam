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
RUN go build -o server cmd/cam-collection-authsecy/cam-collection-authsec.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
# The ingress is not sending the intermediate certificate so we need to bring it along
ADD third_party/rapidssl.crt /usr/local/share/ca-certificates/rapidssl.crt
RUN chmod 644 /usr/local/share/ca-certificates/rapidssl.crt && update-ca-certificates

ENTRYPOINT ["./server"]
FROM golang:1.19
WORKDIR /usr/src/app

COPY policies policies
COPY metrics.json metrics.json
COPY gxfs.json gxfs.json

COPY cam-req-manager cam-req-manager
CMD ["./cam-req-manager"]

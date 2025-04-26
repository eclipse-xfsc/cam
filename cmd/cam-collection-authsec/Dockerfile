# auth-sec is the build stage for creating the workload configuration collection module
FROM golang:1.19
WORKDIR /usr/src/app

COPY cam-collection-authsec cam-collection-authsec

# The ingress is not sending the intermediate certificate so we need to bring it along
ADD third_party/rapidssl.crt /usr/local/share/ca-certificates/rapidssl.crt
RUN chmod 644 /usr/local/share/ca-certificates/rapidssl.crt && update-ca-certificates

CMD ["./cam-collection-authsec"]


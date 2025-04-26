FROM golang:1.19
WORKDIR /usr/src/app
COPY policies/operators.rego policies/operators.rego

COPY cam-eval-manager cam-eval-manager
CMD ["./cam-eval-manager"]

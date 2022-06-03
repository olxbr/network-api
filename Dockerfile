FROM public.ecr.aws/bitnami/golang:1.16 as build-image

WORKDIR /go/src/
ADD . /go/src/

WORKDIR /go/src/api
RUN go get -d -v ./...
RUN CGO_ENABLED=0 make build

FROM public.ecr.aws/lambda/go:1

COPY --from=build-image /go/src/api/bin/network-api /var/task/

CMD ["network-api"]
FROM ubuntu:latest

WORKDIR /build

COPY ./build/knative-agent .

ENTRYPOINT [ "./knative-agent" ]
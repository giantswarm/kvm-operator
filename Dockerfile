FROM alpine:3.7

RUN apk add --update ca-certificates \
    && rm -rf /var/cache/apk/*

RUN mkdir -p /opt/ignition
ADD vendor/github.com/giantswarm/k8scloudconfig/ /opt/ignition

ADD ./kvm-operator /kvm-operator

ENTRYPOINT ["/kvm-operator"]

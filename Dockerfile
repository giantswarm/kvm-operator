FROM alpine:3.7

RUN apk add --update ca-certificates \
    && rm -rf /var/cache/apk/*

ADD ./kvm-operator /kvm-operator

ENTRYPOINT ["/kvm-operator"]

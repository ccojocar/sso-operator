FROM golang:1.14.2

COPY . /go/src/github.com/jenkins-x/sso-operator
WORKDIR /go/src/github.com/jenkins-x/sso-operator
RUN make build

FROM scratch
COPY --from=0 /go/src/github.com/jenkins-x/sso-operator/bin/sso-operator /sso-operator
EXPOSE 8080
ENTRYPOINT ["/sso-operator"]

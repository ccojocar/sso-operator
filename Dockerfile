FROM golang:1.12.6

COPY . /go/src/github.com/jenkins-x/sso-operator
WORKDIR /go/src/github.com/jenkins-x/sso-operator
RUN make all

FROM scratch
COPY --from=0 /go/src/github.com/jenkins-x/sso-operator/bin/sso-operator /sso-operator
EXPOSE 8080
ENTRYPOINT ["/sso-operator"]

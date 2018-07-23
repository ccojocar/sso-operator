FROM scratch
EXPOSE 8080
ENTRYPOINT ["/sso-operator"]
COPY ./bin/ /
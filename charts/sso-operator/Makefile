CHART_REPO := http://jenkins-x-chartmuseum:8080
CURRENT=$(pwd)
NAME := sso-operator
OS := $(shell uname)
VERSION := $(shell cat ../../VERSION)

init: 
	helm init --client-only

setup: init
	helm repo add jenkins-x-api https://chartmuseum.build.cd.jenkins-x.io 	
	helm repo add jenkinsxio https://chartmuseum.jx.cd.jenkins-x.io 

build: clean setup
	helm dependency build
	helm lint

install: clean setup build
	helm install . --name ${NAME}

upgrade: clean setup build
	helm upgrade ${NAME} .

delete:
	helm delete --purge ${NAME}

clean:
	rm -rf charts
	rm -rf ${NAME}*.tgz
	rm -rf requirements.lock

release: clean build
ifeq ($(OS),Darwin)
	sed -i "" -e "s/version:.*/version: $(VERSION)/" Chart.yaml
	sed -i "" -e "s/tag:.*/tag: $(VERSION)/" values.yaml

else ifeq ($(OS),Linux)
	sed -i -e "s/version:.*/version: $(VERSION)/" Chart.yaml
	sed -i -e "s/tag:.*/tag: $(VERSION)/" values.yaml
else
	exit -1
endif
	helm package .
	curl --fail -u $(CHARTMUSEUM_CREDS_USR):$(CHARTMUSEUM_CREDS_PSW) --data-binary "@$(NAME)-$(VERSION).tgz" $(CHART_REPO)/api/charts
	helm repo update
	rm -rf ${NAME}*.tgz

CHART_REPO := http://jenkins-x-chartmuseum:8080
CURRENT=$(pwd)
NAME := sso-operator
OS := $(shell uname)
CHARTMUSEUM_CREDS_USR := $(shell cat /builder/home/basic-auth-user 2> /dev/null)
CHARTMUSEUM_CREDS_PSW := $(shell cat /builder/home/basic-auth-pass 2> /dev/null)

init: 
	helm init --client-only

setup: init
	helm repo add jenkinsx http://chartmuseum.jenkins-x.io 	

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

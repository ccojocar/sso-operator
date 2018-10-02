pipeline {
    agent {
        label "jenkins-go"
    }
    environment {
      DOCKER_REGISTRY   = 'docker.io'
      ORG               = 'jenkinsxio'
      APP_NAME          = 'sso-operator'
      GIT_PROVIDER      = 'github.com'
      CHARTMUSEUM_CREDS = credentials('jenkins-x-chartmuseum')
    }
    stages {
      stage('CI Build and push snapshot') {
        when {
          branch 'PR-*'
        }
        environment {
          PREVIEW_VERSION = "0.0.0-SNAPSHOT-$BRANCH_NAME-$BUILD_NUMBER"
          PREVIEW_NAMESPACE = "$APP_NAME-$BRANCH_NAME".toLowerCase()
          HELM_RELEASE = "$PREVIEW_NAMESPACE".toLowerCase()
        }
        steps {
          dir ('/home/jenkins/go/src/github.com/jenkins-x/sso-operator') {
            checkout scm
            container('go') {
              sh "make all"
              sh 'export VERSION=$PREVIEW_VERSION && skaffold build -f skaffold.yaml'


              sh "jx step post build --image $DOCKER_REGISTRY/$ORG/$APP_NAME:$PREVIEW_VERSION"
            }
          }
        }
      }
      stage('Build Release') {
        environment {
          CHARTMUSEUM_CREDS = credentials('jenkins-x-chartmuseum')
        }
        when {
          branch 'master'
        }
        steps {
          container('go') {
            dir ('/home/jenkins/go/src/github.com/jenkins-x/sso-operator') {
              checkout scm
            }
            dir ('/home/jenkins/go/src/github.com/jenkins-x/sso-operator/charts/sso-operator') {
                // ensure we're not on a detached head
                sh "git checkout master"
                // until we switch to the new kubernetes / jenkins credential implementation use git credentials store
                sh "git config --global credential.helper store"

                sh "jx step git credentials"
            }
            dir ('/home/jenkins/go/src/github.com/jenkins-x/sso-operator') {
              // so we can retrieve the version in later steps
              sh "echo \$(jx-release-version) > VERSION"
            }
            dir ('/home/jenkins/go/src/github.com/jenkins-x/sso-operator') {
              container('go') {
                sh "make all"
                sh 'export VERSION=`cat VERSION` && skaffold build -f skaffold.yaml'

                sh "jx step tag --version \$(cat VERSION)"
                sh "jx step post build --image $DOCKER_REGISTRY/$ORG/$APP_NAME:\$(cat VERSION)"
              }
            }
            dir ('/home/jenkins/go/src/github.com/jenkins-x/sso-operator/charts/sso-operator') {
              sh "make release"
            }
          }
        }
      }
    }
    post {
        always {
            cleanWs()
        }
    }
  }

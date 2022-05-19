pipeline {
  agent {
    label 'nimbus-cloud'
  }
  environment {
    GITHUB_CREDS_PSW = credentials("GITHUB_TOKEN")
  }

  stages {
    stage("Test with Go 1.18") {
      tools {
        go 'Go 1.18'
      }
      steps {
        withEnv(["PATH+EXTRA=${HOME}/go/bin"]) {
          sh 'make checkfmt vet tests'
        }
      }
    }

    stage("Build and publish") {
      parallel{
        stage("Test Openshift build") {
          steps {
            sh 'docker build -f deploy/docker/Dockerfile-rhel .'
          }
        }
        stage("Publish") {
          tools {
            go 'Go 1.18'
          }
          environment {
            RELEASE_TYPE = "alpha"
            VERSION_POSTFIX = "-alpha-${GIT_COMMIT.substring(0, 8)}"
            HARBOR_CREDS = credentials("projects-registry-vmware-tanzu_observability_keights_saas-robot")
            PREFIX = "projects.registry.vmware.com/tanzu_observability_keights_saas"
            DOCKER_IMAGE = "kubernetes-collector-snapshot"
          }
          steps {
            withEnv(["PATH+EXTRA=${HOME}/go/bin"]) {
              sh './hack/jenkins/install_docker_buildx.sh'
              sh 'make semver-cli'
              sh 'echo $HARBOR_CREDS_PSW | docker login $PREFIX -u $HARBOR_CREDS_USR --password-stdin'
              sh 'HARBOR_CREDS_USR=$(echo $HARBOR_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'
            }
          }
        }
      }
    }

    stage('Run Integration Tests') {
      parallel {
        stage("GKE Integration Test") {
          agent {
            label "gke"
          }
          options {
            timeout(time: 20, unit: 'MINUTES')
          }
          tools {
            go 'Go 1.18'
          }
          environment {
            GCP_CREDS = credentials("GCP_CREDS")
            GKE_CLUSTER_NAME = "k8po-jenkins-ci-collector"
            VERSION_POSTFIX = "-alpha-${GIT_COMMIT.substring(0, 8)}"
            PREFIX = "projects.registry.vmware.com/tanzu_observability_keights_saas"
            DOCKER_IMAGE = "kubernetes-collector-snapshot"
            WAVEFRONT_TOKEN = credentials("WAVEFRONT_TOKEN_NIMBA")
          }
          steps {
            withEnv(["PATH+GO=${HOME}/go/bin", "PATH+GCLOUD=${HOME}/google-cloud-sdk/bin"]) {
              lock("integration-test-gke-collector") {
                sh './hack/jenkins/setup-for-integration-test.sh -k gke'
                sh 'make gke-connect-to-cluster'
                sh 'VERSION_POSTFIX=$VERSION_POSTFIX INTEGRATION_TEST_TYPE=cluster-metrics-only make deploy-test'
                sh 'VERSION_POSTFIX=$VERSION_POSTFIX INTEGRATION_TEST_TYPE=node-metrics-only make deploy-test'
                sh 'VERSION_POSTFIX=$VERSION_POSTFIX INTEGRATION_TEST_TYPE=combined make deploy-test'
//                 sh 'VERSION_POSTFIX=$VERSION_POSTFIX INTEGRATION_TEST_TYPE=single-deployment make deploy-test'
                sh 'VERSION_POSTFIX=$VERSION_POSTFIX make deploy-test'
              }
            }
          }
        }
        stage("EKS Integration Test") {
          agent {
            label "eks"
          }
          options {
            timeout(time: 20, unit: 'MINUTES')
          }
          tools {
            go 'Go 1.18'
          }
          environment {
            VERSION_POSTFIX = "-alpha-${GIT_COMMIT.substring(0, 8)}"
            PREFIX = "projects.registry.vmware.com/tanzu_observability_keights_saas"
            DOCKER_IMAGE = "kubernetes-collector-snapshot"
            AWS_SHARED_CREDENTIALS_FILE = credentials("k8po-ci-aws-creds")
            AWS_CONFIG_FILE = credentials("k8po-ci-aws-profile")
            WAVEFRONT_TOKEN = credentials("WAVEFRONT_TOKEN_NIMBA")
          }
          steps {
            withEnv(["PATH+GO=${HOME}/go/bin"]) {
              lock("integration-test-eks") {
                sh './hack/jenkins/setup-for-integration-test.sh -k eks'

                sh 'make target-eks'
//                 sh 'VERSION_POSTFIX=$VERSION_POSTFIX INTEGRATION_TEST_TYPE=cluster-metrics-only make deploy-test'
//                 sh 'VERSION_POSTFIX=$VERSION_POSTFIX INTEGRATION_TEST_TYPE=node-metrics-only make deploy-test'
                sh 'VERSION_POSTFIX=$VERSION_POSTFIX INTEGRATION_TEST_TYPE=combined make deploy-test'
                sh 'VERSION_POSTFIX=$VERSION_POSTFIX INTEGRATION_TEST_TYPE=single-deployment make deploy-test'
                sh 'VERSION_POSTFIX=$VERSION_POSTFIX make deploy-test'
                sh './hack/test/test-wavefront-metrics.sh -t $WAVEFRONT_TOKEN'
              }
            }
          }
        }
      }
    }
  }

  post {
    // Notify only on null->failure or success->failure or failure->success
    failure {
      script {
        if(currentBuild.previousBuild == null) {
          slackSend (channel: '#tobs-k8po-team', color: '#FF0000', message: "CI BUILD FAILED: <${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>")
        }
      }
    }
    regression {
      slackSend (channel: '#tobs-k8po-team', color: '#FF0000', message: "CI BUILD FAILED: <${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>")
    }
    fixed {
      slackSend (channel: '#tobs-k8po-team', color: '#008000', message: "CI BUILD FIXED: <${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>")
    }
    success {
      script {
        if (env.BRANCH_NAME == 'main') {
          sh './hack/jenkins/update_github_status.sh'
        }
      }
    }
  }
}
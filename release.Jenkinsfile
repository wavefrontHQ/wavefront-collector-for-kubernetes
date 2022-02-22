pipeline {
    agent any

    tools {
        go 'Go 1.17'
    }

    environment {
        RELEASE_TYPE = "${params.RELEASE_TYPE}"
        RC_NUMBER = "1"
        BUMP_COMPONENT = "${params.BUMP_COMPONENT}"
        GIT_BRANCH = getCurrentBranchName()
        GIT_CREDENTIAL_ID = 'wf-jenkins-github'
        TOKEN = credentials('GITHUB_TOKEN')
    }

    stages {
      stage("Setup tools") {
        steps {
          withEnv(["PATH+EXTRA=${HOME}/go/bin"]) {
            sh './hack/jenkins/install_docker_buildx.sh'
            sh 'make semver-cli'
          }
        }
      }

      stage("Create Bump Version Branch") {
        steps {
          withEnv(["PATH+EXTRA=${HOME}/go/bin"]){
            sh 'git config --global user.email "svc.wf-jenkins@vmware.com"'
            sh 'git config --global user.name "svc.wf-jenkins"'
            sh 'git remote set-url origin https://${TOKEN}@github.com/wavefronthq/wavefront-collector-for-kubernetes.git'

            sh './hack/jenkins/create-bump-version-branch.sh "${BUMP_COMPONENT}"'
          }
        }
      }

      stage("Publish RC Release") {
        environment {
          HARBOR_CREDS = credentials("projects-registry-vmware-tanzu_observability-robot")
          PREFIX = 'projects.registry.vmware.com/tanzu_observability'
          DOCKER_IMAGE = 'kubernetes-collector'
          RELEASE_TYPE = 'rc'
        }
        steps {
          sh 'echo $HARBOR_CREDS_PSW | docker login $PREFIX -u $HARBOR_CREDS_USR --password-stdin'
          sh 'HARBOR_CREDS_USR=$(echo $HARBOR_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'
        }
      }
      // deploy to GKE and run manual tests
      // now we have confidence in the validity of our RC release
      stage("Deploy and Test") {
        environment {
          GCP_CREDS = credentials("GCP_CREDS")
          GKE_CLUSTER_NAME = "k8po-jenkins-rc-testing"
          WAVEFRONT_TOKEN = credentials("WAVEFRONT_TOKEN_NIMBA")
          WF_CLUSTER = 'nimba'
          RELEASE_TYPE = 'rc'
        }
        steps {
          script {
            env.VERSION = readFile('./release/VERSION').trim()
            env.CURRENT_VERSION = "${env.VERSION}-rc-${env.RC_NUMBER}"
            env.CONFIG_CLUSTER_NAME = "jenkins-${env.CURRENT_VERSION}-test"
          }

          withCredentials([string(credentialsId: 'nimba-wavefront-token', variable: 'WAVEFRONT_TOKEN')]) {
            withEnv(["PATH+GCLOUD=${HOME}/google-cloud-sdk/bin"]) {
              sh './hack/jenkins/setup-for-integration-test.sh'
              sh 'make gke-connect-to-cluster'
              sh './hack/test/deploy/deploy-local-linux.sh'
              sh './hack/test/test-wavefront-metrics.sh -c ${WF_CLUSTER} -t ${WAVEFRONT_TOKEN} -n ${CONFIG_CLUSTER_NAME} -v ${VERSION}'
            }
          }
        }
      }
      stage("Publish GA Harbor Image") {
        when{ environment name: 'RELEASE_TYPE', value: 'release' }
        environment {
          HARBOR_CREDS = credentials("projects-registry-vmware-tanzu_observability-robot")
          RELEASE_TYPE = 'release'
          PREFIX = 'projects.registry.vmware.com/tanzu_observability'
          DOCKER_IMAGE = 'kubernetes-collector'
        }
        steps {
          sh 'echo $HARBOR_CREDS_PSW | docker login $PREFIX -u $HARBOR_CREDS_USR --password-stdin'
          sh 'HARBOR_CREDS_USR=$(echo $HARBOR_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'
        }
      }
      stage("Publish GA Docker Hub") {
        when{ environment name: 'RELEASE_TYPE', value: 'release' }
        environment {
          DOCKERHUB_CREDS=credentials('Dockerhub_svcwfjenkins')
          RELEASE_TYPE = 'release'
          PREFIX = 'wavefronthq'
          DOCKER_IMAGE = 'wavefront-kubernetes-collector'
        }
        steps {
          sh 'echo $DOCKERHUB_CREDS_PSW | docker login -u $DOCKERHUB_CREDS_USR --password-stdin'
          sh 'make publish'
        }
      }
      stage("Create and Merge Bump Version Pull Request") {
        steps {
          sh './hack/jenkins/create-and-merge-pull-request.sh'
        }
      }

      stage("Github Release And Slack Notification") {
        environment {
          GITHUB_CREDS_PSW = credentials("GITHUB_TOKEN")
          CHANNEL_ID = credentials("k8s-assist-slack-ID")
          SLACK_WEBHOOK_URL = credentials("slack_hook_URL")
          BUILD_USER_ID = getBuildUserID()
          BUILD_USER = getBuildUser()
        }
        when{ environment name: 'RELEASE_TYPE', value: 'release' }
        steps {
          sh './hack/jenkins/generate_github_release.sh'
          sh './hack/jenkins/generate_slack_notification.sh'
        }
      }
    }
    post {
        always {
            cleanWs()
        }
    }
}

def getCurrentBranchName() {
      return env.BRANCH_NAME.split("/")[1]
}

def getBuildUser() {
      return "${currentBuild.getBuildCauses()[0].userName}"
}

def getBuildUserID() {
      return "${currentBuild.getBuildCauses()[0].userId}"
}
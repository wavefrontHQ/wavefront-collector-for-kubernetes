pipeline {
    agent any

    tools {
        go 'Go 1.15'
    }

    environment {
        RELEASE_TYPE = "${params.RELEASE_TYPE}"
        RC_NUMBER = "${params.RC_SUFFIX}"
        BUMP_COMPONENT = "${params.BUMP_COMPONENT}"
        GIT_BRANCH = getCurrentBranchName()
        GIT_CREDENTIAL_ID = 'wf-jenkins-github'
    }

    stages {
      stage("buildx") {
        steps {
          sh './hack/butler/install_docker_buildx.sh'
        }
      }
      stage("Bump with PR") {
         steps {
           withEnv(["PATH+EXTRA=${HOME}/go/bin"]){
             sh './hack/butler/create-next-version.sh "${BUMP_COMPONENT}"'
           }
           script {
             env.GIT_BUMP_BRANCH_NAME = readFile('./hack/butler/GIT_BUMP_BRANCH_NAME').trim()
             env.OLD_VERSION = readFile('./hack/butler/OLD_VERSION').trim()
             env.NEXT_VERSION = readFile('./hack/butler/NEXT_VERSION').trim()
           }
           withCredentials([string(credentialsId: 'GITHUB_TOKEN', variable: 'TOKEN')]) {
             sh 'git remote set-url origin https://${TOKEN}@github.com/wavefronthq/wavefront-collector-for-kubernetes.git'
             sh 'git config --global user.email "svc.wf-jenkins@vmware.com"'
             sh 'git config --global user.name "svc.wf-jenkins"'
             sh './hack/butler/bump-to-next-version.sh'
           }
         }
      }

      stage("Publish") {
        parallel {
          stage("Publish to Harbor") {
            environment {
              HARBOR_CREDS = credentials("projects-registry-vmware-tanzu_observability-robot")
            }
            steps {
              sh 'echo $HARBOR_CREDS_PSW | docker login "projects.registry.vmware.com/tanzu_observability" -u $HARBOR_CREDS_USR --password-stdin'
              sh 'PREFIX="projects.registry.vmware.com/tanzu_observability" HARBOR_CREDS_USR=$(echo $HARBOR_CREDS_USR | sed \'s/\\$/\\$\\$/\') DOCKER_IMAGE="kubernetes-collector" make publish'
            }
          }

          stage("Publish to Docker Hub") {
            environment {
              DOCKERHUB_CREDS=credentials('Dockerhub_svcwfjenkins')
            }
            when{ environment name: 'RELEASE_TYPE', value: 'release' }
            steps {
              sh 'echo $DOCKERHUB_CREDS_PSW | docker login -u $DOCKERHUB_CREDS_USR --password-stdin'
              sh 'PREFIX="wavefronthq" make publish'
            }
          }
        }
      }

      // deploy to GKE and EKS and run manual tests
      // now we have confidence in the validity of our RC release
      stage("Deploy and Test") {
        environment {
          GCP_CREDS = credentials("GCP_CREDS")
          GKE_CLUSTER_NAME = "k8po-jenkins-ci"
          WAVEFRONT_TOKEN = credentials("WAVEFRONT_TOKEN_NIMBA")
          WF_CLUSTER = 'nimba'
        }
        when{ environment name: 'RELEASE_TYPE', value: 'rc' }
        steps {
          script {
            env.VERSION = readFile('./release/VERSION').trim()
            env.CONFIG_CLUSTER_NAME = "jenkins-${env.VERSION}-rc-${env.RC_NUMBER}-test"
          }

          withCredentials([string(credentialsId: 'nimba-wavefront-token', variable: 'WAVEFRONT_TOKEN')]) {
            withEnv(["PATH+GCLOUD=/tmp/google-cloud-sdk/bin"]) {
              sh './hack/travis/setup-for-integration-test.sh'
              sh 'GKE_CLUSTER_NAME=k8po-jenkins-rc-testing make gke-connect-to-cluster'
              sh './release/deploy-local.sh'
              sh './hack/kustomize/test-e2e.sh -c ${WF_CLUSTER} -t ${WAVEFRONT_TOKEN} -n ${CONFIG_CLUSTER_NAME} -v ${VERSION}'
            }
          }
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
          sh './hack/butler/generate_github_release.sh'
          sh './hack/butler/generate_slack_notification.sh'
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
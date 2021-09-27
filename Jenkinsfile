pipeline {
    agent any

    tools {
        go 'Go 1.15'
    }

    environment {
        RELEASE_TYPE = "${params.RELEASE_TYPE}"
        RC_NUMBER = "${params.RC_SUFFIX}"
        GIT_BRANCH = getCurrentBranchName()
    }

    stages {
      stage("buildx") {
        steps {
          sh './hack/butler/install_docker_buildx.sh'
        }
      }
//       stage("Publish") {
//         parallel {
//           stage("Publish to Harbor") {
//             environment {
//               HARBOR_CREDS = credentials("projects-registry-vmware-tanzu_observability-robot")
//             }
//             steps {
//               sh 'echo $HARBOR_CREDS_PSW | docker login "projects.registry.vmware.com/tanzu_observability" -u $HARBOR_CREDS_USR --password-stdin'
//               sh 'PREFIX="projects.registry.vmware.com/tanzu_observability" HARBOR_CREDS_USR=$(echo $HARBOR_CREDS_USR | sed \'s/\\$/\\$\\$/\') DOCKER_IMAGE="kubernetes-collector" make publish'
//             }
//           }
//           stage("Publish to Docker Hub") {
//             environment {
//               DOCKERHUB_CREDS=credentials('Dockerhub_svcwfjenkins')
//             }
//             steps {
//               sh 'echo $DOCKERHUB_CREDS_PSW | docker login -u $DOCKERHUB_CREDS_USR --password-stdin'
//               sh 'PREFIX="wavefronthq" make publish'
//             }
//           }
//         }
//       }

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
//           sh './hack/butler/generate_github_release.sh'
          sh './hack/butler/generate_slack_notification.sh'
        }
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
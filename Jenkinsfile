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
//       stage("buildx") {
//         steps {
//           sh './hack/butler/install_docker_buildx.sh'
//         }
//       }
      stage("Bump with PR") {
//       check build status
// bump version by creating branch and PR (default to patch but have a dropdown on our build with parameters)
// use branch in below publish step
//         stage("check build status") {
//             sh 'curl github.com/...'
//         }
         steps {
           withEnv(["PATH+EXTRA=${HOME}/go/bin"]){
             sh './hack/butler/create-next-version.sh "${BUMP_COMPONENT}"'
             sh 'cat ./hack/butler/GIT_BUMP_BRANCH_NAME'
             sh 'cat ./hack/butler/OLD_VERSION'
             sh 'cat ./hack/butler/NEXT_VERSION'
           }
           script {
             env.GIT_BUMP_BRANCH_NAME = readFile('./hack/butler/GIT_BUMP_BRANCH_NAME').trim()
             env.OLD_VERSION = readFile('./hack/butler/OLD_VERSION').trim()
             env.NEXT_VERSION = readFile('./hack/butler/NEXT_VERSION').trim()
           }
           sh 'echo "${GIT_BUMP_BRANCH_NAME}"'
           withCredentials([string(credentialsId: 'GITHUB_TOKEN', variable: 'TOKEN')]) {
             sh 'git remote set-url origin https://${TOKEN}@github.com/wavefronthq/wavefront-collector-for-kubernetes.git'
             sh 'git config --global user.email "svc.wf-jenkins@vmware.com"'
             sh 'git config --global user.name "svc.wf-jenkins"'
//              sh 'git checkout -b ${GIT_BUMP_BRANCH_NAME}'
             sh './hack/butler/bump-to-next-version.sh'
           }
         }
      }

//         deploy to GKE and EKS and run manual tests
// now we have confidence in the validity of our RC release
      stage("Deploy and Test") {
        steps {
          sh 'GKE_CLUSTER_NAME=jenkins-testing-rc make create-gke-cluster'
          sh 'make e2e-test'
        }
      }

//         parallel {
//           stage("Publish to Harbor") {
//             environment {
//               HARBOR_CREDS = credentials("projects-registry-vmware-tanzu_observability-robot")
// GIT_BUMP_BRANCH_NAME = readFile(file: './release/GIT_BUMP_BRANCH_NAME')
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

// TODO: when / how do we want to trigger this?
//       stage("Github Release And Slack Notification") {
//         environment {
//           GITHUB_CREDS_PSW = credentials("GITHUB_TOKEN")
//           CHANNEL_ID = credentials("k8s-assist-slack-ID")
//           SLACK_WEBHOOK_URL = credentials("slack_hook_URL")
//           BUILD_USER_ID = getBuildUserID()
//           BUILD_USER = getBuildUser()
//         }
//         when{ environment name: 'RELEASE_TYPE', value: 'release' }
//         steps {
// //         approve and merge PR into master using gh API
//           sh './hack/butler/generate_github_release.sh'
//           sh './hack/butler/generate_slack_notification.sh'
//         }
//       }
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
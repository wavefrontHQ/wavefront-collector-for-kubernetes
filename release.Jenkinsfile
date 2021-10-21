pipeline {
    agent any

    tools {
        go 'Go 1.15'
    }

    environment {
        RELEASE_TYPE = "${params.RELEASE_TYPE}"
        RC_NUMBER = "1"
        BUMP_COMPONENT = "${params.BUMP_COMPONENT}"
        GIT_BRANCH = getCurrentBranchName()
        GIT_CREDENTIAL_ID = 'wf-jenkins-github'
    }

    stages {
      stage("Github Merge Bumped Version PR to Master") {
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
            sh './hack/butler/bump-version-and-raise-pull-request.sh'
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

def getCurrentBranchName() {
      return env.BRANCH_NAME.split("/")[1]
}

def getBuildUser() {
      return "${currentBuild.getBuildCauses()[0].userName}"
}

def getBuildUserID() {
      return "${currentBuild.getBuildCauses()[0].userId}"
}
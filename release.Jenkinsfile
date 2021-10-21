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
        TOKEN = credentials('GITHUB_TOKEN')
    }

    stages {
      stage("Create Bump Version Branch") {
        steps {
          withEnv(["PATH+EXTRA=${HOME}/go/bin"]){
            sh 'git config --global user.email "svc.wf-jenkins@vmware.com"'
            sh 'git config --global user.name "svc.wf-jenkins"'
            sh 'git remote set-url origin https://${TOKEN}@github.com/wavefronthq/wavefront-collector-for-kubernetes.git'

            sh './hack/butler/create-bump-version-branch.sh "${BUMP_COMPONENT}"'
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
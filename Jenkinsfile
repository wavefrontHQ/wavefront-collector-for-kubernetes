pipeline {
    agent any

    tools {
        go 'Go 1.15'
    }

    environment {
        PREFIX = "harbor-repo.vmware.com/tobs_keights_saas"
        DOCKER_CREDS = credentials("jenkins-wf-test")
        RELEASE_TYPE = "${params.RELEASE_TYPE}"
        RC_NUMBER = "${params.RC_SUFFIX}"
        GITHUB_CREDS = credentials("mamichael-test-github")
        GIT_BRANCH = getCurrentBranchName()
    }

    stages {
      stage("Release collector") {
        steps {
          sh './hack/butler/install_docker_buildx.sh'
          sh 'DOCKER_CREDS_USR=$(echo $DOCKER_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'
        }
      }

      stage("Generate Github Release") {
        when{ environment name: 'RELEASE_TYPE', value: 'release' }
        steps {
          sh './hack/butler/generate_github_release.sh'
        }
      }
    }
}

def getCurrentBranchName() {
      scmInfo = checkout scm
      return scmInfo.GIT_BRANCH
}

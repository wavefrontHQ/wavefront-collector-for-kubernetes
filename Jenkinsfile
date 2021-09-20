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
//         GITHUB_TOKEN = credentials("github-token-for-jenkins-priya")
    }

    stages {
      stage("Release collector") {
        steps {
          sh './hack/butler/install_docker_buildx.sh'
          sh 'DOCKER_CREDS_USR=$(echo $DOCKER_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'
//           sh 'echo Github token: ${GITHUB_TOKEN}'
          sh 'export GITHUB_TOKEN=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'
          sh 'make github-release'
        }
      }
    }
}


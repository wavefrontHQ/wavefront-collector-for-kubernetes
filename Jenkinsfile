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
        GITHUB_TOKEN=GITHUB_CREDS_PWS
    }

    stages {
      stage("Release collector") {
        steps {
          sh './hack/butler/install_docker_buildx.sh'
          sh 'DOCKER_CREDS_USR=$(echo $DOCKER_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'
          echo 'Github token: $(GITHUB_CREDS_USR)'
          sh 'GITHUB_TOKEN=$(GITHUB_CREDS_PWS) make github-release'
        }
      }
    }
}


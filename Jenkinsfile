pipeline {
    agent any

    tools {
        go 'Go 1.15'
    }

    environment {
        PREFIX = "harbor-repo.vmware.com/tobs_keights_saas"
        HARBOR_CREDS = credentials("jenkins-wf-test")
    }

    stages {
      stage("Release") {
        steps {
          sh 'HARBOR_CREDS_USR=$(echo $HARBOR_CREDS_USR | sed \'s/\\$/\\$\\$/\') echo \'${HARBOR_CREDS_PSW}\' | make harbor-docker-login'
//           sh 'VERSION=1.6.2 make container'
          echo "${params.RELEASE_TYPE}"
          echo "${HARBOR_CREDS_USR}"
        }
      }
    }
}


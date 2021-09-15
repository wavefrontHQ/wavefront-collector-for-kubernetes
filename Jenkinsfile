pipeline {
    agent any

    tools {
        go 'Go 1.15'
    }

    environment {
        PREFIX = "harbor-repo.vmware.com/tobs_keights_saas"
        DOCKER_CREDS = credentials("jenkins-wf-test")
        RELEASE_TYPE = params.RELEASE_TYPE
    }

    stages {
      stage("PUBLISH") {
//      when {params.PUBLISH == true}
        steps {
          echo "${params.RELEASE_TYPE}"
          echo "${DOCKER_CREDS_USR}"
          echo "**************publish******************************"
          sh 'DOCKER_CREDS_USR=$(echo $DOCKER_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'  //harbor
//           sh 'DOCKER_CREDS_USR=$(echo $DOCKER_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'  dockerhub
//        github release
        }
      }
    }
}


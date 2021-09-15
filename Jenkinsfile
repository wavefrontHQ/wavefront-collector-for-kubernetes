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
          sh '''
          # 1. get buildx binary from hosted location. We're getting amd64 because our EC2s are ubuntu on amd64
          wget -O docker-buildx https://github.com/docker/buildx/releases/download/v0.5.1/buildx-v0.5.1.linux-amd64
          # 2. make the binary executable
          chmod a+x docker-buildx
          # 3. create a dir (if it does not exist) to keep the binary
          [[ ! -d "~/.docker/cli-plugins" ]] && sudo mkdir -p ~/.docker/cli-plugins
          # 4. move the binary to the dir
          sudo mv docker-buildx ~/.docker/cli-plugins
          # 5. final step - run docker buildx --help
          sudo docker buildx --help
          '''
          sh 'DOCKER_CREDS_USR=$(echo $DOCKER_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'  //harbor
//           sh 'DOCKER_CREDS_USR=$(echo $DOCKER_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'  dockerhub
//        github release
        }
      }
    }
}


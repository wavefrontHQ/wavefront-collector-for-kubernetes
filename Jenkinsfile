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
      stage("Publish") {
        sh './hack/butler/install_docker_buildx.sh'
        parallel {
          stage("Publish to Harbor") {
            environment {
              HARBOR_CREDS = credentials("jenkins-wf-test")
            }
            steps {
              sh 'echo $HARBOR_CREDS_PSW | docker login "harbor-repo.vmware.com/tobs_keights_saas" -u $HARBOR_CREDS_USR --password-stdin'
              sh 'PREFIX="harbor-repo.vmware.com/tobs_keights_saas" HARBOR_CREDS_USR=$(echo $HARBOR_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'
            }
          }
          stage("Publish to Docker Hub") {
            environment {
              DOCKERHUB_CREDS=credentials('dockerhub-credential-shaoh')
            }
            steps {
              sh 'echo $DOCKERHUB_CREDS_PSW | docker login -u $DOCKERHUB_CREDS_USR --password-stdin'
              sh 'PREFIX="helenshao" make publish' // change PREFIX to dockerhub registry
            }
          }
        }
      }

      stage("Github Release") {
        environment {
          GITHUB_CREDS = credentials("mamichael-test-github")
        }
        when{ environment name: 'RELEASE_TYPE', value: 'release' }
        steps {
          sh './hack/butler/generate_github_release.sh'
        }
      }
    }
}

def getCurrentBranchName() {
      return env.BRANCH_NAME.split("/")[1]
}

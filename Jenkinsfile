pipeline {
    agent any

    tools {
        go 'Go 1.15'
    }

    environment {
        HARBOR_CREDS = credentials("jenkins-wf-test")
        RELEASE_TYPE = "${params.RELEASE_TYPE}"
        RC_NUMBER = "${params.RC_SUFFIX}"
        GITHUB_CREDS = credentials("mamichael-test-github")
        GIT_BRANCH = getCurrentBranchName()
        DOCKERHUB_CREDS=credentials('dockerhub-credential-shaoh')
    }

    stages {
      stage("Publish to Registries") {
        steps {
          sh './hack/butler/install_docker_buildx.sh'
          sh 'PREFIX="harbor-repo.vmware.com/tobs_keights_saas" HARBOR_CREDS_USR=$(echo $HARBOR_CREDS_USR | sed \'s/\\$/\\$\\$/\') make release'

          sh 'echo $DOCKERHUB_CREDS_PSW | docker login -u $DOCKERHUB_CREDS_USR --password-stdin'
          sh 'PREFIX="helenshao" make release' // change PREFIX to dockerhub registry
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
      return scm.branches[0].name.replace("*/", "")
}

pipeline {
  agent {
    label 'nimbus-cloud'
  }

  stages {
    stage("Check for go.sum changed") {
        steps {
            def prevCommit = env.GIT_PREVIOUS_COMMIT ?: "HEAD~1"
            def currentCommit = env.GIT_COMMIT
            def statusCode = sh "./hack/diff_dependencies.sh -p ${prevCommit} -c ${currentCommit}", returnStatus: true
            if (statusCode == 1) {

            }
        }
    }
  }

  post {
    // Notify only on null->failure or success->failure or failure->success
    failure {
      script {
        if(currentBuild.previousBuild == null) {
          slackSend (channel: '#open-channel', message: "collector dependencies changed: please remember to create a ticket in selected for development according to the instructions in <https://confluence.eng.vmware.com/display/CNA/Release+Process|\"Collector Repo Licensing\" (under step 2)> (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
        }
      }
    }
    regression {
      slackSend (channel: '#open-channel', message: "collector dependencies changed: please remember to create a ticket in selected for development according to the instructions in <https://confluence.eng.vmware.com/display/CNA/Release+Process|\"Collector Repo Licensing\" (under step 2)> (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
    }
    fixed {
      slackSend (channel: '#open-channel', message: "collector osl dependencies synced (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
    }
  }
}
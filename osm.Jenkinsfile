pipeline {
  agent {
    label 'nimbus-cloud'
  }

  stages {
    stage("Check for go.sum changed") {
        tools {
            go 'Go 1.18'
        }
        steps {
            sh "./hack/diff_dependencies.sh"
        }
    }
  }

  post {
    failure {
      script {
        if(currentBuild.previousBuild == null) {
          slackSend (channel: '#open-channel', message: "Collector dependencies changed: please remember to create a ticket in selected for development according to the instructions in <https://confluence.eng.vmware.com/display/CNA/Release+Process|\"Collector Repo Licensing\" (under step 2)> (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
        }
      }
    }
    regression {
      slackSend (channel: '#open-channel', message: "Collector dependencies changed: please remember to create a ticket in selected for development according to the instructions in <https://confluence.eng.vmware.com/display/CNA/Release+Process|\"Collector Repo Licensing\" (under step 2)> (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
    }
    fixed {
      slackSend (channel: '#open-channel', message: "Collector osl dependencies synced (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
    }
  }
}
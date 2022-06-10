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
          slackSend (channel: '#open-channel', message: "Collector dependencies changed: remember to create a placeholder ticket to file a OSM ticket before release, <https://confluence.eng.vmware.com/display/CNA/Release+Process|see \"Collector Repo Licensing\" for more information> (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
        }
      }
    }
    regression {
      slackSend (channel: '#open-channel', message: "Collector dependencies changed: remember to create a placeholder ticket to file a OSM ticket before release, <https://confluence.eng.vmware.com/display/CNA/Release+Process|see \"Collector Repo Licensing\" for more information> (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
    }
    fixed {
      slackSend (channel: '#open-channel', message: "Collector OSL dependencies in-sync (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
    }
  }
}
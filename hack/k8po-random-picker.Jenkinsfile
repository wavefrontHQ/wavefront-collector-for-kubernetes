pipeline {
  agent any
  options {
    buildDiscarder(logRotator(numToKeepStr: '5'))
  }
  triggers {
    // Every weekday MST 8:30 AM converted to UTC
    cron('30 15 * * 1-5')
  }
  stages {
    stage ("Slack message rando-dev results") {
      steps {
        script {
          ORDER_PICKED = sh (script: './hack/rando-dev.sh', returnStdout: true).trim()
        }
        slackSend (channel: '#tobs-k8po-team', message:
        """
Today's random generator results from <${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]> is:
${ORDER_PICKED}
        """)
      }
    }
  }
}
pipeline {
  agent any
  options {
    buildDiscarder(logRotator(numToKeepStr: '5'))
  }
  triggers {
    // Every weekday MST 8:30 AM converted to UTC
    cron('30 23 * * 0-4')
  }
  stages {
    stage ("Slack message rando-dev results") {
      steps {
        script {
          ORDER_PICKED = sh (script: './hack/rando-dev.sh', returnStdout: true).trim()
        }
        slackSend (channel: '#tobs-k8po-team', message:
        """
Tomorrow's random generator results from <${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]> is:
${ORDER_PICKED}
        """)
      }
    }
  }
}
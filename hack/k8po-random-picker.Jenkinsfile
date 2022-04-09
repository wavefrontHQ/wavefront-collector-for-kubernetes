pipeline {
    agent any
    triggers {
        cron('12 0 * * *')
    }
    stages {
        stage ("SSH into dev env") {
            steps {
                script {
                  ORDER_PICKED = sh (
                    script: './hack/rando-dev.sh',
                    returnStdout: true
                  ).trim()
                }
                slackSend (channel: '#open-channel', message:
                """
Today's random generator results from <${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]> is:
${ORDER_PICKED}
                """)

            }
        }
    }
}
pipeline {
    agent any

    stages {
        stage ("SSH into dev env") {
            steps {
                script {
                  ORDER_PICKED = sh (
                    script: './hack/rando-dev.sh',
                    returnStdout: true
                  ).trim()
                }
                slackSend (channel: '#tobs-k8po-team', color: '#008000', message: """Today's random order run results from <${env.BUILD_URL}> is:
                ${ORDER_PICKED}
                """)

            }
        }
    }
}
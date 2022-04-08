pipeline {
    agent any

    stages {
        stage ("SSH into dev env") {
            steps {
                ORDER_PICKED = sh './hack/rando-dev.sh'
                slackSend (channel: '#open-channel', color: '#FF0000', message: "Today's random order run by <${env.BUILD_URL}> is ${ORDER_PICKED}")
            }
        }
    }
}
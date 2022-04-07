pipeline {
    agent any

    stages {
        stage ("SSH into dev env") {
            steps {
                sh """
                ./hack/rando-dev.sh
                """
            }
        }
    }
}
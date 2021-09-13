pipeline {
    agent any

    tools {
        go 'Go 1.15'
    }

    environment {
    }

    stages {
        if(env.BRANCH_NAME == 'add-jenkinsfile'){
            stage("Doing something") {
                steps {
                    sh 'make container'
                }
            }
        }
    }
}


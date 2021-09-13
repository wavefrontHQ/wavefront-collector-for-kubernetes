pipeline {
    agent any

    tools {
        go 'Go 1.15'
    }

    environment {
    }

    stages {
        if(env.BRANCH_NAME == 'add-jenkinsfile'){
            stage("Doing something and wanting to see Jenkins") {
                steps {
                    sh 'make container'
                }
            }
        }
    }
}


pipeline {
    agent any

    tools {
        go 'Go 1.15'
    }

    stages {
      stage("Doing something and wanting to see Jenkins") {
          steps {
            script {
              if(env.BRANCH_NAME == 'add-jenkinsfile'){
                  make container
              }
            }
          }
      }
    }
}


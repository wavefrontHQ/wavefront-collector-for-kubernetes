pipeline {
    agent any

    tools {
        go 'Go 1.15'
    }

    stages {
      stage("Release") {
          steps {
            script {
              if(env.BRANCH_NAME == 'move-to-butler'){
                  make container
              }
            }
          }
      }
    }
}


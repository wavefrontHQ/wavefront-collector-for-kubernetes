pipeline {
    agent any

    tools {
        go 'Go 1.15'
    }

    stages {
      stage("Release") {
          when { branch 'move-to-butler' }
          steps {
            make container
            }
          }
      }
    }
}


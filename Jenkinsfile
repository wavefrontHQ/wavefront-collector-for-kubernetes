pipeline {
    agent any

    tools {
        go 'Go 1.15'
    }

    stages {
      stage("Release") {
        when { branch 'move-to-butler' }
        steps {
          sh 'VERSION=1.6.2 make container'
          echo "${params.RELEASE_TYPE}"
        }
      }
    }
}


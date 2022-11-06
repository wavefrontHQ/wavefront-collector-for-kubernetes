pipeline {
  agent {
    label 'nimbus-cloud'
  }
  tools {
    go 'Go 1.18'
  }
  stages {
    stage("Setup") {
      steps {
        sh "./hack/setup-diff-dependencies.sh"
      }
    }
    stage("Run OSSPI scan") {
      parallel {
        stage("wavefront-collector-for-kubernetes") {
          steps {
            script {
              try {
                sh "./hack/diff-dependencies.sh -r wavefront-collector-for-kubernetes"
              } catch (err) {
                echo "Caught: ${err}"
                if (env.NEEDS_OSL == null) {
                  env.NEEDS_OSL = 'wavefront-collector-for-kubernetes'
                } else {
                  env.NEEDS_OSL = env.NEEDS_OSL + ', wavefront-collector-for-kubernetes'
                }
                echo "NEEDS_OSL: ${env.NEEDS_OSL}"
              }
            }
          }
        }
        stage('wavefront-operator-for-kubernetes') {
          steps {
            sh 'rm wavefront-operator-for-kubernetes -rf; mkdir wavefront-operator-for-kubernetes'
            dir ('wavefront-operator-for-kubernetes') {
              git branch: 'main',
              credentialsId: 'wf-jenkins-github',
              url: 'https://github.com/wavefrontHQ/wavefront-operator-for-kubernetes.git'
              script {
                try {
                  sh "./../hack/diff-dependencies.sh -r wavefront-operator-for-kubernetes"
                } catch (err) {
                  echo "Caught: ${err}"
                  if (env.NEEDS_OSL == null) {
                    env.NEEDS_OSL = 'wavefront-operator-for-kubernetes'
                  } else {
                    env.NEEDS_OSL = env.NEEDS_OSL + ', wavefront-operator-for-kubernetes'
                  }
                  echo "NEEDS_OSL: ${env.NEEDS_OSL}"
                }
              }
            }
          }
        }
        stage('wavefront-kubernetes-adapter') {
          steps {
            sh 'rm wavefront-kubernetes-adapter -rf; mkdir wavefront-kubernetes-adapter'
            dir ('wavefront-kubernetes-adapter') {
              git branch: 'master',
              credentialsId: 'wf-jenkins-github',
              url: 'https://github.com/wavefrontHQ/wavefront-kubernetes-adapter.git'
              script {
                try {
                  sh "./../hack/diff-dependencies.sh -r wavefront-kubernetes-adapter"
                } catch (err) {
                  echo "Caught: ${err}"
                  if (env.NEEDS_OSL == null) {
                    env.NEEDS_OSL = 'wavefront-kubernetes-adapter'
                  } else {
                    env.NEEDS_OSL = env.NEEDS_OSL + ', wavefront-kubernetes-adapter'
                  }
                  echo "NEEDS_OSL: ${env.NEEDS_OSL}"
                }
              }
            }
          }
        }
        stage('prometheus-storage-adapter') {
          steps {
            sh 'rm prometheus-storage-adapter -rf; mkdir prometheus-storage-adapter'
            dir ('prometheus-storage-adapter') {
              git branch: 'master',
              credentialsId: 'wf-jenkins-github',
              url: 'https://github.com/wavefrontHQ/prometheus-storage-adapter.git'
              script {
                try {
                  sh "./../hack/diff-dependencies.sh -r prometheus-storage-adapter"
                } catch (err) {
                  echo "Caught: ${err}"
                  if (env.NEEDS_OSL == null) {
                    env.NEEDS_OSL = 'prometheus-storage-adapter'
                  } else {
                    env.NEEDS_OSL = env.NEEDS_OSL + ', prometheus-storage-adapter'
                  }
                  echo "NEEDS_OSL: ${env.NEEDS_OSL}"
                }
              }
            }
          }
        }
      }
    }
  }

  post {
    always {
      script {
        if(needToSendDepStatus()) {
           echo "needToSendDepStatus is true"
           slackSend (channel: '#open-channel', message: "Dependency change identified for these repositories: ${env.NEEDS_OSL} (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
        }
      }
    }
    failure {
      script {
        if(currentBuild.previousBuild == null) {
          slackSend (channel: '#open-channel', message: "Build failed (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
        }
      }
    }
    regression {
      slackSend (channel: '#open-channel', message: "Build regressed (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
    }
    fixed {
      slackSend (channel: '#open-channel', message: "Build fixed (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
    }

  }
}

// Send dependency status when either a user triggered the job or if dependency status changed from previous build
def needToSendDepStatus() {
    if (currentBuild.getBuildCauses('hudson.model.Cause$UserIdCause') != null) {
        return true
    }
    def prevBuildRepoStatus = currentBuild.previousBuild.buildVariables["NEEDS_OSL"]
    def prevBuildResult = prevBuildRepoStatus.replaceAll("\\s","").split(',') as List
    def currentBuildResult = env.NEEDS_OSL.replaceAll("\\s","").split(',') as List
    return prevBuildResult.sort() != currentBuildResult.sort()
}
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
        def isStartedByUser = currentBuild.getBuildCauses('hudson.model.Cause$UserIdCause') != null
        if (hasAnyReposDepStatusChanged || isStartedByUser) {
           slackSend (channel: '#closed-channel', message: "Need OSL ${env.NEEDS_OSL} (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
        }
      }
    }
    failure {
      script {
        if(currentBuild.previousBuild == null) {
          slackSend (channel: '#closed-channel', message: "build failed (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
        }
      }
    }
    regression {
      slackSend (channel: '#closed-channel', message: "build regressed (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
    }
    fixed {
      slackSend (channel: '#closed-channel', message: "All of our repositories' open source licenses are latest and updated! (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
    }

  }
}

def hasAnyReposDepStatusChanged() {
    prevBuildRepoStatus = ${currentBuild.previousBuild.buildVariables["NEEDS_OSL"]}
    prevBuildResult = Arrays.asList(prevBuildRepoStatus.replaceAll("\\s","").split(","))
    currentBuildResult = Arrays.asList(${env.NEEDS_OSL}.replaceAll("\\s","").split(","))
    return prevBuildResult.equals(currentBuildResult)
}
pipeline {
  agent {
    label 'nimbus-cloud'
  }
  options {
    buildDiscarder(logRotator(numToKeepStr: '10'))
  }
  triggers {
    // Every weekday MST 9:00 PM converted to UTC
    cron('0 4 * * 1-5')
//     cron('*/7 * * * *')
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
            sh 'rm wavefront-collector-for-kubernetes -rf; mkdir wavefront-collector-for-kubernetes'
            dir ('wavefront-collector-for-kubernetes') {
              git branch: 'main',
              credentialsId: 'wf-jenkins-github',
              url: 'https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes.git'
              script {
                try {
                  sh "./../hack/diff-dependencies.sh -r wavefront-collector-for-kubernetes"
                } catch (err) {
                  if (!err.getMessage().contains("exit code 8")) {
                    error('Caught unexpected error code')
                  }
                  if (env.NEEDS_OSL == null) {
                    env.NEEDS_OSL = 'wavefront-collector-for-kubernetes'
                  } else {
                    env.NEEDS_OSL = env.NEEDS_OSL + ', wavefront-collector-for-kubernetes'
                  }
                }
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
                  if (!err.getMessage().contains("exit code 8")) {
                    error('Caught unexpected error code')
                  }
                  if (env.NEEDS_OSL == null) {
                    env.NEEDS_OSL = 'wavefront-operator-for-kubernetes'
                  } else {
                    env.NEEDS_OSL = env.NEEDS_OSL + ', wavefront-operator-for-kubernetes'
                  }
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
                  if (!err.getMessage().contains("exit code 8")) {
                    error('Caught unexpected error code')
                  }
                  if (env.NEEDS_OSL == null) {
                    env.NEEDS_OSL = 'wavefront-kubernetes-adapter'
                  } else {
                    env.NEEDS_OSL = env.NEEDS_OSL + ', wavefront-kubernetes-adapter'
                  }
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
                  if (!err.getMessage().contains("exit code 8")) {
                    error('Caught unexpected error code')
                  }
                  if (env.NEEDS_OSL == null) {
                    env.NEEDS_OSL = 'prometheus-storage-adapter'
                  } else {
                    env.NEEDS_OSL = env.NEEDS_OSL + ', prometheus-storage-adapter'
                  }
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
           slackSend (channel: '#open-channel', message: "These repositories need a new open source license: ${env.NEEDS_OSL} (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
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
    echo "${currentBuild.buildCauses}" // same as currentBuild.getBuildCauses()
//     echo "${currentBuild.getBuildCauses('hudson.model.Cause$UserIdCause')}"
//     echo "${currentBuild.getBuildCauses('hudson.triggers.TimerTrigger$TimerTriggerCause')}"
    if (currentBuild.getBuildCauses('hudson.triggers.TimerTrigger$TimerTriggerCause').isEmpty()){
      echo 'Need to send status because timer did not trigger the job.'
      return true
    }
    if (currentBuild.getBuildCauses('hudson.model.Cause$UserIdCause') != null) {
      echo 'Need to send status because user triggered the job.'
      return true
    }
    def prevBuildRepoStatus = currentBuild.previousBuild.buildVariables["NEEDS_OSL"]
    def prevBuildResult = prevBuildRepoStatus.replaceAll("\\s","").split(',') as List
    def currentBuildResult = env.NEEDS_OSL.replaceAll("\\s","").split(',') as List
    if(prevBuildResult.sort() != currentBuildResult.sort()) {
      echo 'Need to send status because build result changed from last job.'
      return true
    }
    return false
}
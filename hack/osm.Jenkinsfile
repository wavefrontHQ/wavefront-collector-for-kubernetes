pipeline {
  agent {
    label 'nimbus-cloud'
  }
  tools {
    go 'Go 1.18'
  }
  stages {
    parallel {
      stage("wavefront-collector-for-kubernetes") {
        steps {
          script {
            try {
              sh "./hack/diff_dependencies.sh -r wavefront-collector-for-kubernetes"
            } catch (err) {
              echo "Caught: ${err}"
              if (env.NEEDS_OSL == "") {
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
                sh "./../hack/diff_dependencies.sh -r wavefront-operator-for-kubernetes"
              } catch (err) {
                echo "Caught: ${err}"
                if (env.NEEDS_OSL == "") {
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
            git branch: 'main',
            credentialsId: 'wf-jenkins-github',
            url: 'https://github.com/wavefrontHQ/wavefront-kubernetes-adapter.git'
            script {
              try {
                sh "./../hack/diff_dependencies.sh -r wavefront-kubernetes-adapter"
              } catch (err) {
                echo "Caught: ${err}"
                if (env.NEEDS_OSL == "") {
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
            git branch: 'main',
            credentialsId: 'wf-jenkins-github',
            url: 'https://github.com/wavefrontHQ/prometheus-storage-adapter.git'
            script {
              try {
                sh "./../hack/diff_dependencies.sh -r prometheus-storage-adapter"
              } catch (err) {
                echo "Caught: ${err}"
                if (env.NEEDS_OSL == "") {
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

//   post {
//     failure {
//       script {
//         if(currentBuild.previousBuild == null) {
//           slackSend (channel: '#tobs-k8po-team', message: "@k8po-eng-team Collector dependencies changed: remember to create a JIRA ticket for \"OSM Release\" in \"Selected For Development\" before next collector release, <https://confluence.eng.vmware.com/display/CNA/Release+Process|see \"Collector Repo Licensing\" for more information> (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
//         }
//       }
//     }
//     regression {
//       slackSend (channel: '#tobs-k8po-team', message: "@k8po-eng-team Collector dependencies changed: remember to create a JIRA ticket for \"OSM Release\" in \"Selected For Development\" before next collector release, <https://confluence.eng.vmware.com/display/CNA/Release+Process|see \"Collector Repo Licensing\" for more information> (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
//     }
//     fixed {
//       slackSend (channel: '#tobs-k8po-team', message: "@k8po-eng-team Collector OSL dependencies in-sync (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
//     }
//   }
}
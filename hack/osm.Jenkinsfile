pipeline {
  agent {
    label 'nimbus-cloud'
  }
  tools {
    go 'Go 1.18'
  }
  stages {
    stage('Clone another repository') {
        steps {
            sh 'rm operator -rf; mkdir operator'
            dir ('operator') {
                git branch: 'main',
                credentialsId: 'wf-jenkins-github',
                url: 'https://github.com/wavefrontHQ/wavefront-operator-for-kubernetes.git'
                sh 'pwd'
                sh "./../hack/diff_dependencies.sh"
            }
        }
    }
    stage("Check for go.sum changed") {
        steps {
            sh "./hack/diff_dependencies.sh"
        }
    }
  }

  post {
    failure {
      script {
        if(currentBuild.previousBuild == null) {
          slackSend (channel: '#tobs-k8po-team', message: "@k8po-eng-team Collector dependencies changed: remember to create a JIRA ticket for \"OSM Release\" in \"Selected For Development\" before next collector release, <https://confluence.eng.vmware.com/display/CNA/Release+Process|see \"Collector Repo Licensing\" for more information> (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
        }
      }
    }
    regression {
      slackSend (channel: '#tobs-k8po-team', message: "@k8po-eng-team Collector dependencies changed: remember to create a JIRA ticket for \"OSM Release\" in \"Selected For Development\" before next collector release, <https://confluence.eng.vmware.com/display/CNA/Release+Process|see \"Collector Repo Licensing\" for more information> (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
    }
    fixed {
      slackSend (channel: '#tobs-k8po-team', message: "@k8po-eng-team Collector OSL dependencies in-sync (<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>)")
    }
  }
}
pipeline {
    agent any

    stages {
//         stage("Test with Go 1.15") {
//             tools {
//                 go 'Go 1.15'
//             }
//             steps {
//                 withEnv(["PATH+EXTRA=${HOME}/go/bin"]) {
//                   sh 'make checkfmt vet tests'
//                 }
//             }
//         }
//         stage("Test with Go 1.16") {
//             tools {
//                 go 'Go 1.16'
//             }
//             steps {
//                 withEnv(["PATH+EXTRA=${HOME}/go/bin"]) {
//                   sh 'make checkfmt vet tests'
//                 }
//             }
//         }
//
//         stage("Publish") {
//             tools {
//                 go 'Go 1.15'
//             }
//             environment {
//                 RELEASE_TYPE = "alpha"
//                 VERSION_POSTFIX = "-alpha-${GIT_COMMIT.substring(0, 8)}"
// //                 VERSION_POSTFIX = "-alpha-e0fe165d"
//
//                 HARBOR_CREDS = credentials("projects-registry-vmware-tanzu_observability_keights_saas-robot")
//
//                 PREFIX = "projects.registry.vmware.com/tanzu_observability_keights_saas"
//                 DOCKER_IMAGE = "kubernetes-collector-snapshot"
//             }
//             steps {
//                 withEnv(["PATH+EXTRA=${HOME}/go/bin"]) {
//                     sh './hack/butler/install_docker_buildx.sh'
//
//                     sh 'make semver-cli'
//                     sh 'echo $HARBOR_CREDS_PSW | docker login $PREFIX -u $HARBOR_CREDS_USR --password-stdin'
//                     sh 'HARBOR_CREDS_USR=$(echo $HARBOR_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'
//                 }
//             }
//         }
//
//         stage("Integration Test") {
//             tools {
//                 go 'Go 1.15'
//             }
//             environment {
//                 GCP_CREDS = credentials("GCP_CREDS")
//                 GKE_CLUSTER_NAME = "k8po-jenkins-ci"
//                 WAVEFRONT_TOKEN = credentials("WAVEFRONT_TOKEN_NIMBA")
//             }
//             steps {
//                 withEnv(["PATH+GO=${HOME}/go/bin", "PATH+GCLOUD=${HOME}/google-cloud-sdk/bin"]) {
//                     lock("collector-integration-test") {
//                         sh './hack/jenkins/setup-for-integration-test.sh'
//                         sh 'make gke-connect-to-cluster'
//                         sh 'make integration-test'
//                     }
//                 }
//             }
//         }
        stage("Test remove later") {
            environment {
              CHANNEL_ID = 'G01AZ1WP8UE'
              SLACK_WEBHOOK_URL = credentials("slack_hook_URL")
            }

            steps {
                sh 'curl -X POST --data-urlencode "payload={\"channel\": \"${CHANNEL_ID}\", \"username\": \"jenkins\", \"text\": \"Success!! released by ${BUILD_USER}(${BUILD_USER_ID})!\"}" ${SLACK_WEBHOOK_URL}'
//                error 'fail'
//                 sh 'echo success'
            }
        }
    }
    post {
        regression {
        }
    }
}
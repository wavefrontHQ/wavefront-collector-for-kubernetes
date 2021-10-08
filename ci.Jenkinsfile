pipeline {
    agent any

    environment {
        RELEASE_TYPE = "alpha"
        VERSION_POSTFIX = "-alpha-${GIT_COMMIT.substring(0, 8)}"

        PREFIX = "projects.registry.vmware.com/tanzu_observability_keights_saas"
        DOCKER_IMAGE = "kubernetes-collector-snapshot"

        HARBOR_CREDS = credentials("projects-registry-vmware-tanzu_observability_keights_saas-robot")
    }

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

        stage("Publish") {
            tools {
                go 'Go 1.15'
            }
            steps {
                withEnv(["PATH+EXTRA=${HOME}/go/bin"]) {
                    sh './hack/butler/install_docker_buildx.sh'

                    sh 'make semver-cli'
                    sh 'echo $HARBOR_CREDS_PSW | docker login $PREFIX -u $HARBOR_CREDS_USR --password-stdin'
                    sh 'HARBOR_CREDS_USR=$(echo $HARBOR_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'
                }
            }
        }

        stage("Integration Test") {
            tools {
                go 'Go 1.15'
            }
            environment {
                GCP_CREDS = credentials("GCP_CREDS")
                GKE_CLUSTER_NAME = "k8s-saas-travis-ci"
            }
            steps {
                withEnv(["PATH+EXTRA=${HOME}/go/bin"]) {
                    lock("collector-integration-test") {
                        sh './hack/travis/setup-for-integration-test.sh'
                        sh 'make gke-connect-to-cluster'
                        sh 'make ci-integration-test'
                    }
                }
            }
        }
    }
}
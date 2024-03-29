pipeline {
    agent {
        label 'amazon-linux2'
    }

    options {
        timeout(time: 15, unit: 'MINUTES')
        timestamps()
    }

    environment {
        GITHUB_API_USER = 'draios-jenkins@sysdig.com'
        GITHUB_API_KEY = credentials('jenkins-github-token')
        CURRENT_VERSION = "v0.0.7"
    }

    stages {
        stage('Checkout') {
            //clean
            steps {
                script {
                    sh "ls ${env.WORKSPACE}"
                    sh "docker rmi sysdiglabs/stackdriver-webhook-bridge:latest || true"
                }

                //checkout
                git branch: "${BRANCH}", changelog: false, credentialsId:'github-jenkins-user-token', poll: false, url: 'https://github.com/sysdiglabs/stackdriver-webhook-bridge.git'
            }
        }
        stage('Build') {
            steps {
                script {
                    sh "make image"
                }
            }
        }
        stage('Publish Docker image') {
            environment {
                GIT_HASH = GIT_COMMIT.take(7)
            }
            steps {
                script {
                    env.VERSION_BUILD_NUMBER=env.CURRENT_VERSION+"-"+env.GIT_HASH
                    echo "tag ${env.VERSION_BUILD_NUMBER}"
                }
                withCredentials([usernamePassword(credentialsId: "dockerhub-robot-account", passwordVariable: "DOCKER_PASSWORD", usernameVariable: "DOCKER_USERNAME")]) {
                    sh "docker login -u=${env.DOCKER_USERNAME} -p=${env.DOCKER_PASSWORD}"
                    sh "docker tag sysdiglabs/stackdriver-webhook-bridge:latest sysdiglabs/stackdriver-webhook-bridge:${env.VERSION_BUILD_NUMBER}"
                    sh "docker push sysdiglabs/stackdriver-webhook-bridge:${env.VERSION_BUILD_NUMBER}"
                }
            }
        }
        stage('Cleanup') {
            //clean
            steps {
                script {
                    sh "docker rmi sysdiglabs/stackdriver-webhook-bridge:latest"
                    sh "docker rmi sysdiglabs/stackdriver-webhook-bridge:${env.VERSION_BUILD_NUMBER}"
                }
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}

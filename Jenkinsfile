pipeline {
  agent {
    label 'builder-backend-j8'
  }

  environment {
    GITHUB_API_USER = 'draios-jenkins@sysdig.com'
    GITHUB_API_KEY = credentials('jenkins-github-token')
  }

  stages {
    stage('Checkout') {
      //clean
      steps {
        script {
          sh "ls ${env.WORKSPACE}"
          sh "docker rm sysdiglabs/stackdriver-webhook-bridge || echo \\\"Builder image not found\\\""
        }

        //checkout
        git branch: "${BRANCH_NAME}", changelog: false, credentialsId:'github-jenkins-user-token', poll: false, url: 'https://github.com/sysdiglabs/stackdriver-webhook-bridge.git'
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
                env.VERSION_BUILD_NUMBER="v0.0.6-"+env.GIT_HASH
                echo "tag ${env.VERSION_BUILD_NUMBER}"
            }
            withCredentials([usernamePassword(credentialsId: "docker-hub", passwordVariable: "DOCKER_PASSWORD", usernameVariable: "DOCKER_USERNAME")]) {
                sh "docker login -u=${env.DOCKER_USERNAME} -p=${env.DOCKER_PASSWORD}"
            }
        }
    }
    stage('Cleanup') {
      //clean
      steps {
        script {
          sh "docker rm sysdiglabs/stackdriver-webhook-bridge || echo \\\"Builder image not found\\\""
        }
      }
    }
  }
  post {
        always {
      echo 'One way or another, I have finished'
        }
        success {
      echo 'I succeeeded!'
        }
        unstable {
      echo 'I am unstable :/'
        }
        failure {
      echo 'I failed :('
        }
        changed {
      echo 'Things were different before...'
        }
  }
}

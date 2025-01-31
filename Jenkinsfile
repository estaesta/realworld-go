pipeline {
    agent any

    environment {
        DOCKER_IMAGE_NAME = 'realworld-go'
        DOCKER_CONTAINER_NAME = 'realworld-go'
        PORT = 8082
        DB_PATH = credentials('DB_PATH')
        JWT_SECRET = credentials('JWT_SECRET')
    }
    stages {
        stage('Scan Filesystem') {
            steps {
                sh "docker run --rm -v /tmp/trivy:/root/.cache/trivy -v ${env.WORKSPACE}:/myapp trivy fs --format table -o /myapp/trivy-fs-report.html /myapp"
            }
        }

        stage('Build') {
            steps {
                sh '''
                    docker build -t $DOCKER_IMAGE_NAME .
                    docker tag $DOCKER_IMAGE_NAME $DOCKER_IMAGE_NAME:latest
                    docker tag $DOCKER_IMAGE_NAME $DOCKER_IMAGE_NAME:$GIT_COMMIT
                    '''
            }
        }

        stage('Scan Image') {
            steps {
                sh "docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v /tmp/trivy:/root/.cache/trivy -v ${env.WORKSPACE}:/myapp trivy image --format table -o /myapp/trivy-image-report.html $DOCKER_IMAGE_NAME"
            }
        }

        stage('Deploy') {
            steps {
                sh '''
                    docker compose down || true
                    docker compose up -d
                '''
            }
        }
        stage('Test') {
            steps {
                sh '''
                    sleep 10
                    curl -s http://localhost:$PORT/health | grep -q '"status":"UP"'
                '''
            }
        }
    }
    post {
        always {
            archiveArtifacts artifacts: 'trivy-fs-report.html', fingerprint: true
            archiveArtifacts artifacts: 'trivy-image-report.html', fingerprint: true
        }
    }
}

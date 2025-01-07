pipeline {
    agent any

    environment {
        DOCKER_IMAGE_NAME = 'realworld-go:latest'
        DOCKER_CONTAINER_NAME = 'realworld-go'
        PORT = 8080
    }
    stages {
        stage('Scan') {
            steps {
                sh "docker run -v $PWD:/myapp aquasec/trivy fs --format table -o trivy-fs-report.html /myapp"
            }
        }

        stage('Build') {
            steps {
                sh 'docker build -t $DOCKER_IMAGE_NAME .'
            }
        }

        stage('Scan Image') {
            steps {
                sh "docker run -v $PWD:/myapp aquasec/trivy image --format table -o trivy-image-report.html $DOCKER_IMAGE_NAME"
            }
        }

        stage('Deploy') {
            steps {
                sh '''
                    docker stop $DOCKER_CONTAINER_NAME || true
                    docker rm $DOCKER_CONTAINER_NAME || true
                    docker run -d -p $PORT:8080 --name $DOCKER_CONTAINER_NAME $DOCKER_IMAGE_NAME
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
}

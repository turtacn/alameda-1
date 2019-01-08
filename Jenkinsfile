node('go11') {
  stage('checkout') {
        git branch: 'master', url: "https://github.com/containers-ai/alameda.git"
  }
  stage("Build Operator") {
    sh """
      export GOROOT=/usr/local/go
      export GOPATH=/go/src/workspace
      mkdir -p /go/src/workspace/src/github.com/containers-ai
      mv ${env.WORKSPACE} /go/src/workspace/src/github.com/containers-ai/alameda
      cd /go/src/workspace/src/github.com/containers-ai/alameda/operator
      make manager
    """
  }
  stage("Build Datahub") {
    sh """
      export GOROOT=/usr/local/go
      export GOPATH=/go/src/workspace
      cd /go/src/workspace/src/github.com/containers-ai/alameda/datahub
      pwd
      make datahub
    """
  }
  stage("Test Operator") {
    sh """
      export GOROOT=/usr/local/go
      export GOPATH=/go/src/workspace
      cd /go/src/workspace/src/github.com/containers-ai/alameda/operator
      make test
    """
  }
  stage("Test Datahub") {
    sh """
      export GOROOT=/usr/local/go
      export GOPATH=/go/src/workspace
      cd /go/src/workspace/src/github.com/containers-ai/alameda/datahub
      curl -s https://codecov.io/bash | bash -s - -t ee1341b8-56f3-4319-8146-afc464130075
    """
  }
}

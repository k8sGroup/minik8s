BINARY_NAME=../bin
KUBE_CONTROLLER_MANAGER=../cmd/kube-controller-manager/controller-manager.go
API_SERVER=../cmd/kube-apiserver/apiserver.go

build controller:
	go build -o ${BINARY_NAME}/controller.exe ${KUBE_CONTROLLER_MANAGER}

build apiserver:
	go build -o ${BINARY_NAME}/apiserver.exe ${API_SERVER}




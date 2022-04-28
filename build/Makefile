BINARY_NAME=../bin
KUBE_CONTROLLER_MANAGER=../cmd/kube-controller-manager/controller-manager.go
API_SERVER=../cmd/kube-apiserver/apiserver.go

build:
	go build -o ${BINARY_NAME}/controller.exe ${KUBE_CONTROLLER_MANAGER}
	go build -o ${BINARY_NAME}/apiserver.exe ${API_SERVER}

run:
	${BINARY_NAME}/apiserver.exe
	${BINARY_NAME}/controller.exe





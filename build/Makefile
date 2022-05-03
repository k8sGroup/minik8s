BINARY_NAME=../bin
KUBE_CONTROLLER_MANAGER=../cmd/kube-controller-manager/controller-manager.go
API_SERVER=../cmd/kube-apiserver/apiserver.go
SCHEDULER=../cmd/kube-scheduler/scheduler.go
KUBELET=../cmd/kubelet/kubeletMain.go

build:
	go build -o ${BINARY_NAME}/controller.exe ${KUBE_CONTROLLER_MANAGER}
	go build -o ${BINARY_NAME}/apiserver.exe ${API_SERVER}
	go build -o ${BINARY_NAME}/scheduler.exe ${SCHEDULER}
#go build -o ${BINARY_NAME}/kubelet.exe ${KUBELET}

run:
	${BINARY_NAME}/apiserver.exe
	${BINARY_NAME}/scheduler.exe
	${BINARY_NAME}/controller.exe
	#${BINARY_NAME}/kubelet.exe





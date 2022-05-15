# Bug Report

## 复现
首先运行api-server和kube-controller-manager，然后进入etcd的docker， 需要开启两个窗口watch replicaset和deployment

```bash
etcdctl watch --prefix /registry/rs/default
etcdctl watch --prefix /registry/deployment/default
```

修改`deployment.yaml`文件为如下所示

```yaml
kind: Deployment
metadata:
  name: nginx-test
  labels:
    version: v1
    type: deployment
spec:
  replicas: 3
  strategy:
    type: rollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  template:
    metadata:
      name: nginx
    spec:
      nodeName: apollo
      containers:
      - name: nginx
        image: nginx:latest
```

命令行输入

``` bash
# build kubectl
go build -o odin ./cmd/kubectl/kubectl.go
# apply deployment.yaml
./odin apply ./test/deployment/deployment.yaml
```

修改`deployment.yaml`中的replicas为5

运行命令
``` bash
./odin apply ./test/deployment/deployment.yaml
```

观察到 watch replicaset 的窗口会出现三个 replicaset 。其中一个名称为 nginx-test 的 replicaset 是预期之外的 。 deployment 应该控制 replicaset 的名称，由 deployment 产生的 replicaset 名称为 deployment name + UUID

并且，在无操作一段时间之后，发现 watch replicaset 的窗口一直在更新名字为 nginx-test 的 replicaset 的信息， 显然出现了异常。
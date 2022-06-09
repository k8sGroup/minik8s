# flannel配置docker跨主机通信

## 配置

### Master Node: etcdv2

在负责网络配置的etcd v2中首先运行如下命令：

- network限制了容器可以获得的ip的范围
- type表明采用vxlan
- 这个key的前缀可以自行指定，但是最后一定要以config结尾
- 其他的字段可以自行查询

```bash
etcdctlv2 set /registry/network/test/config '{"Network": "172.16.0.0/16", "SubnetLen": 24, "SubnetMin": "172.16.1.0","SubnetMax": "172.16.32.0", "Backend": {"Type": "vxlan"}}'
```

### Work Node: flanneld

在Worker Node上边首先安装好flanneld可执行文件和生成docker环境变量的脚本。

flannel目录下有三个文件

```
README.md
flanneld 						# flanneld的可执行文件
mk-docker-opts.sh 	# 生成docker_opts.env的脚本
```

scp命令将目录拷贝到远程的/root目录下边

```bash
scp -r ./flannel root@xx.xx.xx.xx:/root
```

然后运行：

- etcd-endpoints：指定etcd v2的ip和端口
- iface-regex：正则表达式匹配interface，其实就是网卡
- etcd-prefix：和前边etcdctl命令里边的前缀对应上
- public-ip：填写自己的浮动ip

（这种方式为前台运行，部署时要后台运行）

```bash
./flanneld --etcd-endpoints="http://10.119.11.164:2379" \
--iface-regex="ens*|enp*" \
--ip-masq=true \
--etcd-prefix=/registry/network/test \
--public-ip=xxx.xxx.xxx.xxx
```

跑起来之后，flanneld会从etcd拿到信息来配置自己的网络，然后要执行mk-docker-opts.sh脚本以生成docker_opts.env文件，用作docker引擎环境变量的配置

```bash
./flannel/mk-docker-opts.sh -c
```

生成/run/docker_opts.env之后，更改docker.service的配置文件。

原始的文件行ExecStart：

```
ExecStart=/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock
```

在ExecStart前边加一行环境变量文件，然后更改ExecStart行，更改后的效果如下

```
EnvironmentFile=/run/docker_opts.env
ExecStart=/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock $DOCKER_OPTS
```

把service重新加载一遍，然后重启docker服务，记得在**重启之前要把/etc/docker/daemon.json文件删除！！！**

```bash
systemctl daemon-reload
systemctl restart docker
```

现在运行ifconfig可以发现docker0的ip已经变了

## 测试

下面进行测试

建议使用chn1234wanghaotian/render进行测试，里边安装了一些必备的网络工具，不需要再安装。

这个镜像运行/usr/bin/render之后会监听10080端口并且返回Hello World

在Host A中：

```bash
# 拉取镜像
docker pull chn1234wanghaotian/render
# 启动container，这里不是/bin/bash而是/usr/bin/render
docker run -itd --name render -p 10080:10080 chn1234wanghaotian/render /usr/bin/render
# 进入container A
docker exec -it render /bin/bash
# 在container A中查看ip
ifconfig
# 假设这个container A的ip是172.16.25.2
```

在Host B中：

```bash
# 拉取镜像
docker pull chn1234wanghaotian/render
# 启动container，这里不是/bin/bash而是/usr/bin/render
docker run -itd --name render -p 10080:10080 chn1234wanghaotian/render /usr/bin/render
# 进入container B
docker exec -it render /bin/bash
# 尝试访问Host A中container A
curl 172.16.10.2:10080
# 返回Hello World即为成功
```


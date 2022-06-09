# 关于环境配置

## 安装rabbitmq

```bash
apt-get install erlang-nox     									# 安装erlang
wget -O- https://www.rabbitmq.com/rabbitmq-release-signing-key.asc|apt-key add -
apt-get update
apt-get install rabbitmq-server									# 安装成功自动启动
rabbitmq-plugins enable rabbitmq_management   	# 启用插件
rabbitmqctl add_user root 123456								# 增加用户
rabbitmqctl set_user_tags root administrator		# 给普通用户分配管理员角色 
rabbitmqctl add_vhost admin
rabbitmqctl set_permissions -p admin root "." "." ".*"
systemctl restart rabbitmq-server.service				# 重启
```

## prometheus和etcd

将`control-plane`文件夹拷贝至master的/root目录下，应该可以看到/root/control-plane的目录

在`~/.bashrc`的末尾加入如下配置

```
# minik8s control-plane env
export CONTROL_PLANE=/root/control-plane
export ETCD_HOME=$CONTROL_PLANE/multi-etcd-instance

# etcd alias
alias etcdctlv2="ETCDCTL_API=2 $ETCD_HOME/etcdctl"
alias etcdctlv3="ETCDCTL_API=3 $ETCD_HOME/etcdctl --endpoints 127.0.0.1:12379"

# prometheus home 
export PROMETHEUS_HOME=$CONTROL_PLANE/prometheus
```

刷新环境变量

```bash
source ~/.bashrc
```

启动prometheuse和etcd，记得**修改你自己的prometheus.yml**

```bash
$ETCD_HOME/start-etcd.sh
$PROMETHEUS_HOME/start-prometheus.sh
```

如果想要让他们停止运行

```bash
$ETCD_HOME/stop-etcd.sh
$PROMETHEUS_HOME/stop-prometheus.sh
```

控制etcdv2和etcdv3使用如下命令，即使用etcdctlv2和etcdctlv3替代了etcdctl

```bash
etcdctlv3 put foo barv3
etcdctlv3 get foo

etcdctlv2 set foo barv2
etcdctlv2 get foo
```


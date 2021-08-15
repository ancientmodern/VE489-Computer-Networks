# 2021 SU VE489 Project

**Group 16**

**Name: 陆昊融，娄辰飞**

**Student ID: 518370910194, **

**Demo Video Link (Task 2): https://jbox.sjtu.edu.cn/l/611Hsm**

In this video I typed some wrong commands when executing the container for minecraft server, but the result is correct.



## Task 2 Procedure

As we have already tested the communication between two containers on the same host in Task 1, in Task 2 we will not implement a server and a client on the same host again. We will directly implement cross-host communication, IPTV, gRPC-web, and Minecraft.

For simplicity, we will call the first host machine "host1" and the second host machine "host2".

host1:

![image-20210716035515817](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716035515817.png)



host2:

![image-20210716035553378](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716035553378.png)



### 1. Cross-Host Etcd Configuration

The procedure is similar to Task1, only the `etcd.service` need to be changed.

```bash
# Copy from Task1
tar xzvf etcd-v3.3.10-linux-amd64.tar.gz
cd etcd-v3.3.10-linux-amd64
chmod +x {etcd,etcdctl}
cp {etcd,etcdctl} /usr/bin/

# Done both on host1 and host2
systemctl stop etcd
rm -rf /var/lib/etcd # Clear the previous cluster
vim /lib/systemd/system/etcd.service

##############################################################################
# /lib/systemd/system/etcd.service
[Unit]
Description=etcd
After=network.target

[Service]
Environment=ETCD_NAME=etcd-1 # 'etcd-2' on host2
Environment=ETCD_DATA_DIR=/var/lib/etcd
Environment=ETCD_LISTEN_CLIENT_URLS=http://192.168.122.128:2379,http://127.0.0.1:2379
Environment=ETCD_LISTEN_PEER_URLS=http://192.168.122.128:2380
Environment=ETCD_ADVERTISE_CLIENT_URLS=http://192.168.122.128:2379,http://127.0.0.1:2379
Environment=ETCD_INITIAL_ADVERTISE_PEER_URLS=http://192.168.122.128:2380
Environment=ETCD_INITIAL_CLUSTER_STATE=new
Environment=ETCD_INITIAL_CLUSTER_TOKEN=etcd-cluster-token
Environment=ETCD_INITIAL_CLUSTER=etcd-1=http://192.168.122.128:2380,etcd-2=http://192.168.122.129:2380
ExecStart=/usr/bin/etcd

[Install]
WantedBy=multi-user.target
##############################################################################

systemctl daemon-reload
systemctl restart etcd
```

Then by running `etcdctl member list`, we have,

![image-20210716035726109](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716035726109.png)



### 2. Cross-Host Flannel Configuration

The procedure is actually the same as Task1, except that we need to delete the previous flannel network,

```bash
# Copy from Task1
wget https://github.com/flannel-io/flannel/releases/download/v0.14.0/flanneld-amd64
chmod +x flanneld-amd64
cp flanneld-amd64 /usr/bin/flanneld

# Done both on host1 and host2
systemctl stop flannel
ifconfig flannel.1 down
ip link delete flannel.1
vim /lib/systemd/system/flannel.service

##############################################################################
# /lib/systemd/system/flannel.service
[Unit]
Description=flannel
After=etcd.service network.target

[Service]
ExecStart=/usr/bin/flanneld --etcd-endpoints=http://192.168.122.128:2379 -etcd-prefix=/docker-flannel/network --iface=ens33 # 192.168.122.129 on host2

[Install]
WantedBy=multi-user.target
##############################################################################

etcdctl set /docker-flannel/network/config {"Network":"10.3.0.0/16", "SubnetLen": 24, "Backend": {"Type": "vxlan"}} # Set flannel network
etcdctl get /docker-flannel/network/config # Check flannel network config
systemctl daemon-reload
systemctl restart flannel

# Check the flannel running status
cat /run/flannel/subnet.env
# You should see something like this
##############################################################################
FLANNEL_NETWORK=10.3.0.0/16
FLANNEL_SUBNET=10.3.30.1/24 # 10.3.50.1 on host2
FLANNEL_MTU=1450
FLANNEL_IPMASQ=false
##############################################################################
```

Just like Task1, we need to add the configuration of flannel into the docker runtime,

```bash
vim /lib/systemd/system/docker.service

# modify this line by adding the exec options
##############################################################################
ExecStart=/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock --bip=10.3.30.1//24 --ip-masq=true --mtu=1450 # 10.3.50.1 on host2
##############################################################################

systemctl daemon-reload
systemctl restart docker
```

Then in the container, we can see that the container now has an IP address within the subnet.

![image-20210716120916003](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716120916003.png)

And the container on host1 now can communicate with the container on host2.

![image-20210716121020863](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716121020863.png)



### 3. Cross-Host IPTV

Now that the basic communication is enabled, we are going to deploy some applications on our containers. We will first depoly IPTV, as its deployment is the most straightforward.

#### IPTV Server

First we create an IPTV-server container on host2,

```bash
docker run -it -d --name iptv_grpcweb ubuntu:latest /bin/bash
docker start iptv_grpcweb
docker attach iptv_grpcweb

# In the container, install dependencies
git clone https://github.com/free5gc/IPTV.git
apt udpdate
apt install -y ffmpeg nodejs curl
curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
apt update && apt install -y yarn

apt install software-properties-common # instead of python-software-properties
add-apt-repository ppa:gias-kay-lee/npm
apt-get update
apt-get install npm

wget https://dl.google.com/go/go1.12.9.linux-amd64.tar.gz
sudo tar -C /usr/local -zxvf go1.12.9.linux-amd64.tar.gz
mkdir -p ~/go/{bin,pkg,src}
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export GOROOT=/usr/local/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin:$GOROOT/bin' >> ~/.bashrc
echo 'export GO111MODULE=on' >> ~/.bashrc
echo 'export GOPROXY=https://goproxy.io' >> ~/.bashrc 
source ~/.bashrc

go version # Check go version
# go version go1.12.9 linux/amd64

# Install go packages
go get -u github.com/gin-contrib/static
go get -u github.com/gin-gonic/gin
go get -u github.com/urfave/cli
go get -u gopkg.in/yaml.v2

# Build Web Client
cd IPTV/web-client
yarn install
yarn build
cd ..
vim iptvcfg.conf # Configure iptv details, change IP to the container's IP

vim iptv.go
##############################################################################
	"github.com/free5gc/IPTV/factory"
    "github.com/free5gc/IPTV/iptv-server"
    "github.com/free5gc/IPTV/version"
##############################################################################
vim iptv-server/iptv_server.go
##############################################################################
	"github.com/free5gc/IPTV/factory"
	"github.com/free5gc/IPTV/iptv-server/hls-channel"
##############################################################################

mkdir hls        # Create chache folder, default is ./hls
go run iptv.go
```

Now we can see the IPTV server is listening and serving HTTP on 10.3.50.3:8888

![image-20210716133830364](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716133830364.png)



#### IPTV Client

Then we create an IPTV client container on host2,

```bash
docker run -v /tmp/.X11-unix:/tmp/.X11-unix -e DISPLAY=$DISPLAY -h $HOSTNAME -v $HOME/.Xauthority:/home/li/.Xauthority -itd --name=firefox ubuntu:latest
docker start firefox
docker attach firefox
xhost + # access control disabled, clients can connect from any host

# In the container
apt update
apt install -y firefox xorg openbox # install firefox and x11 components
firefox # Open firefox browser, it should be a GUI
```

In the firefox, enter `<IPTV Server IP>:8888`, we can see the GUI of IPTV,

![image-20210716140642560](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716140642560.png)

And on the server side, we have the logs,

![image-20210716140721199](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716140721199.png)

We can confirm that cross-host IPTV works well.



### 4. Cross-Host gRPC-Web

Then we will deploy gRPC-Web, as its setup is similar to IPTV. We deploy them in the same containers.

#### gRPC-Web Server

Using the same container `iptv_grpcweb` on host2,

```bash
# In the container for server
# Install dependencies
git clone https://github.com/grpc/grpc-web.git
wget https://github.com/protocolbuffers/protobuf/releases/download/v3.17.3/protoc-3.17.3-linux-x86_64.zip
unzip protoc-3.17.3-linux-x86_64.zip
mv bin/protoc /usr/bin
wget https://github.com/grpc/grpc-web/releases/download/1.2.1/protoc-gen-grpc-web-1.2.1-linux-x86_64
mv protoc-gen-grpc-web-1.2.1-linux-x86_64 /usr/bin/protoc-gen-grpc-web
wget https://github.com/improbable-eng/grpc-web/releases/download/v0.14.0/grpcwebproxy-v0.14.0-linux-x86_64.zip
mv dist/grpcwebproxy-v0.14.0-linux-x86_64 /usr/bin/grpcwebproxy
rm -rf dist bin include readme.txt # Clear the remaining file

# Generate Protobuf Messages and Client Service Stub
cd grpc-web/net/grpc/gateway/examples/helloworld/
protoc -I=. helloworld.proto \
  --js_out=import_style=commonjs:. \
  --grpc-web_out=import_style=commonjs,mode=grpcwebtext:.
# This will generate helloworld_pb.js and helloworld_grpc_web_pb.js under the same directory

# Compile the Client JavaScript Code
npm install # This took long time on my computer
npx webpack client.js

# Run the Example
node server.js &
python3 -m http.server 8081 &
grpcwebproxy \
    --backend_addr=localhost:9090 \
    --run_tls_server=false \
    --allow_all_origins
```

Now the grpc-web is running,

![image-20210716144230968](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716144230968.png)



#### gRPC-Web Client

Using the same container `firefox` on host1,

```bash
# In the container for client
firefox
```

In the firefox, enter `<IPTV Server IP>:8081`, and press F12 to open the console,

![image-20210716144438933](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716144438933.png)

We receive the the message "Hello! World", which yields that our gRPC-web server and client work well.



### 5. Cross-Host Minecraft

Now we come to the most entertaining but also the most problem-prone part, Minecraft. 

#### Minecraft Server

The image of Minecraft server we used is built by kitematic, and we directly pulled it from Docker Hub,

```bash
docker pull kitematic/minecraft 
docker run --name=mc_server kitematic/minecraft
docker exec -it mc_server bash

# In the docker
ps aux | grep java | grep -v grep | grep -v sh
```

![image-20210716150319093](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716150319093.png)

We can see that the Minecraft is running, but we need to change some configuration,

```bash
sed -i '/online-mode/s/true/false/g' /data/server.properties # Set online-mode to false
exit
# Exit the container
# On host2, restart mc_server
docker kill mc_server
docker start mc_server
```

Now the Minecraft Server should be running.



#### Minecraft Client

We create a new container `mc_client` on host1 supporting GUI,

```bash
xhost +
docker run -v /tmp/.X11-unix:/tmp/.X11-unix -e DISPLAY=$DISPLAY -h $HOSTNAME -v $HOME/.Xauthority:/home/li/.Xauthority -itd --name=mc_client ubuntu:latest
docker start mc_client
docker attach mc_client

# In the container, install dependencies
apt update && apt install -y wget xorg openbox # Install wget and X11 components
apt install -y openjdk-11-jdk openjdk-8-jdk # Install java11 and java8
wget http://ci.huangyuhui.net/job/HMCL/188/artifact/HMCL/build/libs/HMCL-3.3.188.jar # Install HMCL

# Open HMCL launcher
java -jar HMCL-3.3.188.jar
```

Then in HMCL launcher, we install a new game of version 1.12.2.

![image-20210716151543968](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716151543968.png)

Also, we should change the java directory in the global game settings to java8, as Minecraft of 1.12.2 version does not support java11.

![image-20210716151703202](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716151703202.png)

After configuration, we can start playing Minecraft in the container!

![image-20210716152008575](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716152008575.png)

We choose "Multiplayer" mode, and add a new server of our Minecraft Server's IP address,

![image-20210716152152915](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716152152915.png)We can see that the Minecraft Server is working well, then we can join this server.

![image-20210716152510595](C:\Users\ancientmodern\AppData\Roaming\Typora\typora-user-images\image-20210716152510595.png)

Enjoy the Minecraft!



### Discussion

+ When changing the flannel config, although the config stored in the etcd cluster has been changed, the network interface still remains the same. Therefore, when we want to change the flannel config, we need to delete the original network interface `flannel.1` by,

```bash
ifconfig flannel.1 down
ip link delete flannel.1
systemctl restart flannel
```

+ When installing Minecraft client, we need to make sure that the X11 is completely installed. As the HMCL launcher can be opened correctly, I think the X11 has already been installed well, however, when I try to launch the Minecraft client, I got an error `java.lang.ArrayIndexOutOfBoundsException`. The suggestion of Mojang is updating/reinstalling the graphic driver, while in our case is installing X11 components completely, and type `xhost +` in the host machine.




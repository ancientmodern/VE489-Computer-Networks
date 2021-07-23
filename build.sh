#!/bin/bash

make docker
sshpass -p 1 scp output/client root@10.3.46.2:~/

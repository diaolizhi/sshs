# sshs

![](./demo.gif)

# 说明
- 下载二进制文件，重命名为 sshs
- 编写配置文件 servers.txt
- 执行 `./sshs -c ./servers.txt` 

# 配置文件
- 一个服务器一行，格式为：`用户名@ip:端口#备注#密钥路径`
- 非必填项：端口、密钥路径
```txt
root@1.1.1.1#server-01
root@2.2.2.2:2222#server-02
root@3.3.3.3:2222#server-02#~/.ssh/id_rsa
```

# 参考
- [ssh-select](https://github.com/iwittkau/ssh-select)
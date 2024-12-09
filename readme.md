## 此程序是实现阿里云DDNS。&#x20;

市面上能找到的此类程序大部分都不支持阿里云子域名更新。花了点时间写了本程序。

只需要修改config.json目录下的默认值;

```

    "AccessKeyID": "AccessKeyID",
    "AccessKeySecret": "AccessKeySecret",
    "DomainName": "home.example.com",
    "Record": "@",
    "RecordType": "A"


accessKeyId： 改为阿里云的 accessKey;

accessKeySecret： 改为阿里云的 accessKeySecret

DomainName： “home.example.com”，改成你的域名，支持三级或者更多级子域名;

record： “@”，子域名设置，一般是www或者@.或者你想设置的任何子域名

RecordType： “A” A记录。就是更新的你IPV4地址。
```

### &#x20;用法：

&#x20;1.windows下，解压压缩包，修改config.json目录下的默认值。双击运行程序，程序会自动更新阿里云的DNS记录。

也可以用命令行加参数运行：&#x20;

    aliddns.exe -c config.json 

定时执行可以用windows计划任务。

&#x20;2.linux下，解压压缩包，将config.json复制到/etc/aliddns/目录下。aliddns复制到usr/bin目录下,

加执行权限。

    chmod +x aliddns

执行程序

    aliddns -c /etc/aliddns/config.json

可以用crontab定时运行程序。 参考的crontab命令：

     */5 * * * * aliddns -c /etc/aliddns/config.json

### &#x20;注意：

1.程序会自动判断当前的IP地址和DNS记录的IP是否一致，如果一致则不更新。&#x20;

2.程序会自动判断当前的IP地址是否是公网IP地址，如果不是则不更新。&#x20;

### 2024年12月9日更新&#x20;

增加了更新判断，如果要更新的ip和DNS记录的IP一致，则不更新。

此程序是实现阿里云DDNS。

市面上能找到的此类程序大部分都不支持阿里云子域名更新。花了点时间写了本程序。&#x20;

只需要修改config.json目录下的默认值&#x20;

accessKeyId： 改为阿里云的 accessKey&#x20;

accessKeySecret： 改为阿里云的 accessKeySecret

DomainName： “home.example.com”，改成你的域名，支持三级或者更多级子域名&#x20;

record： “@”，子域名设置，一般是www或者@.或者你想设置的任何子域名

RecordType： “A” A记录。就是更新的你IPV4地址。

2024年12月9日更新
增加了更新判断，如果要更新的ip和DNS记录的IP一致，则不更新。

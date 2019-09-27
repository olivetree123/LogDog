# LogDog

###
LogDog 是一个日志采集、解析和上报的工具。  
类似于 logstash，但是比 logstash 小巧。  
类似于 filebeat，但是能够解析日志。

### Install
```javascript
wget -O - https://olivetree.oss-cn-hangzhou.aliyuncs.com/install.sh | bash
```

### Configuration
```javascript
// vim /etc/logdog/logdog.toml
[input]
    [input.accesslog]
        type = "file"
        path = "/Users/gao/test111"
        format = "text"
        regex = "(?P<level>\\w+) (?P<time>\\S+) \\w+ (?P<name>\\w+)"


[handler]
    [handler.accesslog]
         script_path = "/Users/gao/code/gowork/src/logDog/myHandler.lua"
        [handler.accesslog.add_data]
            school_id = "123"
            class_id = "123"

[output]
    [output.other]
        type = "stdout"
    [output.accesslog2]
        type = "http"
        http_url = "http://localhost:9000"
        [output.accesslog2.headers]
            Auth = "123456"
    [output.accesslog]
        type = "redis"
        redis_addr = "localhost:6379"
        redis_db = 0
        redis_key = "logdog"

```

### Run
```javascript
logdog
```
# exporterpush
Remote data writing tool.(remote write data to prometheus and pushgateway)

# 用途
推送exporter数据到Prometheus、Pushgateway中，除了上述的目标数据源外还对接了腾讯的barad监控系统的指标推送。

# 使用方式
### 编译
```
go build -o exporterpush
```

### 启动
```
./exporterpush -config config.yaml
```

# 配置文件
```
# my global config
global:
  scrape_interval: 15 # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  disk: /dev/vda
  net_interface: eth0
  scrape_target_types:
    node_exporter: http://127.0.0.1:9100/metrics #--指定抓取的exporter路径(目前这里暂时支持node_exporter和ck的exporter)
    clickhouse_exporter: http://127.0.0.1:9363/metrics
  log:
    log_save_path: /tmp/logs #--日志路径
    log_file_name: exportpush #-日志文件名
    log_file_ext: .log。     #--日志文件后缀

# push plugin configuration
barad:   #--------- 腾讯barad监控系统对接配置
  is_use: false
  app_id: 1
  instance_id: 22
  node_id: 33
  project_id: 0
  namespace: pce/upclickhouse
  static_configs:
    - destination:
        - http://xxx.barad.tencentyun.com/upclickhouse.cgi

prometheus: #---数据写入远程prometheus配置
  is_use: true
  static_configs:
    - destination:
        - http://127.0.0.1:9090/api/v1/write
      labels:
        cluster_name: test #--自定标签，推送指标的时候添加自定义的标签

pushgateway: #---数据写入远程pushgateway配置
  is_use: false
  job_name: push_job
  static_configs:
    - destination:
        - http://127.0.0.1:9091
      labels:
        cluster_name: test 

```

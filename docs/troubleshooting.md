# 故障排查指南

本文档提供Task Scheduler System常见问题的排查和解决方案。

## 目录

- [服务启动问题](#服务启动问题)
- [任务执行问题](#任务执行问题)
- [数据库问题](#数据库问题)
- [分布式锁问题](#分布式锁问题)
- [网络和连接问题](#网络和连接问题)
- [性能问题](#性能问题)
- [监控和日志问题](#监控和日志问题)

## 服务启动问题

### 问题1: 服务无法启动

**症状**:
```
Error: Failed to start server: listen tcp :8080: bind: address already in use
```

**原因**: 端口已被占用

**解决方案**:

```bash
# 查找占用端口的进程
netstat -tulpn | grep 8080
# 或
lsof -i :8080

# 终止进程
kill -9 <PID>

# 或修改配置使用其他端口
vim config.yaml
# 修改 server.http_port 为其他端口
```

### 问题2: 配置文件加载失败

**症状**:
```
Error: Failed to load configuration: config file not found
```

**原因**: 配置文件路径错误或文件不存在

**解决方案**:

```bash
# 检查配置文件是否存在
ls -la config.yaml

# 使用绝对路径指定配置文件
./task-scheduler -config /path/to/config.yaml

# 或复制示例配置
cp config.yaml.example config.yaml
```

### 问题3: 数据库连接失败

**症状**:
```
Error: Failed to connect to database: dial tcp 127.0.0.1:3306: connect: connection refused
```

**原因**: 数据库未启动或连接信息错误

**解决方案**:

```bash
# 检查MySQL是否运行
systemctl status mysql
# 或
docker-compose ps mysql

# 测试数据库连接
mysql -h localhost -u scheduler -p

# 检查配置文件中的数据库信息
grep -A 10 "database:" config.yaml

# 启动数据库
docker-compose up -d mysql
```

### 问题4: Redis连接失败

**症状**:
```
Error: Failed to connect to Redis: dial tcp 127.0.0.1:6379: connect: connection refused
```

**原因**: Redis未启动或连接信息错误

**解决方案**:

```bash
# 检查Redis是否运行
systemctl status redis
# 或
docker-compose ps redis

# 测试Redis连接
redis-cli ping

# 启动Redis
docker-compose up -d redis
```

### 问题5: 权限不足

**症状**:
```
Error: Permission denied
```

**原因**: 用户权限不足

**解决方案**:

```bash
# 检查文件权限
ls -la task-scheduler

# 添加执行权限
chmod +x task-scheduler

# 检查日志目录权限
ls -la /var/log/task-scheduler.log

# 创建日志目录并设置权限
sudo mkdir -p /var/log
sudo chown scheduler:scheduler /var/log/task-scheduler.log
```

## 任务执行问题

### 问题6: 任务一直处于pending状态

**症状**: 任务创建成功但从不执行

**可能原因**:
1. 调度器未启动
2. 工作线程池已满
3. 分布式锁未释放
4. 节点不健康

**排查步骤**:

```bash
# 1. 检查调度器状态
curl http://localhost:8080/health

# 2. 查看日志
docker-compose logs -f task-scheduler-1 | grep "scheduler"

# 3. 检查任务队列
redis-cli LLEN scheduler:task:queue

# 4. 检查分布式锁
redis-cli KEYS "scheduler:lock:*"

# 5. 检查节点健康状态
redis-cli SMEMBERS "scheduler:nodes"

# 6. 查看工作线程状态
# 在日志中搜索 "worker pool"
```

**解决方案**:

```bash
# 如果是锁未释放
redis-cli DEL "scheduler:lock:task:123"

# 如果是节点不健康
# 重启节点
docker-compose restart task-scheduler-1

# 如果是工作线程池已满
# 增加工作线程数
vim config.yaml
# 修改 scheduler.worker_pool_size
docker-compose restart task-scheduler-1
```

### 问题7: 任务重复执行

**症状**: 同一个任务被多次执行

**可能原因**:
1. 分布式锁失效
2. 锁TTL过短
3. 多个节点同时获取到任务
4. 时钟不同步

**排查步骤**:

```bash
# 1. 检查任务执行记录
mysql -e "SELECT * FROM task_execution_history WHERE task_id='task-123' ORDER BY created_at"

# 2. 检查分布式锁配置
grep -A 5 "lock:" config.yaml

# 3. 检查节点时钟
date
# 在所有节点上执行

# 4. 查看Redis锁信息
redis-cli GET "scheduler:lock:task:123"
redis-cli TTL "scheduler:lock:task:123"
```

**解决方案**:

```bash
# 1. 增加锁TTL
vim config.yaml
# 修改 lock.redis.ttl 为更大的值（如 5m）

# 2. 同步时钟
sudo ntpdate -s time.nist.gov
# 或配置NTP服务

# 3. 检查网络延迟
ping redis-host

# 4. 重启服务
docker-compose restart
```

### 问题8: 任务执行失败但未重试

**症状**: 任务失败后直接标记为failed，没有重试

**可能原因**:
1. 未配置重试策略
2. 重试次数已达上限
3. 重试策略配置错误

**排查步骤**:

```bash
# 1. 查看任务配置
curl http://localhost:8080/tasks/task-123 | jq '.data.retry_policy'

# 2. 查看任务重试次数
curl http://localhost:8080/tasks/task-123 | jq '.data.retry_count'

# 3. 查看日志
docker-compose logs task-scheduler-1 | grep "task-123" | grep "retry"
```

**解决方案**:

```bash
# 创建任务时配置重试策略
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "任务",
    "execution_mode": "immediate",
    "retry_policy": {
      "max_retries": 3,
      "retry_interval": "10s",
      "backoff_factor": 2.0
    }
  }'
```

### 问题9: 回调失败

**症状**: 任务执行成功但回调失败

**可能原因**:
1. 回调URL不可达
2. 回调超时
3. 服务发现失败
4. 网络问题

**排查步骤**:

```bash
# 1. 测试回调URL
curl -X POST http://callback-url/endpoint

# 2. 查看回调日志
docker-compose logs task-scheduler-1 | grep "callback"

# 3. 检查服务发现
# 如果使用Consul
curl http://localhost:8500/v1/catalog/service/my-service

# 4. 检查网络连通性
ping callback-host
telnet callback-host 80
```

**解决方案**:

```bash
# 1. 增加回调超时时间
# 在创建任务时设置
"callback_config": {
  "timeout": "60s"
}

# 2. 检查服务发现配置
vim config.yaml
# 检查 discovery 配置

# 3. 使用静态地址
"callback_config": {
  "url": "http://direct-ip:port/endpoint",
  "discovery_type": "static"
}
```

### 问题10: 子任务不执行

**症状**: 父任务创建成功但子任务不执行

**可能原因**:
1. 并发限制设置为0
2. 父任务未启动
3. 子任务配置错误

**排查步骤**:

```bash
# 1. 查看父任务配置
curl http://localhost:8080/tasks/parent-task-id | jq '.data.concurrency_limit'

# 2. 查看子任务列表
curl http://localhost:8080/tasks?parent_id=parent-task-id

# 3. 查看子任务状态
curl http://localhost:8080/tasks/sub-task-id/status
```

**解决方案**:

```bash
# 设置合理的并发限制
"concurrency_limit": 5  # 允许5个子任务并发执行

# 或设置为0表示无限制
"concurrency_limit": 0
```

## 数据库问题

### 问题11: 连接池耗尽

**症状**:
```
Error: Error 1040: Too many connections
```

**原因**: 数据库连接数达到上限

**排查步骤**:

```bash
# 1. 查看当前连接数
mysql -e "SHOW PROCESSLIST"

# 2. 查看最大连接数
mysql -e "SHOW VARIABLES LIKE 'max_connections'"

# 3. 查看应用连接池配置
grep -A 5 "database:" config.yaml
```

**解决方案**:

```bash
# 1. 增加MySQL最大连接数
# 编辑 /etc/mysql/my.cnf
[mysqld]
max_connections = 500

# 重启MySQL
systemctl restart mysql

# 2. 调整应用连接池
vim config.yaml
database:
  max_open_conns: 50
  max_idle_conns: 10

# 3. 检查连接泄漏
# 查看长时间运行的查询
mysql -e "SELECT * FROM information_schema.processlist WHERE time > 60"
```

### 问题12: 慢查询

**症状**: 数据库查询响应慢

**排查步骤**:

```bash
# 1. 启用慢查询日志
mysql -e "SET GLOBAL slow_query_log = 'ON'"
mysql -e "SET GLOBAL long_query_time = 2"

# 2. 查看慢查询
mysql -e "SELECT * FROM mysql.slow_log ORDER BY query_time DESC LIMIT 10"

# 3. 分析查询
mysql -e "EXPLAIN SELECT * FROM tasks WHERE status = 'pending'"

# 4. 查看表大小
mysql -e "SELECT table_name, table_rows, data_length, index_length 
          FROM information_schema.tables 
          WHERE table_schema = 'task_scheduler'"
```

**解决方案**:

```bash
# 1. 添加索引
mysql -e "CREATE INDEX idx_tasks_status ON tasks(status)"
mysql -e "CREATE INDEX idx_tasks_created_at ON tasks(created_at)"

# 2. 优化查询
# 避免SELECT *，只查询需要的字段
# 使用LIMIT限制结果集

# 3. 定期清理历史数据
mysql -e "DELETE FROM tasks 
          WHERE status IN ('success', 'failed') 
          AND completed_at < DATE_SUB(NOW(), INTERVAL 30 DAY)"

# 4. 优化表
mysql -e "OPTIMIZE TABLE tasks"
```

### 问题13: 数据库锁等待

**症状**:
```
Error: Lock wait timeout exceeded
```

**排查步骤**:

```bash
# 1. 查看锁等待
mysql -e "SHOW ENGINE INNODB STATUS\G" | grep -A 20 "TRANSACTIONS"

# 2. 查看正在执行的事务
mysql -e "SELECT * FROM information_schema.innodb_trx"

# 3. 查看锁等待
mysql -e "SELECT * FROM information_schema.innodb_lock_waits"
```

**解决方案**:

```bash
# 1. 终止阻塞的事务
mysql -e "KILL <thread_id>"

# 2. 增加锁等待超时时间
mysql -e "SET GLOBAL innodb_lock_wait_timeout = 120"

# 3. 优化事务
# 减小事务大小
# 避免长时间持有锁
# 使用合适的隔离级别
```

## 分布式锁问题

### 问题14: 锁获取失败

**症状**: 无法获取分布式锁

**可能原因**:
1. Redis连接失败
2. 锁已被其他节点持有
3. 锁TTL配置不当

**排查步骤**:

```bash
# 1. 检查Redis连接
redis-cli ping

# 2. 查看锁状态
redis-cli GET "scheduler:lock:task:123"
redis-cli TTL "scheduler:lock:task:123"

# 3. 查看所有锁
redis-cli KEYS "scheduler:lock:*"

# 4. 查看日志
docker-compose logs task-scheduler-1 | grep "lock"
```

**解决方案**:

```bash
# 1. 手动释放锁（谨慎操作）
redis-cli DEL "scheduler:lock:task:123"

# 2. 等待锁过期
# 锁会在TTL后自动过期

# 3. 调整锁配置
vim config.yaml
lock:
  redis:
    ttl: 5m
    retry_interval: 1s
    max_retries: 5
```

### 问题15: 锁未释放

**症状**: 任务完成后锁仍然存在

**可能原因**:
1. 程序异常退出
2. 释放锁失败
3. 网络问题

**排查步骤**:

```bash
# 1. 查看锁
redis-cli KEYS "scheduler:lock:*"

# 2. 查看锁的TTL
redis-cli TTL "scheduler:lock:task:123"

# 3. 查看任务状态
curl http://localhost:8080/tasks/task-123/status
```

**解决方案**:

```bash
# 1. 清理过期锁
redis-cli --scan --pattern "scheduler:lock:*" | while read key; do
    ttl=$(redis-cli TTL "$key")
    if [ "$ttl" -eq -1 ]; then
        echo "Deleting key without TTL: $key"
        redis-cli DEL "$key"
    fi
done

# 2. 设置合理的TTL
# 确保锁有TTL，避免永久锁

# 3. 实现锁自动续期
# 系统已实现，确保启用
```

## 网络和连接问题

### 问题16: 服务间通信失败

**症状**: 节点间无法通信

**排查步骤**:

```bash
# 1. 检查网络连通性
ping node-2-ip

# 2. 检查端口
telnet node-2-ip 8080
telnet node-2-ip 9090

# 3. 检查防火墙
sudo ufw status
sudo iptables -L

# 4. 检查Docker网络
docker network ls
docker network inspect task-scheduler_task-scheduler-network
```

**解决方案**:

```bash
# 1. 开放必要端口
sudo ufw allow 8080/tcp
sudo ufw allow 9090/tcp

# 2. 检查Docker网络配置
# 确保所有容器在同一网络

# 3. 使用服务名而不是IP
# 在Docker Compose中使用服务名
```

### 问题17: DNS解析失败

**症状**: 无法解析服务名

**排查步骤**:

```bash
# 1. 测试DNS解析
nslookup mysql-service
dig mysql-service

# 2. 检查/etc/hosts
cat /etc/hosts

# 3. 检查DNS配置
cat /etc/resolv.conf
```

**解决方案**:

```bash
# 1. 添加hosts记录
echo "10.0.0.10 mysql-service" | sudo tee -a /etc/hosts

# 2. 配置DNS服务器
# 编辑 /etc/resolv.conf
nameserver 8.8.8.8
nameserver 8.8.4.4

# 3. 使用IP地址
# 在配置中使用IP而不是域名
```

## 性能问题

### 问题18: CPU使用率高

**症状**: CPU使用率持续高于80%

**排查步骤**:

```bash
# 1. 查看进程CPU使用
top -p $(pgrep task-scheduler)

# 2. 生成CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof

# 3. 分析profile
go tool pprof cpu.prof
# 在pprof中执行: top, list <function>

# 4. 查看goroutine数量
curl http://localhost:8080/debug/pprof/goroutine?debug=1
```

**解决方案**:

```bash
# 1. 优化热点代码
# 根据profile结果优化

# 2. 减少工作线程数
vim config.yaml
scheduler:
  worker_pool_size: 10

# 3. 增加任务轮询间隔
scheduler:
  task_poll_interval: 10s

# 4. 水平扩展
# 增加节点数量
```

### 问题19: 内存使用高

**症状**: 内存使用持续增长

**排查步骤**:

```bash
# 1. 查看内存使用
docker stats --no-stream

# 2. 生成内存profile
curl http://localhost:8080/debug/pprof/heap > heap.prof

# 3. 分析profile
go tool pprof heap.prof
# 在pprof中执行: top, list <function>

# 4. 检查goroutine泄漏
curl http://localhost:8080/debug/pprof/goroutine?debug=1 | grep "goroutine profile"
```

**解决方案**:

```bash
# 1. 重启服务（临时方案）
docker-compose restart task-scheduler-1

# 2. 修复内存泄漏
# 根据profile结果修复代码

# 3. 限制内存使用
# docker-compose.yml
services:
  task-scheduler-1:
    deploy:
      resources:
        limits:
          memory: 1G

# 4. 调整GC参数
export GOGC=50  # 更频繁的GC
```

### 问题20: 响应时间慢

**症状**: API响应时间超过1秒

**排查步骤**:

```bash
# 1. 测量响应时间
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/tasks

# curl-format.txt:
# time_total: %{time_total}\n

# 2. 查看慢查询
mysql -e "SELECT * FROM mysql.slow_log ORDER BY query_time DESC LIMIT 10"

# 3. 检查Redis延迟
redis-cli --latency

# 4. 查看系统负载
uptime
iostat
```

**解决方案**:

```bash
# 1. 添加数据库索引
mysql -e "CREATE INDEX idx_tasks_status ON tasks(status)"

# 2. 使用缓存
# 实现查询结果缓存

# 3. 优化查询
# 减少JOIN
# 使用LIMIT
# 只查询需要的字段

# 4. 增加资源
# 增加CPU和内存
```

## 监控和日志问题

### 问题21: 日志文件过大

**症状**: 日志文件占用大量磁盘空间

**解决方案**:

```bash
# 1. 配置日志轮转
# /etc/logrotate.d/task-scheduler
/var/log/task-scheduler.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
}

# 2. 手动轮转
logrotate -f /etc/logrotate.d/task-scheduler

# 3. 清理旧日志
find /var/log -name "task-scheduler*.log*" -mtime +7 -delete

# 4. 调整日志级别
vim config.yaml
log:
  level: info  # 从debug改为info
```

### 问题22: Prometheus指标缺失

**症状**: Grafana显示"No data"

**排查步骤**:

```bash
# 1. 检查指标端点
curl http://localhost:9091/metrics

# 2. 检查Prometheus配置
cat prometheus.yml

# 3. 检查Prometheus targets
curl http://localhost:9090/api/v1/targets

# 4. 查看Prometheus日志
docker-compose logs prometheus
```

**解决方案**:

```bash
# 1. 确保指标端点可访问
curl http://localhost:9091/metrics

# 2. 更新Prometheus配置
vim prometheus.yml
scrape_configs:
  - job_name: 'task-scheduler'
    static_configs:
      - targets: ['task-scheduler-1:9091']

# 3. 重启Prometheus
docker-compose restart prometheus

# 4. 检查防火墙
sudo ufw allow 9091/tcp
```

## 获取帮助

如果以上方案无法解决您的问题，请：

1. 查看[完整文档](../README.md)
2. 搜索[GitHub Issues](https://github.com/yourusername/task-scheduler/issues)
3. 创建新Issue并提供：
   - 详细的问题描述
   - 重现步骤
   - 相关日志
   - 环境信息
4. 加入[讨论区](https://github.com/yourusername/task-scheduler/discussions)
5. 发送邮件至 support@example.com

## 参考资源

- [运维指南](operations.md)
- [配置文档](configuration.md)
- [监控文档](monitoring-observability.md)
- [API文档](api-http.md)


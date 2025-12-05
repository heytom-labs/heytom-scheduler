# 运维指南

本文档提供Task Scheduler System的日常运维操作指南。

## 目录

- [日常运维](#日常运维)
- [监控和告警](#监控和告警)
- [备份和恢复](#备份和恢复)
- [故障排查](#故障排查)
- [性能调优](#性能调优)
- [安全加固](#安全加固)
- [升级和回滚](#升级和回滚)

## 日常运维

### 服务状态检查

#### 检查服务健康状态

```bash
# HTTP健康检查
curl http://localhost:8080/health

# 检查所有节点
for port in 8080 8081; do
    echo "Checking node on port $port"
    curl http://localhost:$port/health
done
```

#### 检查服务进程

```bash
# 检查进程是否运行
ps aux | grep task-scheduler

# 检查端口监听
netstat -tulpn | grep -E '8080|9090|9091'

# 使用systemd检查（如果使用systemd管理）
systemctl status task-scheduler
```

#### 检查Docker容器

```bash
# 查看容器状态
docker-compose ps

# 查看容器健康状态
docker inspect --format='{{.State.Health.Status}}' task-scheduler-node-1

# 查看容器资源使用
docker stats --no-stream
```

### 日志管理

#### 查看日志

```bash
# 查看实时日志
tail -f /var/log/task-scheduler.log

# 查看Docker日志
docker-compose logs -f task-scheduler-1

# 查看最近100行日志
docker-compose logs --tail=100 task-scheduler-1

# 查看特定时间段的日志
docker-compose logs --since="2024-01-01T00:00:00" --until="2024-01-02T00:00:00" task-scheduler-1
```

#### 日志轮转

配置logrotate：

```bash
# /etc/logrotate.d/task-scheduler
/var/log/task-scheduler.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 scheduler scheduler
    postrotate
        systemctl reload task-scheduler
    endscript
}
```

#### 日志分析

```bash
# 统计错误日志
grep '"level":"error"' /var/log/task-scheduler.log | wc -l

# 查找特定任务的日志
grep 'task_id":"550e8400-e29b-41d4-a716-446655440000"' /var/log/task-scheduler.log

# 统计任务状态分布
grep '"message":"Task execution completed"' /var/log/task-scheduler.log | \
    jq -r '.status' | sort | uniq -c
```

### 数据库维护

#### MySQL维护

```bash
# 连接数据库
mysql -h localhost -u scheduler -p task_scheduler

# 检查表状态
SHOW TABLE STATUS;

# 优化表
OPTIMIZE TABLE tasks;

# 检查表完整性
CHECK TABLE tasks;

# 修复表
REPAIR TABLE tasks;

# 查看慢查询
SELECT * FROM mysql.slow_log ORDER BY query_time DESC LIMIT 10;
```

#### 数据清理

```bash
# 清理30天前的已完成任务
DELETE FROM tasks 
WHERE status IN ('success', 'failed', 'cancelled') 
AND completed_at < DATE_SUB(NOW(), INTERVAL 30 DAY);

# 清理孤立的子任务
DELETE FROM tasks 
WHERE parent_id IS NOT NULL 
AND parent_id NOT IN (SELECT id FROM tasks);
```

### Redis维护

```bash
# 连接Redis
redis-cli

# 查看信息
INFO

# 查看内存使用
INFO memory

# 查看所有键
KEYS *

# 查看特定前缀的键
KEYS scheduler:lock:*

# 清理过期键
redis-cli --scan --pattern "scheduler:lock:*" | xargs redis-cli DEL

# 持久化数据
SAVE
```

## 监控和告警

### Prometheus监控

#### 关键指标

```promql
# 任务成功率
rate(task_scheduler_tasks_completed_total{status="success"}[5m]) / 
rate(task_scheduler_tasks_completed_total[5m])

# 任务失败率
rate(task_scheduler_tasks_completed_total{status="failed"}[5m]) / 
rate(task_scheduler_tasks_completed_total[5m])

# 平均任务执行时间
rate(task_scheduler_task_execution_duration_seconds_sum[5m]) / 
rate(task_scheduler_task_execution_duration_seconds_count[5m])

# 任务队列大小
task_scheduler_task_queue_size

# 活跃节点数
task_scheduler_active_nodes_count

# 节点健康状态
task_scheduler_node_health_status
```

#### 告警规则

创建 `prometheus-alerts.yml`：

```yaml
groups:
  - name: task_scheduler
    interval: 30s
    rules:
      # 任务失败率过高
      - alert: HighTaskFailureRate
        expr: |
          rate(task_scheduler_tasks_completed_total{status="failed"}[5m]) / 
          rate(task_scheduler_tasks_completed_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "任务失败率过高"
          description: "任务失败率超过10%，当前值: {{ $value }}"

      # 任务队列积压
      - alert: TaskQueueBacklog
        expr: task_scheduler_task_queue_size > 1000
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "任务队列积压"
          description: "任务队列大小: {{ $value }}"

      # 节点不健康
      - alert: NodeUnhealthy
        expr: task_scheduler_node_health_status == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "节点不健康"
          description: "节点 {{ $labels.node_id }} 不健康"

      # 服务不可用
      - alert: ServiceDown
        expr: up{job="task-scheduler"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "服务不可用"
          description: "实例 {{ $labels.instance }} 不可用"

      # 任务执行时间过长
      - alert: SlowTaskExecution
        expr: |
          rate(task_scheduler_task_execution_duration_seconds_sum[5m]) / 
          rate(task_scheduler_task_execution_duration_seconds_count[5m]) > 60
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "任务执行时间过长"
          description: "平均任务执行时间: {{ $value }}秒"
```

### Grafana仪表板

#### 导入仪表板

1. 访问Grafana: `http://localhost:3001`
2. 登录（默认: admin/admin）
3. 点击 "+" -> "Import"
4. 上传仪表板JSON文件或输入仪表板ID

#### 关键面板

- **任务概览**: 任务创建/完成/失败数量
- **任务执行时间**: 任务执行时间分布
- **任务状态**: 各状态任务数量
- **节点状态**: 节点健康状态和负载
- **队列监控**: 任务队列大小趋势
- **API性能**: HTTP/gRPC请求延迟和吞吐量

### 告警通知

#### 配置Alertmanager

```yaml
# alertmanager.yml
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname', 'cluster']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'default'

receivers:
  - name: 'default'
    email_configs:
      - to: 'ops@example.com'
        from: 'alertmanager@example.com'
        smarthost: 'smtp.example.com:587'
        auth_username: 'alertmanager@example.com'
        auth_password: 'password'
    
    webhook_configs:
      - url: 'http://slack-webhook-url'
        send_resolved: true
```

## 备份和恢复

### 数据库备份

#### 自动备份脚本

```bash
#!/bin/bash
# backup-database.sh

BACKUP_DIR="/backup/mysql"
DATE=$(date +%Y%m%d-%H%M%S)
BACKUP_FILE="$BACKUP_DIR/task_scheduler_$DATE.sql"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份数据库
mysqldump -h localhost -u scheduler -p$DB_PASSWORD task_scheduler > $BACKUP_FILE

# 压缩备份
gzip $BACKUP_FILE

# 删除7天前的备份
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_FILE.gz"
```

#### 配置定时备份

```bash
# 添加到crontab
crontab -e

# 每天凌晨2点备份
0 2 * * * /opt/task-scheduler/scripts/backup-database.sh
```

#### Docker环境备份

```bash
#!/bin/bash
# backup-docker.sh

BACKUP_DIR="/backup/docker"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p $BACKUP_DIR

# 备份数据库
docker-compose exec -T mysql mysqldump -u root -p$MYSQL_ROOT_PASSWORD task_scheduler > \
    $BACKUP_DIR/database_$DATE.sql

# 备份Redis
docker-compose exec -T redis redis-cli SAVE
docker cp task-scheduler-redis-1:/data/dump.rdb $BACKUP_DIR/redis_$DATE.rdb

# 备份配置文件
cp config.yaml $BACKUP_DIR/config_$DATE.yaml

# 压缩备份
tar czf $BACKUP_DIR/backup_$DATE.tar.gz $BACKUP_DIR/*_$DATE.*

# 清理临时文件
rm $BACKUP_DIR/*_$DATE.{sql,rdb,yaml}

echo "Backup completed: $BACKUP_DIR/backup_$DATE.tar.gz"
```

### 数据恢复

#### 从备份恢复数据库

```bash
# 解压备份
gunzip task_scheduler_20240101-020000.sql.gz

# 恢复数据库
mysql -h localhost -u scheduler -p task_scheduler < task_scheduler_20240101-020000.sql

# Docker环境恢复
docker-compose exec -T mysql mysql -u scheduler -p task_scheduler < backup.sql
```

#### 恢复Redis数据

```bash
# 停止Redis
docker-compose stop redis

# 替换dump.rdb
docker cp redis_backup.rdb task-scheduler-redis-1:/data/dump.rdb

# 启动Redis
docker-compose start redis
```

### 灾难恢复

#### 完整恢复流程

1. **准备新环境**
   ```bash
   # 安装Docker和Docker Compose
   # 克隆代码仓库
   git clone https://github.com/yourusername/task-scheduler.git
   cd task-scheduler
   ```

2. **恢复配置**
   ```bash
   # 复制配置文件
   cp /backup/config.yaml .
   ```

3. **启动基础服务**
   ```bash
   # 启动MySQL和Redis
   docker-compose up -d mysql redis
   
   # 等待服务就绪
   sleep 30
   ```

4. **恢复数据**
   ```bash
   # 恢复数据库
   docker-compose exec -T mysql mysql -u scheduler -p task_scheduler < /backup/database.sql
   
   # 恢复Redis
   docker cp /backup/redis.rdb task-scheduler-redis-1:/data/dump.rdb
   docker-compose restart redis
   ```

5. **启动应用**
   ```bash
   # 启动所有服务
   docker-compose up -d
   
   # 验证服务
   curl http://localhost:8080/health
   ```

## 故障排查

### 常见问题

#### 1. 服务无法启动

**症状**: 服务启动失败或立即退出

**排查步骤**:

```bash
# 查看日志
docker-compose logs task-scheduler-1

# 检查配置文件
cat config.yaml

# 检查端口占用
netstat -tulpn | grep -E '8080|9090'

# 检查依赖服务
docker-compose ps mysql redis
```

**常见原因**:
- 配置文件错误
- 端口被占用
- 数据库连接失败
- Redis连接失败

#### 2. 任务不执行

**症状**: 任务创建成功但一直处于pending状态

**排查步骤**:

```bash
# 检查调度器状态
curl http://localhost:8080/health

# 查看任务队列
redis-cli LLEN scheduler:task:queue

# 查看工作线程
# 在日志中搜索 "worker"

# 检查分布式锁
redis-cli KEYS "scheduler:lock:*"
```

**常见原因**:
- 调度器未启动
- 工作线程池已满
- 分布式锁未释放
- 节点不健康

#### 3. 任务重复执行

**症状**: 同一个任务被多个节点执行

**排查步骤**:

```bash
# 检查分布式锁
redis-cli GET "scheduler:lock:task:123"

# 检查节点注册
redis-cli SMEMBERS "scheduler:nodes"

# 查看任务绑定
mysql -e "SELECT id, node_id FROM tasks WHERE id='task-123'"
```

**常见原因**:
- 分布式锁失效
- 锁TTL过短
- Redis连接不稳定
- 时钟不同步

#### 4. 数据库连接池耗尽

**症状**: 出现"too many connections"错误

**排查步骤**:

```bash
# 检查当前连接数
mysql -e "SHOW PROCESSLIST"

# 检查最大连接数
mysql -e "SHOW VARIABLES LIKE 'max_connections'"

# 检查连接池配置
grep -A 5 "database:" config.yaml
```

**解决方案**:
- 增加数据库最大连接数
- 调整应用连接池大小
- 检查连接泄漏

#### 5. 内存泄漏

**症状**: 内存使用持续增长

**排查步骤**:

```bash
# 查看内存使用
docker stats --no-stream

# 生成内存profile
curl http://localhost:8080/debug/pprof/heap > heap.prof

# 分析profile
go tool pprof heap.prof
```

**解决方案**:
- 重启服务
- 检查goroutine泄漏
- 优化代码

### 性能问题排查

#### CPU使用率高

```bash
# 查看CPU使用
top -p $(pgrep task-scheduler)

# 生成CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof

# 分析profile
go tool pprof cpu.prof
```

#### 响应时间慢

```bash
# 查看API延迟
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/tasks

# curl-format.txt内容:
# time_namelookup:  %{time_namelookup}\n
# time_connect:  %{time_connect}\n
# time_starttransfer:  %{time_starttransfer}\n
# time_total:  %{time_total}\n

# 查看慢查询
mysql -e "SELECT * FROM mysql.slow_log ORDER BY query_time DESC LIMIT 10"
```

## 性能调优

### 应用层优化

#### 调整工作线程池

```yaml
scheduler:
  worker_pool_size: 50  # 根据CPU核心数调整
```

#### 调整任务轮询间隔

```yaml
scheduler:
  task_poll_interval: 2s  # 减少延迟
```

#### 调整心跳间隔

```yaml
scheduler:
  heartbeat_interval: 5s
  heartbeat_timeout: 15s
```

### 数据库优化

#### 添加索引

```sql
-- 任务状态索引
CREATE INDEX idx_tasks_status ON tasks(status);

-- 任务创建时间索引
CREATE INDEX idx_tasks_created_at ON tasks(created_at);

-- 任务节点索引
CREATE INDEX idx_tasks_node_id ON tasks(node_id);

-- 复合索引
CREATE INDEX idx_tasks_status_created ON tasks(status, created_at);
```

#### 调整连接池

```yaml
database:
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime: 5m
```

#### 查询优化

```sql
-- 使用EXPLAIN分析查询
EXPLAIN SELECT * FROM tasks WHERE status = 'pending' ORDER BY created_at LIMIT 100;

-- 优化查询
SELECT id, name, status FROM tasks 
WHERE status = 'pending' 
AND created_at > DATE_SUB(NOW(), INTERVAL 1 HOUR)
ORDER BY created_at 
LIMIT 100;
```

### Redis优化

#### 调整内存策略

```bash
# redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
```

#### 启用持久化

```bash
# redis.conf
save 900 1
save 300 10
save 60 10000

appendonly yes
appendfsync everysec
```

### 系统层优化

#### 调整文件描述符限制

```bash
# /etc/security/limits.conf
scheduler soft nofile 65536
scheduler hard nofile 65536

# 验证
ulimit -n
```

#### 调整TCP参数

```bash
# /etc/sysctl.conf
net.core.somaxconn = 1024
net.ipv4.tcp_max_syn_backlog = 2048
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 30

# 应用配置
sysctl -p
```

## 安全加固

### 网络安全

#### 配置防火墙

```bash
# 允许必要端口
ufw allow 8080/tcp  # HTTP
ufw allow 9090/tcp  # gRPC
ufw allow 3306/tcp  # MySQL (仅内网)
ufw allow 6379/tcp  # Redis (仅内网)

# 启用防火墙
ufw enable
```

#### 使用TLS/SSL

```yaml
server:
  tls_enabled: true
  tls_cert_file: /etc/task-scheduler/server.crt
  tls_key_file: /etc/task-scheduler/server.key
```

### 访问控制

#### 数据库访问控制

```sql
-- 创建只读用户
CREATE USER 'scheduler_ro'@'%' IDENTIFIED BY 'password';
GRANT SELECT ON task_scheduler.* TO 'scheduler_ro'@'%';

-- 限制IP访问
CREATE USER 'scheduler'@'10.0.0.%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON task_scheduler.* TO 'scheduler'@'10.0.0.%';
```

#### Redis访问控制

```bash
# redis.conf
requirepass your-strong-password
bind 127.0.0.1 10.0.0.1
```

### 审计日志

#### 启用审计日志

```yaml
log:
  audit_enabled: true
  audit_file: /var/log/task-scheduler-audit.log
```

#### 审计事件

- 任务创建/删除
- 配置变更
- 用户登录/登出
- 权限变更

## 升级和回滚

### 滚动升级

#### 准备工作

```bash
# 备份数据
./scripts/backup-database.sh

# 备份配置
cp config.yaml config.yaml.backup

# 拉取新镜像
docker pull your-registry/task-scheduler:v1.1.0
```

#### 执行升级

```bash
# 逐个升级节点
docker-compose stop task-scheduler-1
docker-compose up -d task-scheduler-1

# 等待节点就绪
sleep 30
curl http://localhost:8080/health

# 升级下一个节点
docker-compose stop task-scheduler-2
docker-compose up -d task-scheduler-2
```

#### Kubernetes滚动升级

```bash
# 更新镜像
kubectl set image deployment/task-scheduler \
    task-scheduler=your-registry/task-scheduler:v1.1.0 \
    -n task-scheduler

# 查看升级状态
kubectl rollout status deployment/task-scheduler -n task-scheduler

# 暂停升级
kubectl rollout pause deployment/task-scheduler -n task-scheduler

# 恢复升级
kubectl rollout resume deployment/task-scheduler -n task-scheduler
```

### 回滚

#### Docker回滚

```bash
# 停止服务
docker-compose down

# 使用旧版本镜像
docker-compose up -d task-scheduler:v1.0.0

# 恢复配置
cp config.yaml.backup config.yaml

# 恢复数据（如需要）
mysql -u scheduler -p task_scheduler < backup.sql
```

#### Kubernetes回滚

```bash
# 查看历史版本
kubectl rollout history deployment/task-scheduler -n task-scheduler

# 回滚到上一个版本
kubectl rollout undo deployment/task-scheduler -n task-scheduler

# 回滚到特定版本
kubectl rollout undo deployment/task-scheduler --to-revision=2 -n task-scheduler
```

## 容量规划

### 估算资源需求

#### 计算公式

```
所需节点数 = (峰值任务创建速率 × 平均任务执行时间) / (单节点工作线程数 × 60)

示例:
- 峰值: 1000任务/分钟
- 平均执行时间: 30秒
- 工作线程: 20

所需节点数 = (1000 × 30) / (20 × 60) = 25节点
```

#### 资源配置建议

| 负载级别 | 任务/分钟 | CPU | 内存 | 节点数 |
|---------|----------|-----|------|--------|
| 低 | < 100 | 2核 | 2GB | 1-2 |
| 中 | 100-1000 | 4核 | 4GB | 2-5 |
| 高 | 1000-10000 | 8核 | 8GB | 5-20 |
| 超高 | > 10000 | 16核 | 16GB | 20+ |

### 扩容策略

#### 水平扩容

```bash
# Docker Compose扩容
docker-compose up -d --scale task-scheduler=5

# Kubernetes扩容
kubectl scale deployment/task-scheduler --replicas=5 -n task-scheduler
```

#### 垂直扩容

```yaml
# 增加资源限制
resources:
  limits:
    cpu: "2"
    memory: "4Gi"
  requests:
    cpu: "1"
    memory: "2Gi"
```

## 最佳实践

1. **定期备份**: 每天自动备份数据库和配置
2. **监控告警**: 配置关键指标告警
3. **日志管理**: 使用日志轮转，保留适当时间
4. **安全更新**: 定期更新依赖和系统补丁
5. **性能测试**: 定期进行压力测试
6. **文档更新**: 保持运维文档更新
7. **演练**: 定期进行故障演练和恢复演练
8. **容量规划**: 提前规划资源需求

## 参考资源

- [Prometheus最佳实践](https://prometheus.io/docs/practices/)
- [MySQL性能优化](https://dev.mysql.com/doc/refman/8.0/en/optimization.html)
- [Redis最佳实践](https://redis.io/docs/management/optimization/)
- [Docker生产环境最佳实践](https://docs.docker.com/config/containers/resource_constraints/)
- [Kubernetes生产环境最佳实践](https://kubernetes.io/docs/setup/best-practices/)


# Docker部署指南

本文档说明如何使用Docker和Docker Compose部署Task Scheduler System。

## 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少2GB可用内存
- 至少10GB可用磁盘空间

## 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/yourusername/task-scheduler.git
cd task-scheduler
```

### 2. 配置环境

复制并编辑配置文件：

```bash
cp config.yaml.example config.dev.yaml
```

根据需要修改配置，特别是数据库密码等敏感信息。

### 3. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 4. 验证部署

```bash
# 检查健康状态
curl http://localhost:8080/health

# 访问管理后台
open http://localhost:3000

# 访问Prometheus
open http://localhost:9090

# 访问Grafana
open http://localhost:3001
```

默认登录凭据：
- Grafana: admin / admin

## 服务说明

Docker Compose配置包含以下服务：

### 核心服务

#### task-scheduler-1 和 task-scheduler-2

- **说明**: 任务调度器后端服务（2个节点）
- **端口**: 
  - Node 1: 8080 (HTTP), 9090 (gRPC), 9091 (Metrics)
  - Node 2: 8081 (HTTP), 9092 (gRPC), 9093 (Metrics)
- **依赖**: MySQL, Redis, Consul

#### admin-ui

- **说明**: Vue管理后台
- **端口**: 3000
- **依赖**: task-scheduler-1

### 基础设施服务

#### mysql

- **说明**: MySQL数据库
- **端口**: 3306
- **数据卷**: mysql_data
- **初始化脚本**: scripts/init-mysql.sql

#### redis

- **说明**: Redis缓存和分布式锁
- **端口**: 6379
- **数据卷**: redis_data

#### consul

- **说明**: Consul服务发现
- **端口**: 8500 (HTTP), 8600 (DNS)
- **数据卷**: consul_data
- **UI**: http://localhost:8500

### 监控服务

#### prometheus

- **说明**: Prometheus指标收集
- **端口**: 9090
- **配置**: prometheus.yml
- **数据卷**: prometheus_data

#### grafana

- **说明**: Grafana可视化
- **端口**: 3001
- **数据卷**: grafana_data
- **默认密码**: admin / admin

## 配置说明

### 环境变量

可以通过环境变量覆盖配置：

```bash
# 设置环境变量
export DB_PASSWORD=your-secure-password
export REDIS_PASSWORD=your-redis-password

# 启动服务
docker-compose up -d
```

### 自定义配置

编辑 `config.dev.yaml` 文件：

```yaml
database:
  password: ${DB_PASSWORD}

lock:
  redis:
    password: ${REDIS_PASSWORD}
```

### 数据持久化

Docker Compose使用命名卷持久化数据：

- `mysql_data`: MySQL数据
- `redis_data`: Redis数据
- `consul_data`: Consul数据
- `prometheus_data`: Prometheus数据
- `grafana_data`: Grafana数据

查看卷：

```bash
docker volume ls | grep task-scheduler
```

## 常用操作

### 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 启动特定服务
docker-compose up -d task-scheduler-1 mysql redis
```

### 停止服务

```bash
# 停止所有服务
docker-compose stop

# 停止特定服务
docker-compose stop task-scheduler-1
```

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart task-scheduler-1
```

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f task-scheduler-1

# 查看最近100行日志
docker-compose logs --tail=100 task-scheduler-1
```

### 进入容器

```bash
# 进入后端容器
docker-compose exec task-scheduler-1 sh

# 进入MySQL容器
docker-compose exec mysql bash

# 进入Redis容器
docker-compose exec redis sh
```

### 清理资源

```bash
# 停止并删除容器
docker-compose down

# 停止并删除容器和卷（会删除所有数据）
docker-compose down -v

# 停止并删除容器、卷和镜像
docker-compose down -v --rmi all
```

## 扩缩容

### 扩容后端服务

```bash
# 扩容到3个实例
docker-compose up -d --scale task-scheduler-1=3

# 注意：需要配置负载均衡器
```

### 单节点部署

如果只需要单节点，可以注释掉 `docker-compose.yml` 中的 `task-scheduler-2` 服务。

## 数据库管理

### 连接数据库

```bash
# 使用docker-compose exec
docker-compose exec mysql mysql -u scheduler -p task_scheduler

# 使用MySQL客户端
mysql -h localhost -P 3306 -u scheduler -p task_scheduler
```

### 备份数据库

```bash
# 备份到文件
docker-compose exec mysql mysqldump -u scheduler -p task_scheduler > backup.sql

# 使用root用户备份
docker-compose exec mysql mysqldump -u root -p task_scheduler > backup.sql
```

### 恢复数据库

```bash
# 从备份恢复
docker-compose exec -T mysql mysql -u scheduler -p task_scheduler < backup.sql
```

### 初始化数据库

数据库会在首次启动时自动初始化。如需重新初始化：

```bash
# 停止服务
docker-compose stop mysql

# 删除数据卷
docker volume rm task-scheduler_mysql_data

# 重新启动
docker-compose up -d mysql
```

## Redis管理

### 连接Redis

```bash
# 使用docker-compose exec
docker-compose exec redis redis-cli

# 如果设置了密码
docker-compose exec redis redis-cli -a your-password
```

### 查看Redis信息

```bash
# 查看信息
docker-compose exec redis redis-cli info

# 查看所有键
docker-compose exec redis redis-cli keys '*'

# 清空数据库
docker-compose exec redis redis-cli flushdb
```

## 监控和调试

### 查看资源使用

```bash
# 查看容器资源使用
docker stats

# 查看特定容器
docker stats task-scheduler-node-1
```

### 查看网络

```bash
# 查看网络
docker network ls

# 查看网络详情
docker network inspect task-scheduler_task-scheduler-network
```

### 健康检查

```bash
# 查看容器健康状态
docker-compose ps

# 查看健康检查日志
docker inspect --format='{{json .State.Health}}' task-scheduler-node-1 | jq
```

## 生产环境部署

### 安全加固

1. **修改默认密码**

编辑 `docker-compose.yml`：

```yaml
environment:
  MYSQL_ROOT_PASSWORD: your-strong-password
  MYSQL_PASSWORD: your-strong-password
  REDIS_PASSWORD: your-strong-password
```

2. **使用密钥文件**

```bash
# 创建密钥文件
echo "your-strong-password" > secrets/db_password.txt
echo "your-strong-password" > secrets/redis_password.txt

# 修改权限
chmod 600 secrets/*.txt
```

在 `docker-compose.yml` 中使用：

```yaml
secrets:
  db_password:
    file: ./secrets/db_password.txt
  redis_password:
    file: ./secrets/redis_password.txt

services:
  mysql:
    secrets:
      - db_password
    environment:
      MYSQL_PASSWORD_FILE: /run/secrets/db_password
```

3. **限制网络访问**

```yaml
services:
  mysql:
    networks:
      - backend
    # 不暴露端口到主机

networks:
  backend:
    internal: true
  frontend:
    internal: false
```

### 资源限制

```yaml
services:
  task-scheduler-1:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

### 日志管理

```yaml
services:
  task-scheduler-1:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### 自动重启

```yaml
services:
  task-scheduler-1:
    restart: unless-stopped
```

### 使用外部服务

如果使用外部托管的数据库或Redis：

```yaml
services:
  task-scheduler-1:
    environment:
      DB_HOST: your-rds-endpoint.amazonaws.com
      REDIS_HOST: your-elasticache-endpoint.amazonaws.com
    # 移除mysql和redis服务
```

## 故障排查

### 容器无法启动

```bash
# 查看容器日志
docker-compose logs task-scheduler-1

# 查看容器详情
docker inspect task-scheduler-node-1

# 检查端口占用
netstat -tulpn | grep 8080
```

### 数据库连接失败

```bash
# 检查MySQL是否运行
docker-compose ps mysql

# 查看MySQL日志
docker-compose logs mysql

# 测试连接
docker-compose exec task-scheduler-1 nc -zv mysql 3306
```

### Redis连接失败

```bash
# 检查Redis是否运行
docker-compose ps redis

# 查看Redis日志
docker-compose logs redis

# 测试连接
docker-compose exec task-scheduler-1 nc -zv redis 6379
```

### 服务发现问题

```bash
# 检查Consul是否运行
docker-compose ps consul

# 查看Consul UI
open http://localhost:8500

# 查看注册的服务
curl http://localhost:8500/v1/catalog/services
```

### 磁盘空间不足

```bash
# 清理未使用的镜像
docker image prune -a

# 清理未使用的卷
docker volume prune

# 清理未使用的网络
docker network prune

# 清理所有未使用的资源
docker system prune -a --volumes
```

## 性能优化

### 调整工作线程数

编辑 `config.dev.yaml`：

```yaml
scheduler:
  worker_pool_size: 20  # 根据CPU核心数调整
```

### 调整数据库连接池

```yaml
database:
  max_open_conns: 50
  max_idle_conns: 10
```

### 使用SSD存储

确保Docker数据目录在SSD上：

```bash
# 查看Docker数据目录
docker info | grep "Docker Root Dir"

# 如需迁移，编辑 /etc/docker/daemon.json
{
  "data-root": "/path/to/ssd/docker"
}
```

## 升级

### 升级镜像

```bash
# 拉取最新镜像
docker-compose pull

# 重新创建容器
docker-compose up -d

# 查看升级状态
docker-compose ps
```

### 回滚

```bash
# 使用特定版本的镜像
docker-compose down
docker-compose up -d task-scheduler:v1.0.0
```

## 备份和恢复

### 完整备份

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="./backups/$(date +%Y%m%d-%H%M%S)"
mkdir -p $BACKUP_DIR

# 备份数据库
docker-compose exec -T mysql mysqldump -u root -p$MYSQL_ROOT_PASSWORD task_scheduler > $BACKUP_DIR/database.sql

# 备份配置
cp config.dev.yaml $BACKUP_DIR/
cp docker-compose.yml $BACKUP_DIR/

# 备份卷
docker run --rm -v task-scheduler_mysql_data:/data -v $(pwd)/$BACKUP_DIR:/backup alpine tar czf /backup/mysql_data.tar.gz -C /data .
docker run --rm -v task-scheduler_redis_data:/data -v $(pwd)/$BACKUP_DIR:/backup alpine tar czf /backup/redis_data.tar.gz -C /data .

echo "Backup completed: $BACKUP_DIR"
```

### 恢复

```bash
#!/bin/bash
# restore.sh

BACKUP_DIR=$1

# 停止服务
docker-compose down

# 恢复卷
docker run --rm -v task-scheduler_mysql_data:/data -v $(pwd)/$BACKUP_DIR:/backup alpine tar xzf /backup/mysql_data.tar.gz -C /data
docker run --rm -v task-scheduler_redis_data:/data -v $(pwd)/$BACKUP_DIR:/backup alpine tar xzf /backup/redis_data.tar.gz -C /data

# 启动服务
docker-compose up -d

echo "Restore completed from: $BACKUP_DIR"
```

## 参考资源

- [Docker官方文档](https://docs.docker.com/)
- [Docker Compose文档](https://docs.docker.com/compose/)
- [Docker最佳实践](https://docs.docker.com/develop/dev-best-practices/)

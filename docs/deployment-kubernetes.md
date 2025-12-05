# Kubernetes部署指南

本文档详细说明如何在Kubernetes集群中部署Task Scheduler System。

## 前置要求

- Kubernetes 1.24+
- kubectl已配置并可访问集群
- 至少3个工作节点（用于高可用部署）
- 存储类（StorageClass）已配置
- (可选) Ingress Controller（如nginx-ingress）
- (可选) cert-manager（用于TLS证书管理）

## 部署步骤

### 1. 创建命名空间

```bash
kubectl apply -f k8s/namespace.yaml
```

验证命名空间创建：

```bash
kubectl get namespace task-scheduler
```

### 2. 配置密钥

编辑 `k8s/secrets.yaml` 文件，修改默认密码：

```yaml
stringData:
  db-password: "your-secure-db-password"
  db-root-password: "your-secure-root-password"
  redis-password: "your-secure-redis-password"
```

**重要**: 在生产环境中，建议使用外部密钥管理系统，如：
- HashiCorp Vault
- AWS Secrets Manager
- Azure Key Vault
- Google Secret Manager

应用密钥配置：

```bash
kubectl apply -f k8s/secrets.yaml
```

### 3. 配置ConfigMap

根据需要修改 `k8s/configmap.yaml` 中的配置，然后应用：

```bash
kubectl apply -f k8s/configmap.yaml
```

### 4. 配置RBAC

为应用创建ServiceAccount和权限：

```bash
kubectl apply -f k8s/rbac.yaml
```

### 5. 部署数据库

#### MySQL

```bash
kubectl apply -f k8s/mysql-deployment.yaml
```

等待MySQL就绪：

```bash
kubectl wait --for=condition=ready pod -l app=mysql -n task-scheduler --timeout=300s
```

验证MySQL运行状态：

```bash
kubectl get pods -n task-scheduler -l app=mysql
kubectl logs -n task-scheduler -l app=mysql
```

#### 使用外部数据库

如果使用外部托管的数据库（如AWS RDS、Azure Database等），修改ConfigMap中的数据库连接信息，跳过MySQL部署步骤。

### 6. 部署Redis

```bash
kubectl apply -f k8s/redis-deployment.yaml
```

等待Redis就绪：

```bash
kubectl wait --for=condition=ready pod -l app=redis -n task-scheduler --timeout=300s
```

### 7. 构建和推送Docker镜像

#### 后端镜像

```bash
# 构建镜像
docker build -t your-registry/task-scheduler:v1.0.0 .

# 推送到镜像仓库
docker push your-registry/task-scheduler:v1.0.0
```

#### 前端镜像

```bash
# 构建镜像
cd web
docker build -t your-registry/task-scheduler-admin-ui:v1.0.0 .

# 推送到镜像仓库
docker push your-registry/task-scheduler-admin-ui:v1.0.0
```

修改部署文件中的镜像地址：

```yaml
# k8s/task-scheduler-deployment.yaml
image: your-registry/task-scheduler:v1.0.0

# k8s/admin-ui-deployment.yaml
image: your-registry/task-scheduler-admin-ui:v1.0.0
```

### 8. 部署应用

```bash
# 部署后端服务
kubectl apply -f k8s/task-scheduler-deployment.yaml

# 部署前端UI
kubectl apply -f k8s/admin-ui-deployment.yaml
```

等待Pod就绪：

```bash
kubectl wait --for=condition=ready pod -l app=task-scheduler -n task-scheduler --timeout=300s
kubectl wait --for=condition=ready pod -l app=admin-ui -n task-scheduler --timeout=300s
```

查看部署状态：

```bash
kubectl get pods -n task-scheduler
kubectl get services -n task-scheduler
```

### 9. 配置Ingress（可选）

如果需要从集群外部访问服务，配置Ingress：

编辑 `k8s/ingress.yaml`，修改域名：

```yaml
spec:
  tls:
  - hosts:
    - task-scheduler.your-domain.com
  rules:
  - host: task-scheduler.your-domain.com
```

应用Ingress配置：

```bash
kubectl apply -f k8s/ingress.yaml
```

### 10. 配置自动扩缩容（可选）

```bash
# 部署HPA
kubectl apply -f k8s/hpa.yaml

# 部署PDB（防止过度驱逐）
kubectl apply -f k8s/pdb.yaml
```

验证HPA状态：

```bash
kubectl get hpa -n task-scheduler
```

## 验证部署

### 检查Pod状态

```bash
kubectl get pods -n task-scheduler
```

所有Pod应该处于 `Running` 状态。

### 检查服务

```bash
kubectl get services -n task-scheduler
```

### 查看日志

```bash
# 查看后端日志
kubectl logs -n task-scheduler -l app=task-scheduler --tail=100

# 查看特定Pod日志
kubectl logs -n task-scheduler <pod-name>
```

### 测试API

#### 通过端口转发测试

```bash
# 转发HTTP端口
kubectl port-forward -n task-scheduler svc/task-scheduler-http 8080:8080

# 在另一个终端测试
curl http://localhost:8080/health
```

#### 通过Ingress测试

```bash
curl https://task-scheduler.your-domain.com/health
```

### 访问管理后台

```bash
# 通过端口转发
kubectl port-forward -n task-scheduler svc/admin-ui 3000:80

# 浏览器访问
open http://localhost:3000
```

## 监控和日志

### 查看指标

```bash
# 转发Prometheus指标端口
kubectl port-forward -n task-scheduler svc/task-scheduler-metrics 9091:9091

# 访问指标
curl http://localhost:9091/metrics
```

### 集成Prometheus

如果集群中已部署Prometheus，添加ServiceMonitor：

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: task-scheduler
  namespace: task-scheduler
spec:
  selector:
    matchLabels:
      app: task-scheduler
  endpoints:
  - port: metrics
    interval: 30s
```

### 日志收集

推荐使用以下方案收集日志：

- **EFK Stack**: Elasticsearch + Fluentd + Kibana
- **PLG Stack**: Promtail + Loki + Grafana
- **Cloud Native**: 使用云厂商的日志服务

## 升级和回滚

### 滚动升级

```bash
# 更新镜像
kubectl set image deployment/task-scheduler \
  task-scheduler=your-registry/task-scheduler:v1.1.0 \
  -n task-scheduler

# 查看升级状态
kubectl rollout status deployment/task-scheduler -n task-scheduler
```

### 回滚

```bash
# 查看历史版本
kubectl rollout history deployment/task-scheduler -n task-scheduler

# 回滚到上一个版本
kubectl rollout undo deployment/task-scheduler -n task-scheduler

# 回滚到特定版本
kubectl rollout undo deployment/task-scheduler --to-revision=2 -n task-scheduler
```

## 扩缩容

### 手动扩缩容

```bash
# 扩容到5个副本
kubectl scale deployment/task-scheduler --replicas=5 -n task-scheduler

# 查看扩容状态
kubectl get pods -n task-scheduler -l app=task-scheduler
```

### 自动扩缩容

HPA会根据CPU和内存使用率自动调整副本数：

```bash
# 查看HPA状态
kubectl get hpa -n task-scheduler

# 查看详细信息
kubectl describe hpa task-scheduler-hpa -n task-scheduler
```

## 备份和恢复

### 数据库备份

#### MySQL备份

```bash
# 创建备份Job
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: mysql-backup
  namespace: task-scheduler
spec:
  template:
    spec:
      containers:
      - name: backup
        image: mysql:8.0
        command:
        - /bin/sh
        - -c
        - |
          mysqldump -h mysql-service -u scheduler -p\$MYSQL_PASSWORD task_scheduler > /backup/backup-\$(date +%Y%m%d-%H%M%S).sql
        env:
        - name: MYSQL_PASSWORD
          valueFrom:
            secretKeyRef:
              name: task-scheduler-secrets
              key: db-password
        volumeMounts:
        - name: backup
          mountPath: /backup
      volumes:
      - name: backup
        persistentVolumeClaim:
          claimName: backup-pvc
      restartPolicy: Never
EOF
```

### 配置备份

```bash
# 备份ConfigMap和Secrets
kubectl get configmap -n task-scheduler -o yaml > backup-configmap.yaml
kubectl get secret -n task-scheduler -o yaml > backup-secrets.yaml
```

## 故障排查

### Pod无法启动

```bash
# 查看Pod事件
kubectl describe pod <pod-name> -n task-scheduler

# 查看Pod日志
kubectl logs <pod-name> -n task-scheduler

# 查看上一次容器日志（如果容器重启）
kubectl logs <pod-name> -n task-scheduler --previous
```

### 服务无法访问

```bash
# 检查Service
kubectl get svc -n task-scheduler
kubectl describe svc task-scheduler-http -n task-scheduler

# 检查Endpoints
kubectl get endpoints -n task-scheduler

# 测试Pod网络
kubectl run -it --rm debug --image=busybox --restart=Never -n task-scheduler -- sh
# 在Pod中测试
wget -O- http://task-scheduler-http:8080/health
```

### 数据库连接问题

```bash
# 检查MySQL Pod
kubectl get pods -n task-scheduler -l app=mysql
kubectl logs -n task-scheduler -l app=mysql

# 测试数据库连接
kubectl run -it --rm mysql-client --image=mysql:8.0 --restart=Never -n task-scheduler -- \
  mysql -h mysql-service -u scheduler -p
```

### 分布式锁问题

```bash
# 检查Redis
kubectl get pods -n task-scheduler -l app=redis
kubectl logs -n task-scheduler -l app=redis

# 测试Redis连接
kubectl run -it --rm redis-client --image=redis:7-alpine --restart=Never -n task-scheduler -- \
  redis-cli -h redis-service ping
```

## 性能优化

### 资源配置

根据实际负载调整资源请求和限制：

```yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

### 数据库优化

- 增加连接池大小
- 配置适当的索引
- 定期清理历史数据

### Redis优化

- 配置持久化策略
- 调整内存限制
- 使用Redis Cluster（大规模部署）

## 安全最佳实践

1. **使用非root用户运行容器**（已在Dockerfile中配置）
2. **启用Pod Security Policy/Pod Security Standards**
3. **配置Network Policy限制Pod间通信**
4. **使用TLS加密通信**
5. **定期更新镜像和依赖**
6. **使用外部密钥管理系统**
7. **启用审计日志**
8. **配置RBAC最小权限原则**

## 高可用配置

### 多可用区部署

```yaml
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchExpressions:
        - key: app
          operator: In
          values:
          - task-scheduler
      topologyKey: topology.kubernetes.io/zone
```

### 数据库高可用

- 使用MySQL主从复制或Group Replication
- 使用云厂商的托管数据库服务
- 配置自动备份和故障转移

### Redis高可用

- 使用Redis Sentinel
- 使用Redis Cluster
- 使用云厂商的托管Redis服务

## 参考资源

- [Kubernetes官方文档](https://kubernetes.io/docs/)
- [Helm Charts](https://helm.sh/)
- [Kubernetes最佳实践](https://kubernetes.io/docs/concepts/configuration/overview/)

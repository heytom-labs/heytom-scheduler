# 数据库模型和 Repo 实现说明

## 📁 已创建的文件

### 1. 数据库模型层 (`internal/data/`)

#### `model.go` - 数据库模型定义
- **Task** - 任务表模型
- **TaskExecution** - 任务执行记录表模型
- **Metadata** - JSON 元数据类型（实现了 `sql.Scanner` 和 `driver.Valuer`）
- 枚举类型：`TaskType`、`TaskStatus`、`ExecutionStatus`

#### `task.go` - 任务仓储实现
实现了 `biz.TaskRepo` 接口的所有方法：
- `CreateTask` - 创建任务
- `GetTask` - 获取任务详情
- `UpdateTask` - 更新任务
- `DeleteTask` - 删除任务
- `ListTasks` - 任务列表查询（支持分页、状态筛选、类型筛选、关键词搜索）
- `UpdateTaskStatus` - 更新任务状态
- `UpdateTaskNextRunTime` - 更新下次执行时间
- `IncrementExecutionCount` - 增加执行次数统计

#### `execution.go` - 执行记录仓储实现
实现了 `biz.ExecutionRepo` 接口的所有方法：
- `CreateExecution` - 创建执行记录
- `GetExecution` - 获取执行记录详情
- `UpdateExecution` - 更新执行记录
- `ListExecutions` - 执行记录列表查询（支持分页、任务ID筛选、状态筛选）
- `UpdateExecutionStatus` - 更新执行状态

#### `data.go` - 数据层初始化（已更新）
- 集成 GORM
- MySQL 数据库连接初始化
- 自动表结构迁移
- Wire 依赖注入配置

### 2. 业务层接口 (`internal/biz/`)

#### `task.go` - 业务模型和仓储接口
- **Task** - 任务业务模型
- **TaskExecution** - 执行记录业务模型
- **TaskListFilter** - 任务列表过滤条件
- **ExecutionListFilter** - 执行记录列表过滤条件
- **TaskRepo** - 任务仓储接口定义
- **ExecutionRepo** - 执行记录仓储接口定义

### 3. 数据库脚本 (`scripts/`)

#### `init_db.sql` - MySQL 建表脚本
- 创建数据库 `heytom_scheduler`
- 创建 `tasks` 表（任务表）
- 创建 `task_executions` 表（执行记录表）
- 包含示例数据

## 📊 数据库表结构

### tasks 表（任务表）
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | BIGINT | 任务ID（主键） |
| name | VARCHAR(255) | 任务名称 |
| description | TEXT | 任务描述 |
| type | VARCHAR(20) | 任务类型（immediate/scheduled/cron/interval） |
| status | VARCHAR(20) | 任务状态（pending/running/paused/completed/failed/cancelled） |
| schedule | VARCHAR(255) | 调度配置 |
| handler | VARCHAR(255) | 处理器名称 |
| payload | TEXT | 任务负载（JSON） |
| timeout | INT | 超时时间（秒） |
| metadata | JSON | 元数据 |
| next_run_time | DATETIME | 下次执行时间 |
| execution_count | BIGINT | 执行次数 |
| success_count | BIGINT | 成功次数 |
| failed_count | BIGINT | 失败次数 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

**索引**：
- 主键：`id`
- 普通索引：`name`, `type`, `status`, `next_run_time`, `created_at`

### task_executions 表（执行记录表）
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | BIGINT | 执行记录ID（主键） |
| task_id | BIGINT | 任务ID |
| task_name | VARCHAR(255) | 任务名称 |
| status | VARCHAR(20) | 执行状态（queued/executing/success/failed/timeout/cancelled） |
| node_id | VARCHAR(100) | 执行节点ID |
| start_time | DATETIME | 开始时间 |
| end_time | DATETIME | 结束时间 |
| duration | INT | 执行耗时（毫秒） |
| result | TEXT | 执行结果 |
| error | TEXT | 错误信息 |
| retry_count | INT | 重试次数 |
| payload | TEXT | 执行负载（JSON） |
| created_at | DATETIME | 创建时间 |

**索引**：
- 主键：`id`
- 普通索引：`task_id`, `status`, `node_id`, `created_at`

## 🚀 使用方法

### 1. 初始化数据库
```bash
mysql -u root -p < scripts/init_db.sql
```

### 2. 配置数据库连接
在 `configs/config.yaml` 中配置数据库连接信息：
```yaml
data:
  database:
    source: "user:password@tcp(127.0.0.1:3306)/heytom_scheduler?charset=utf8mb4&parseTime=True&loc=Local"
```

### 3. 安装依赖
```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
```

### 4. 重新生成 Wire 依赖注入代码
```bash
cd cmd/heytom-scheduler
wire
```

## 🔧 技术栈
- **ORM**: GORM v2
- **数据库**: MySQL 5.7+
- **依赖注入**: Google Wire
- **字符集**: UTF8MB4（支持 Emoji）

## 📝 注意事项
1. 所有时间字段使用 `DATETIME` 类型
2. JSON 字段使用 MySQL 的 `JSON` 类型（需要 MySQL 5.7+）
3. 使用了 GORM 的 `AutoMigrate` 功能，会自动创建/更新表结构
4. Metadata 字段实现了自定义类型，可以直接在 Go 中使用 `map[string]string`
5. 分页查询默认按 `id DESC` 排序
6. 所有 Repo 方法都支持 Context 传递

## 🎯 下一步
1. 实现 Service 层业务逻辑
2. 实现 gRPC/HTTP 服务接口
3. 实现任务调度核心逻辑
4. 添加单元测试

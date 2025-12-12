-- ============================================
-- HeyTom Scheduler 数据库初始化脚本
-- MySQL 5.7+
-- ============================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `heytom_scheduler` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `heytom_scheduler`;

-- ============================================
-- 任务表
-- ============================================
CREATE TABLE IF NOT EXISTS `tasks` (
  `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `name` VARCHAR(255) NOT NULL COMMENT '任务名称',
  `description` TEXT COMMENT '任务描述',
  `type` VARCHAR(20) NOT NULL COMMENT '任务类型: IMMEDIATE(立即执行), SCHEDULED(指定时间), CRON(Cron表达式), INTERVAL(固定间隔)',
  `status` VARCHAR(20) NOT NULL DEFAULT 'PENDING' COMMENT '任务状态: PENDING(等待中), RUNNING(运行中), PAUSED(已暂停), COMPLETED(已完成), FAILED(失败), CANCELLED(已取消)',
  `schedule` VARCHAR(255) DEFAULT NULL COMMENT '调度配置: Cron表达式、时间戳或间隔秒数',
  `handler` VARCHAR(255) NOT NULL COMMENT '处理器名称',
  `payload` TEXT COMMENT '任务负载(JSON格式)',
  `timeout` INT(11) DEFAULT 300 COMMENT '超时时间(秒)',
  `metadata` JSON COMMENT '元数据',
  `next_run_time` DATETIME DEFAULT NULL COMMENT '下次执行时间',
  `execution_count` BIGINT(20) DEFAULT 0 COMMENT '执行次数',
  `success_count` BIGINT(20) DEFAULT 0 COMMENT '成功次数',
  `failed_count` BIGINT(20) DEFAULT 0 COMMENT '失败次数',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`),
  KEY `idx_next_run_time` (`next_run_time`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务表';

-- ============================================
-- 任务执行记录表
-- ============================================
CREATE TABLE IF NOT EXISTS `task_executions` (
  `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '执行记录ID',
  `task_id` BIGINT(20) NOT NULL COMMENT '任务ID',
  `task_name` VARCHAR(255) NOT NULL COMMENT '任务名称',
  `status` VARCHAR(20) NOT NULL COMMENT '执行状态: QUEUED(队列中), EXECUTING(执行中), SUCCESS(成功), EXECUTION_FAILED(失败), TIMEOUT(超时), EXECUTION_CANCELLED(已取消)',
  `node_id` VARCHAR(100) DEFAULT NULL COMMENT '执行节点ID',
  `start_time` DATETIME DEFAULT NULL COMMENT '开始时间',
  `end_time` DATETIME DEFAULT NULL COMMENT '结束时间',
  `duration` INT(11) DEFAULT NULL COMMENT '执行耗时(毫秒)',
  `result` TEXT COMMENT '执行结果',
  `error` TEXT COMMENT '错误信息',
  `retry_count` INT(11) DEFAULT 0 COMMENT '重试次数',
  `payload` TEXT COMMENT '执行负载(JSON格式)',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_status` (`status`),
  KEY `idx_node_id` (`node_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务执行记录表';

-- ============================================
-- 示例数据（可选）
-- ============================================

-- 插入示例任务
INSERT INTO `tasks` (`name`, `description`, `type`, `status`, `schedule`, `handler`, `payload`, `timeout`, `metadata`) VALUES
('示例立即执行任务', '这是一个立即执行的示例任务', 'IMMEDIATE', 'PENDING', '', 'example_handler', '{"key":"value"}', 300, '{"category":"example"}'),
('示例Cron任务', '每5分钟执行一次', 'CRON', 'PENDING', '0 */5 * * * *', 'cron_handler', '{"type":"cron"}', 600, '{"category":"scheduled"}'),
('示例定时任务', '指定时间执行', 'SCHEDULED', 'PENDING', '2025-12-31T23:59:59Z', 'scheduled_handler', '{"type":"scheduled"}', 300, '{"category":"scheduled"}');

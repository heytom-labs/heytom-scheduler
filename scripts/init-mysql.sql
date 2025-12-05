-- Initialize MySQL database for Task Scheduler

-- Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id VARCHAR(36),
    execution_mode VARCHAR(20) NOT NULL,
    schedule_config JSON,
    callback_config JSON,
    retry_policy JSON,
    concurrency_limit INT DEFAULT 0,
    alert_policy JSON,
    status VARCHAR(20) NOT NULL,
    retry_count INT DEFAULT 0,
    node_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    metadata JSON,
    INDEX idx_parent_id (parent_id),
    INDEX idx_status (status),
    INDEX idx_node_id (node_id),
    INDEX idx_execution_mode (execution_mode),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create execution history table
CREATE TABLE IF NOT EXISTS execution_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) NOT NULL,
    output TEXT,
    error TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_task_id (task_id),
    INDEX idx_start_time (start_time),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create nodes table for tracking scheduler nodes
CREATE TABLE IF NOT EXISTS scheduler_nodes (
    node_id VARCHAR(100) PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL,
    last_heartbeat TIMESTAMP NOT NULL,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status),
    INDEX idx_last_heartbeat (last_heartbeat)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create alerts table for tracking alert notifications
CREATE TABLE IF NOT EXISTS alerts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_id VARCHAR(36) NOT NULL,
    alert_type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    channels JSON,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL,
    metadata JSON,
    INDEX idx_task_id (task_id),
    INDEX idx_alert_type (alert_type),
    INDEX idx_sent_at (sent_at),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample data for testing (optional)
-- Uncomment the following lines if you want sample data

-- INSERT INTO tasks (id, name, description, execution_mode, status) VALUES
-- ('sample-task-1', 'Sample Immediate Task', 'A sample task for testing', 'immediate', 'pending'),
-- ('sample-task-2', 'Sample Scheduled Task', 'A scheduled task for testing', 'scheduled', 'pending');

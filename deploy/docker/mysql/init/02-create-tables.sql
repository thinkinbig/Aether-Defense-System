-- Database schema for Aether Defense System
-- This script creates all required tables for the system

USE aether_defense;

-- ============================================
-- Trade Domain Tables
-- ============================================

-- Order main table
CREATE TABLE IF NOT EXISTS `trade_order` (
  `id` BIGINT NOT NULL COMMENT 'Primary key, Snowflake algorithm ID',
  `user_id` BIGINT NOT NULL COMMENT 'User ID, sharding key',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT 'Status: 1=Pending Payment, 2=Closed, 3=Paid, 4=Finished, 5=Refunded',
  `total_amount` INT NOT NULL COMMENT 'Total order amount in cents',
  `pay_amount` INT NOT NULL COMMENT 'Actual payment amount in cents',
  `pay_channel` TINYINT DEFAULT NULL COMMENT 'Payment channel: 1=Alipay, 2=WeChat',
  `out_trade_no` VARCHAR(64) DEFAULT NULL COMMENT 'Third-party payment transaction number',
  `pay_time` DATETIME DEFAULT NULL COMMENT 'Payment success time',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `version` INT NOT NULL DEFAULT 0 COMMENT 'Optimistic lock version number',
  PRIMARY KEY (`id`),
  KEY `idx_user_status` (`user_id`, `status`) COMMENT 'Client-side query index',
  KEY `idx_out_trade_no` (`out_trade_no`) COMMENT 'Payment callback idempotency index'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Order main table';

-- Order items table
CREATE TABLE IF NOT EXISTS `trade_order_item` (
  `id` BIGINT NOT NULL COMMENT 'Primary key, Snowflake algorithm ID',
  `order_id` BIGINT NOT NULL COMMENT 'Order ID',
  `user_id` BIGINT NOT NULL COMMENT 'Redundant field for sharding binding',
  `course_id` BIGINT NOT NULL COMMENT 'Course ID',
  `course_name` VARCHAR(128) NOT NULL COMMENT 'Snapshot: course name at purchase time',
  `price` INT NOT NULL COMMENT 'Snapshot: unit price at purchase time in cents',
  `real_pay_amount` INT NOT NULL COMMENT 'Actual payment allocation amount in cents',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_order` (`order_id`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Order items table';

-- ============================================
-- Promotion Domain Tables
-- ============================================

-- Coupon record table
CREATE TABLE IF NOT EXISTS `promotion_coupon_record` (
  `id` BIGINT NOT NULL COMMENT 'Primary key, Snowflake algorithm ID',
  `user_id` BIGINT NOT NULL COMMENT 'User ID, sharding key',
  `template_id` BIGINT NOT NULL COMMENT 'Coupon template ID',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1=Unused, 2=Used, 3=Expired',
  `use_time` DATETIME DEFAULT NULL COMMENT 'Usage time',
  `order_id` BIGINT DEFAULT NULL COMMENT 'Associated order ID',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_user_template` (`user_id`, `template_id`) COMMENT 'Limit: one coupon per user per template (optional)',
  KEY `idx_user_status` (`user_id`, `status`),
  KEY `idx_order` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Coupon record table';

-- ============================================
-- User Domain Tables
-- ============================================

-- User table
CREATE TABLE IF NOT EXISTS `user` (
  `id` BIGINT NOT NULL COMMENT 'Primary key, Snowflake algorithm ID',
  `username` VARCHAR(64) NOT NULL COMMENT 'Username',
  `mobile` VARCHAR(20) NOT NULL COMMENT 'Mobile phone number',
  `email` VARCHAR(128) DEFAULT NULL COMMENT 'Email address',
  `avatar` VARCHAR(256) DEFAULT NULL COMMENT 'Avatar URL',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT 'Status: 1=Normal, 2=Banned',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_mobile` (`mobile`),
  UNIQUE KEY `uniq_username` (`username`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='User table';

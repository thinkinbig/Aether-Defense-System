-- Database schema for Aether Defense System
-- This script creates all required tables for the system

USE aether_defense;

-- ============================================
-- Trade Domain Tables
-- ============================================

-- Order main table
CREATE TABLE IF NOT EXISTS `trade_order` (
  `id` BIGINT NOT NULL COMMENT '主键，雪花算法ID',
  `user_id` BIGINT NOT NULL COMMENT '用户ID，分片键',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待支付, 2关闭, 3已支付, 4完成, 5退款',
  `total_amount` INT NOT NULL COMMENT '订单总金额，单位：分',
  `pay_amount` INT NOT NULL COMMENT '实付金额，单位：分',
  `pay_channel` TINYINT DEFAULT NULL COMMENT '支付渠道：1支付宝, 2微信',
  `out_trade_no` VARCHAR(64) DEFAULT NULL COMMENT '第三方支付流水号',
  `pay_time` DATETIME DEFAULT NULL COMMENT '支付成功时间',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `version` INT NOT NULL DEFAULT 0 COMMENT '乐观锁版本号',
  PRIMARY KEY (`id`),
  KEY `idx_user_status` (`user_id`, `status`) COMMENT 'C端查询索引',
  KEY `idx_out_trade_no` (`out_trade_no`) COMMENT '支付回调幂等索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='订单主表';

-- Order items table
CREATE TABLE IF NOT EXISTS `trade_order_item` (
  `id` BIGINT NOT NULL COMMENT '主键，雪花算法ID',
  `order_id` BIGINT NOT NULL COMMENT '订单ID',
  `user_id` BIGINT NOT NULL COMMENT '冗余字段，用于绑定分片',
  `course_id` BIGINT NOT NULL COMMENT '课程ID',
  `course_name` VARCHAR(128) NOT NULL COMMENT '快照：课程名称',
  `price` INT NOT NULL COMMENT '快照：购买时的单价，单位：分',
  `real_pay_amount` INT NOT NULL COMMENT '实付分摊金额，单位：分',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_order` (`order_id`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='订单明细表';

-- ============================================
-- Promotion Domain Tables
-- ============================================

-- Coupon record table
CREATE TABLE IF NOT EXISTS `promotion_coupon_record` (
  `id` BIGINT NOT NULL COMMENT '主键，雪花算法ID',
  `user_id` BIGINT NOT NULL COMMENT '用户ID，分片键',
  `template_id` BIGINT NOT NULL COMMENT '优惠券模板ID',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1未使用, 2已使用, 3已过期',
  `use_time` DATETIME DEFAULT NULL COMMENT '使用时间',
  `order_id` BIGINT DEFAULT NULL COMMENT '关联订单ID',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_user_template` (`user_id`, `template_id`) COMMENT '限制每个用户每种券只能领一张（可选）',
  KEY `idx_user_status` (`user_id`, `status`),
  KEY `idx_order` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='优惠券领取表';

-- ============================================
-- User Domain Tables
-- ============================================

-- User table
CREATE TABLE IF NOT EXISTS `user` (
  `id` BIGINT NOT NULL COMMENT '主键，雪花算法ID',
  `username` VARCHAR(64) NOT NULL COMMENT '用户名',
  `mobile` VARCHAR(20) NOT NULL COMMENT '手机号',
  `email` VARCHAR(128) DEFAULT NULL COMMENT '邮箱',
  `avatar` VARCHAR(256) DEFAULT NULL COMMENT '头像URL',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1正常, 2禁用',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_mobile` (`mobile`),
  UNIQUE KEY `uniq_username` (`username`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户表';

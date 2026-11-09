-- 破局 Breakout 任务管理系统数据库初始化脚本
-- 数据库版本：MySQL 8.0+
-- 创建时间：2026-11-09

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `team_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '所属团队ID',
  `nickname` varchar(50) NOT NULL DEFAULT '' COMMENT '昵称',
  `avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '头像URL',
  `phone` varchar(20) NOT NULL DEFAULT '' COMMENT '手机号',
  `email` varchar(100) NOT NULL DEFAULT '' COMMENT '邮箱',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '密码（bcrypt加密）',
  `role` tinyint unsigned NOT NULL DEFAULT '3' COMMENT '角色：1-所有者 2-管理员 3-成员 4-访客',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态：0-禁用 1-正常',
  `wx_openid` varchar(100) DEFAULT NULL COMMENT '微信openid',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_phone` (`phone`),
  UNIQUE KEY `uk_wx_openid` (`wx_openid`),
  KEY `idx_team_id` (`team_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ----------------------------
-- Table structure for team
-- ----------------------------
DROP TABLE IF EXISTS `team`;
CREATE TABLE `team` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '团队ID',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '团队名称',
  `owner_id` bigint unsigned NOT NULL COMMENT '所有者ID',
  `description` varchar(500) NOT NULL DEFAULT '' COMMENT '团队描述',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态：0-禁用 1-正常',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_owner_id` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队表';

-- ----------------------------
-- Table structure for project
-- ----------------------------
DROP TABLE IF EXISTS `project`;
CREATE TABLE `project` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '项目ID',
  `team_id` bigint unsigned NOT NULL COMMENT '所属团队ID',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '项目名称',
  `description` varchar(1000) NOT NULL DEFAULT '' COMMENT '项目描述',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态：0-归档 1-进行中 2-暂停',
  `start_time` datetime DEFAULT NULL COMMENT '开始时间',
  `end_time` datetime DEFAULT NULL COMMENT '结束时间',
  `owner_id` bigint unsigned NOT NULL COMMENT '项目负责人ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_team_id` (`team_id`),
  KEY `idx_owner_id` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目表';

-- ----------------------------
-- Table structure for project_member
-- ----------------------------
DROP TABLE IF EXISTS `project_member`;
CREATE TABLE `project_member` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `project_id` bigint unsigned NOT NULL COMMENT '项目ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `role` tinyint unsigned NOT NULL DEFAULT '2' COMMENT '角色：1-管理员 2-成员',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_user` (`project_id`, `user_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目成员表';

-- ----------------------------
-- Table structure for task
-- ----------------------------
DROP TABLE IF EXISTS `task`;
CREATE TABLE `task` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `team_id` bigint unsigned NOT NULL COMMENT '所属团队ID',
  `project_id` bigint unsigned NOT NULL COMMENT '所属项目ID',
  `parent_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '父任务ID，0表示根任务',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '任务标题',
  `description` text COMMENT '任务描述',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态：1-待办 2-进行中 3-待复核 4-待验收 5-已完成 6-已关闭',
  `priority` tinyint unsigned NOT NULL DEFAULT '2' COMMENT '优先级：1-高 2-中 3-低',
  `assignee_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '负责人ID',
  `reviewer_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '复核人ID',
  `acceptor_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '验收人ID',
  `creator_id` bigint unsigned NOT NULL COMMENT '创建人ID',
  `due_date` datetime DEFAULT NULL COMMENT '截止时间',
  `start_date` datetime DEFAULT NULL COMMENT '开始时间',
  `estimate` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '预估工时（小时）',
  `actual` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '实际工时（小时）',
  `blocked_reason` varchar(500) NOT NULL DEFAULT '' COMMENT '阻塞原因',
  `blocked_since` datetime DEFAULT NULL COMMENT '阻塞开始时间',
  `closed_evidence` text COMMENT '关闭证据',
  `sort_order` int unsigned NOT NULL DEFAULT '0' COMMENT '排序权重',
  `is_milestone` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否为里程碑：0-否 1-是',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_assignee_id` (`assignee_id`),
  KEY `idx_status` (`status`),
  KEY `idx_due_date` (`due_date`),
  KEY `idx_blocked_since` (`blocked_since`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务表';

-- ----------------------------
-- Table structure for task_relation
-- ----------------------------
DROP TABLE IF EXISTS `task_relation`;
CREATE TABLE `task_relation` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '关系ID',
  `predecessor_task_id` bigint unsigned NOT NULL COMMENT '前置任务ID',
  `successor_task_id` bigint unsigned NOT NULL COMMENT '后置任务ID',
  `type` char(2) NOT NULL DEFAULT 'FS' COMMENT '依赖类型：FS/SS/FF/SF',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_predecessor_successor` (`predecessor_task_id`, `successor_task_id`),
  KEY `idx_successor_task_id` (`successor_task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务依赖关系表';

-- ----------------------------
-- Table structure for comment
-- ----------------------------
DROP TABLE IF EXISTS `comment`;
CREATE TABLE `comment` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '评论ID',
  `task_id` bigint unsigned NOT NULL COMMENT '任务ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `parent_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '父评论ID，0表示根评论',
  `content` text NOT NULL COMMENT '评论内容',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';

-- ----------------------------
-- Table structure for attachment
-- ----------------------------
DROP TABLE IF EXISTS `attachment`;
CREATE TABLE `attachment` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '附件ID',
  `task_id` bigint unsigned NOT NULL COMMENT '任务ID',
  `user_id` bigint unsigned NOT NULL COMMENT '上传用户ID',
  `file_path` varchar(255) NOT NULL DEFAULT '' COMMENT '文件路径',
  `file_name` varchar(255) NOT NULL DEFAULT '' COMMENT '文件名',
  `file_size` bigint unsigned NOT NULL DEFAULT '0' COMMENT '文件大小（字节）',
  `file_type` varchar(50) NOT NULL DEFAULT '' COMMENT '文件类型',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='附件表';

-- ----------------------------
-- Table structure for notification
-- ----------------------------
DROP TABLE IF EXISTS `notification`;
CREATE TABLE `notification` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '通知ID',
  `user_id` bigint unsigned NOT NULL COMMENT '接收用户ID',
  `task_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '关联任务ID',
  `type` tinyint unsigned NOT NULL COMMENT '通知类型：1-任务分配 2-任务逾期 3-任务阻塞 4-依赖变更 5-复核请求 6-验收请求 7-周报生成 8-负载过高',
  `content` varchar(500) NOT NULL DEFAULT '' COMMENT '通知内容',
  `is_read` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否已读：0-未读 1-已读',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_is_read` (`is_read`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知表';

-- ----------------------------
-- Table structure for weekly_report
-- ----------------------------
DROP TABLE IF EXISTS `weekly_report`;
CREATE TABLE `weekly_report` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '周报ID',
  `team_id` bigint unsigned NOT NULL COMMENT '团队ID',
  `project_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '项目ID，0表示全团队',
  `start_date` varchar(20) NOT NULL COMMENT '开始日期',
  `end_date` varchar(20) NOT NULL COMMENT '结束日期',
  `content` longtext NOT NULL COMMENT '周报内容（JSON格式）',
  `generated_by` bigint unsigned NOT NULL DEFAULT '0' COMMENT '生成人ID，0表示系统自动生成',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_team_id` (`team_id`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_start_date` (`start_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周报表';

-- ----------------------------
-- Table structure for ai_action
-- ----------------------------
DROP TABLE IF EXISTS `ai_action`;
CREATE TABLE `ai_action` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'AI操作ID',
  `task_id` bigint unsigned NOT NULL COMMENT '关联任务ID',
  `user_id` bigint unsigned NOT NULL COMMENT '操作用户ID',
  `action_type` tinyint unsigned NOT NULL COMMENT '操作类型：1-任务拆解 2-周报生成 3-风险评估',
  `prompt` text NOT NULL COMMENT '请求提示词',
  `response` longtext NOT NULL COMMENT 'AI返回结果',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态：0-失败 1-成功',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action_type` (`action_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI操作记录表';

SET FOREIGN_KEY_CHECKS = 1;
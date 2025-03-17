CREATE TABLE `user_sessions` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `device_id` varchar(64) NOT NULL COMMENT '设备ID',
  `token` varchar(255) NOT NULL COMMENT '会话token',
  `device_type` tinyint(1) NOT NULL COMMENT '设备类型：1-Web，2-Android，3-iOS，4-PC',
  `device_info` varchar(255) DEFAULT NULL COMMENT '设备信息',
  `ip` varchar(64) DEFAULT NULL COMMENT 'IP地址',
  `is_online` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否在线：0-离线，1-在线',
  `last_active_time` datetime NOT NULL COMMENT '最后活跃时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_device` (`user_id`,`device_id`),
  KEY `idx_token` (`token`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户会话表';
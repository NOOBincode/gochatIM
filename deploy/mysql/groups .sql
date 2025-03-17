CREATE TABLE `groups` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '群组ID',
  `name` varchar(100) NOT NULL COMMENT '群组名称',
  `avatar` varchar(255) DEFAULT NULL COMMENT '群头像URL',
  `description` varchar(500) DEFAULT NULL COMMENT '群描述',
  `owner_id` bigint(20) NOT NULL COMMENT '群主ID',
  `max_members` int(11) DEFAULT '200' COMMENT '最大成员数',
  `status` tinyint(1) DEFAULT '1' COMMENT '状态: 0-解散, 1-正常',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_owner_id` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群组表';
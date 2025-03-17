CREATE TABLE `user_conversations` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `conversation_id` varchar(64) NOT NULL COMMENT '会话ID',
  `unread_count` int(11) NOT NULL DEFAULT '0' COMMENT '未读消息数',
  `last_ack_message_id` varchar(50) DEFAULT NULL COMMENT '最后确认的消息ID',
  `is_top` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否置顶：0-否，1-是',
  `is_mute` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否免打扰：0-否，1-是',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态：0-已删除，1-正常',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_conversation` (`user_id`,`conversation_id`),
  KEY `idx_conversation_id` (`conversation_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户会话关系表';
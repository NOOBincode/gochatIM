CREATE TABLE `conversations` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `conversation_id` varchar(64) NOT NULL COMMENT '会话ID',
  `type` tinyint(4) NOT NULL COMMENT '会话类型：1-单聊，2-群聊',
  `creator_id` bigint(20) unsigned NOT NULL COMMENT '创建者ID',
  `last_message_id` varchar(50) DEFAULT NULL COMMENT '最后一条消息ID',
  `last_message_content` varchar(255) DEFAULT NULL COMMENT '最后一条消息内容',
  `last_message_time` datetime DEFAULT NULL COMMENT '最后一条消息时间',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态：1-正常，2-已删除',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_conversation_id` (`conversation_id`),
  KEY `idx_creator_id` (`creator_id`),
  KEY `idx_last_message_time` (`last_message_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会话表';
1.负责结构化存储消息
2.读取消息
3.存储客户端分配的信息
4.读取客户端分配的信息
5.激活客户端

定义一套通用存储接口


CREATE TABLE `message` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `msg_uid` int(11) NOT NULL DEFAULT '0' COMMENT '消息用户id',
  `msg_mid` int(11) NOT NULL DEFAULT '0' COMMENT '消息id',
  `msg_content` varchar(255) NOT NULL DEFAULT '' COMMENT '消息内容',
  `msg_type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '消息类型',
  `msg_time` int(11) NOT NULL DEFAULT '0' COMMENT '消息时间',
  `msg_from` int(11) NOT NULL DEFAULT '0' COMMENT '消息发送者',
  `msg_to` int(11) NOT NULL DEFAULT '0' COMMENT '消息接受者',
  `msg_group` int(11) NOT NULL DEFAULT '0' COMMENT '消息组',
  PRIMARY KEY (`id`),
  KEY `idx_msg_uid` (`msg_uid`),
  KEY `idx_msg_mid` (`msg_mid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8

CREATE TABLE `user_group` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL DEFAULT '0' COMMENT '用户',
  `group_id` int(11) NOT NULL DEFAULT '0' COMMENT '组',
  PRIMARY KEY (`id`),
  KEY `idx_user` (`user_id`),
  KEY `idx_group` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8
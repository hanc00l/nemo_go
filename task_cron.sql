/*
 Navicat Premium Data Transfer

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 50738
 Source Host           : localhost:3306
 Source Schema         : nemo

 Target Server Type    : MySQL
 Target Server Version : 50738
 File Encoding         : 65001

 Date: 30/05/2022 10:54:01
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for task_cron
-- ----------------------------
DROP TABLE IF EXISTS `task_cron`;
CREATE TABLE `task_cron` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `task_id` char(36) NOT NULL,
  `task_name` varchar(100) NOT NULL,
  `args` varchar(2000) DEFAULT NULL,
  `kwargs` varchar(8000) DEFAULT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  `cron_rule` varchar(200) NOT NULL COMMENT '定时规则',
  `lastrun_datetime` datetime DEFAULT NULL COMMENT '上次运行时间',
  `status` varchar(10) NOT NULL COMMENT '状态enable or disable',
  `run_count` int(10) DEFAULT NULL COMMENT '启动次数',
  `comment` varchar(200) DEFAULT NULL COMMENT '定时任务说明',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3166 DEFAULT CHARSET=utf8;

SET FOREIGN_KEY_CHECKS = 1;

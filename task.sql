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

 Date: 29/05/2022 21:07:47
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for task
-- ----------------------------
DROP TABLE IF EXISTS `task`;
CREATE TABLE `task` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `task_id` char(36) NOT NULL,
  `task_name` varchar(100) NOT NULL,
  `args` varchar(2000) DEFAULT NULL,
  `kwargs` varchar(4000) DEFAULT NULL,
  `worker` varchar(100) DEFAULT NULL,
  `state` varchar(40) NOT NULL,
  `result` varchar(4000) DEFAULT NULL,
  `received` datetime DEFAULT NULL,
  `retried` datetime DEFAULT NULL,
  `revoked` datetime DEFAULT NULL,
  `started` datetime DEFAULT NULL,
  `succeeded` datetime DEFAULT NULL,
  `failed` datetime DEFAULT NULL,
  `progress_message` varchar(100) DEFAULT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  `cron_id` char(36) DEFAULT NULL COMMENT 'the id for cron task',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3212 DEFAULT CHARSET=utf8;

SET FOREIGN_KEY_CHECKS = 1;

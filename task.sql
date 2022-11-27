/*
 Navicat Premium Data Transfer

 Source Server         : nemo
 Source Server Type    : MySQL
 Source Server Version : 50739
 Source Host           : localhost:3306
 Source Schema         : nemo

 Target Server Type    : MySQL
 Target Server Version : 50739
 File Encoding         : 65001

 Date: 27/11/2022 22:54:01
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS `task`;

-- ----------------------------
-- Table structure for task_main
-- ----------------------------
DROP TABLE IF EXISTS `task_main`;
CREATE TABLE `task_main` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `task_id` char(36) NOT NULL,
  `task_name` varchar(100) NOT NULL,
  `kwargs` varchar(8000) DEFAULT NULL,
  `state` varchar(40) NOT NULL,
  `result` varchar(4000) DEFAULT NULL,
  `received` datetime NOT NULL,
  `started` datetime DEFAULT NULL,
  `succeeded` datetime DEFAULT NULL,
  `progress_message` varchar(100) DEFAULT NULL,
  `cron_id` char(36) DEFAULT NULL COMMENT 'the id for cron task',
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=52 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for task_run
-- ----------------------------
DROP TABLE IF EXISTS `task_run`;
CREATE TABLE `task_run` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `task_id` char(36) NOT NULL,
  `task_name` varchar(100) NOT NULL,
  `kwargs` varchar(8000) DEFAULT NULL,
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
  `main_id` char(36) DEFAULT NULL COMMENT 'the id for main task',
  `last_run_id` char(36) DEFAULT NULL COMMENT 'last runtask id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=154 DEFAULT CHARSET=utf8;



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

 Date: 04/11/2022 10:11:33
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for key_word
-- ----------------------------
DROP TABLE IF EXISTS `key_word`;
CREATE TABLE `key_word` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `org_id` int(11) NOT NULL,
  `key_word` varchar(511) COLLATE utf8mb4_bin NOT NULL,
  `search_time` varchar(63) COLLATE utf8mb4_bin DEFAULT NULL,
  `exclude_words` varchar(2047) COLLATE utf8mb4_bin DEFAULT NULL,
  `check_mod` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `is_delete` tinyint(4) NOT NULL DEFAULT '0',
  `count` int(10) unsigned DEFAULT NULL,
  `create_datetime` datetime DEFAULT NULL,
  `update_datetime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC;

SET FOREIGN_KEY_CHECKS = 1;

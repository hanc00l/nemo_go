-- MySQL dump 10.13  Distrib 5.7.41, for osx10.18 (x86_64)
--
-- Host: 127.0.0.1    Database: nemo
-- ------------------------------------------------------
-- Server version	5.7.41

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `domain`
--

DROP TABLE IF EXISTS `domain`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `domain` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `domain` varchar(100) NOT NULL,
  `org_id` int(10) unsigned DEFAULT NULL,
  `workspace_id` int(11) NOT NULL,
  `pin_index` int(11) NOT NULL DEFAULT '0',
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_domain_org_id` (`org_id`),
  KEY `fk_domain_workspace_id` (`workspace_id`),
  CONSTRAINT `fk_domain_org_id` FOREIGN KEY (`org_id`) REFERENCES `organization` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_domain_workspace_id` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=1652 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `domain`
--

LOCK TABLES `domain` WRITE;
/*!40000 ALTER TABLE `domain` DISABLE KEYS */;
/*!40000 ALTER TABLE `domain` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `domain_attr`
--

DROP TABLE IF EXISTS `domain_attr`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `domain_attr` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `r_id` int(10) unsigned NOT NULL,
  `source` varchar(40) DEFAULT NULL,
  `tag` varchar(40) NOT NULL,
  `content` varchar(4000) DEFAULT NULL,
  `hash` char(32) DEFAULT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `index_domain_attr_hash` (`hash`) USING BTREE,
  KEY `index_domain_attr_ip_id` (`r_id`),
  CONSTRAINT `domain_attr_ibfk_1` FOREIGN KEY (`r_id`) REFERENCES `domain` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=38502 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `domain_attr`
--

LOCK TABLES `domain_attr` WRITE;
/*!40000 ALTER TABLE `domain_attr` DISABLE KEYS */;
/*!40000 ALTER TABLE `domain_attr` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `domain_color_tag`
--

DROP TABLE IF EXISTS `domain_color_tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `domain_color_tag` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `r_id` int(10) unsigned NOT NULL,
  `color` char(20) NOT NULL,
  `create_datetime` datetime DEFAULT NULL,
  `update_datetime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `fk_domain_color_tag_rid_unique` (`r_id`),
  KEY `fk_domain_color_tag_rid` (`r_id`) USING BTREE,
  CONSTRAINT `fk_domain_color_tag_rid` FOREIGN KEY (`r_id`) REFERENCES `domain` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `domain_color_tag`
--

LOCK TABLES `domain_color_tag` WRITE;
/*!40000 ALTER TABLE `domain_color_tag` DISABLE KEYS */;
/*!40000 ALTER TABLE `domain_color_tag` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `domain_memo`
--

DROP TABLE IF EXISTS `domain_memo`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `domain_memo` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `r_id` int(10) unsigned NOT NULL,
  `content` varchar(10000) DEFAULT NULL,
  `create_datetime` datetime DEFAULT NULL,
  `update_datetime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `fk_domain_memo_rid_unique` (`r_id`),
  CONSTRAINT `fk_domain_memo_rid` FOREIGN KEY (`r_id`) REFERENCES `domain` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `domain_memo`
--

LOCK TABLES `domain_memo` WRITE;
/*!40000 ALTER TABLE `domain_memo` DISABLE KEYS */;
/*!40000 ALTER TABLE `domain_memo` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ip`
--

DROP TABLE IF EXISTS `ip`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ip` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `ip` varchar(128) NOT NULL,
  `ip_int` bigint(20) NOT NULL,
  `org_id` int(10) unsigned DEFAULT NULL,
  `location` varchar(200) DEFAULT NULL,
  `status` varchar(20) DEFAULT NULL,
  `workspace_id` int(11) NOT NULL,
  `pin_index` int(11) NOT NULL DEFAULT '0',
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `index_ip_org_id` (`org_id`),
  KEY `fk_ip_workspace_id` (`workspace_id`),
  CONSTRAINT `fk_ip_org_id` FOREIGN KEY (`org_id`) REFERENCES `organization` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_ip_workspace_id` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=40 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ip`
--

LOCK TABLES `ip` WRITE;
/*!40000 ALTER TABLE `ip` DISABLE KEYS */;
/*!40000 ALTER TABLE `ip` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ip_attr`
--

DROP TABLE IF EXISTS `ip_attr`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ip_attr` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `r_id` int(10) unsigned NOT NULL,
  `source` varchar(40) DEFAULT NULL,
  `tag` varchar(40) NOT NULL,
  `content` varchar(4000) DEFAULT NULL,
  `hash` char(32) DEFAULT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `index_ip_attr_hash` (`hash`) USING BTREE,
  KEY `index_ip_attr_ip_id` (`r_id`),
  CONSTRAINT `fk_ip_attr_ip_id` FOREIGN KEY (`r_id`) REFERENCES `ip` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ip_attr`
--

LOCK TABLES `ip_attr` WRITE;
/*!40000 ALTER TABLE `ip_attr` DISABLE KEYS */;
/*!40000 ALTER TABLE `ip_attr` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ip_color_tag`
--

DROP TABLE IF EXISTS `ip_color_tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ip_color_tag` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `r_id` int(10) unsigned NOT NULL,
  `color` char(20) NOT NULL,
  `create_datetime` datetime DEFAULT NULL,
  `update_datetime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `fk_ip_color_tag_rid_unique` (`r_id`),
  KEY `fk_ip_color_tag_rid` (`r_id`),
  CONSTRAINT `ip_color_tag_ibfk_1` FOREIGN KEY (`r_id`) REFERENCES `ip` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ip_color_tag`
--

LOCK TABLES `ip_color_tag` WRITE;
/*!40000 ALTER TABLE `ip_color_tag` DISABLE KEYS */;
/*!40000 ALTER TABLE `ip_color_tag` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ip_memo`
--

DROP TABLE IF EXISTS `ip_memo`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ip_memo` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `r_id` int(10) unsigned NOT NULL,
  `content` varchar(10000) DEFAULT NULL,
  `create_datetime` datetime DEFAULT NULL,
  `update_datetime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `fk_ip_memo_rid_unqie` (`r_id`),
  CONSTRAINT `fk_ip_memo_rid` FOREIGN KEY (`r_id`) REFERENCES `ip` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ip_memo`
--

LOCK TABLES `ip_memo` WRITE;
/*!40000 ALTER TABLE `ip_memo` DISABLE KEYS */;
/*!40000 ALTER TABLE `ip_memo` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `key_word`
--

DROP TABLE IF EXISTS `key_word`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `key_word` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `org_id` int(11) NOT NULL,
  `key_word` varchar(511) COLLATE utf8mb4_bin NOT NULL,
  `search_time` varchar(63) COLLATE utf8mb4_bin DEFAULT NULL,
  `exclude_words` varchar(2047) COLLATE utf8mb4_bin DEFAULT NULL,
  `check_mod` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `is_delete` tinyint(4) NOT NULL DEFAULT '0',
  `count` int(10) unsigned DEFAULT NULL,
  `workspace_id` int(11) NOT NULL,
  `create_datetime` datetime DEFAULT NULL,
  `update_datetime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  KEY `fk_key_word_workspace_id` (`workspace_id`),
  CONSTRAINT `fk_key_word_workspace_id` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `key_word`
--

LOCK TABLES `key_word` WRITE;
/*!40000 ALTER TABLE `key_word` DISABLE KEYS */;
/*!40000 ALTER TABLE `key_word` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `organization`
--

DROP TABLE IF EXISTS `organization`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `organization` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `org_name` varchar(200) NOT NULL,
  `status` varchar(20) NOT NULL,
  `sort_order` int(10) unsigned NOT NULL DEFAULT '100',
  `workspace_id` int(11) NOT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_org_workspace_id` (`workspace_id`),
  CONSTRAINT `fk_org_workspace_id` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=49 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `organization`
--

LOCK TABLES `organization` WRITE;
/*!40000 ALTER TABLE `organization` DISABLE KEYS */;
/*!40000 ALTER TABLE `organization` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `port`
--

DROP TABLE IF EXISTS `port`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `port` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `ip_id` int(10) unsigned NOT NULL,
  `port` int(11) NOT NULL,
  `status` varchar(20) NOT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `index_port_ip_port` (`ip_id`,`port`),
  CONSTRAINT `fk_port_ip` FOREIGN KEY (`ip_id`) REFERENCES `ip` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=106 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `port`
--

LOCK TABLES `port` WRITE;
/*!40000 ALTER TABLE `port` DISABLE KEYS */;
/*!40000 ALTER TABLE `port` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `port_attr`
--

DROP TABLE IF EXISTS `port_attr`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `port_attr` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `r_id` int(10) unsigned NOT NULL,
  `source` varchar(40) DEFAULT NULL,
  `tag` varchar(40) NOT NULL,
  `content` varchar(4000) DEFAULT NULL,
  `hash` char(32) DEFAULT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `index_port_attr_hash` (`hash`),
  KEY `fk_port_attr_r_id` (`r_id`),
  CONSTRAINT `fk_port_attr_r_id` FOREIGN KEY (`r_id`) REFERENCES `port` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=259 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `port_attr`
--

LOCK TABLES `port_attr` WRITE;
/*!40000 ALTER TABLE `port_attr` DISABLE KEYS */;
/*!40000 ALTER TABLE `port_attr` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `task_cron`
--

DROP TABLE IF EXISTS `task_cron`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `task_cron` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `task_id` char(36) NOT NULL,
  `task_name` varchar(100) NOT NULL,
  `kwargs` varchar(8000) DEFAULT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  `cron_rule` varchar(200) NOT NULL COMMENT '定时规则',
  `lastrun_datetime` datetime DEFAULT NULL COMMENT '上次运行时间',
  `status` varchar(10) NOT NULL COMMENT '状态enable or disable',
  `run_count` int(11) DEFAULT NULL COMMENT '启动次数',
  `comment` varchar(200) DEFAULT NULL COMMENT '定时任务说明',
  `workspace_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_task_cron_workspace_id` (`workspace_id`),
  CONSTRAINT `fk_task_cron_workspace_id` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `task_cron`
--

LOCK TABLES `task_cron` WRITE;
/*!40000 ALTER TABLE `task_cron` DISABLE KEYS */;
/*!40000 ALTER TABLE `task_cron` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `task_main`
--

DROP TABLE IF EXISTS `task_main`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
  `workspace_id` int(11) NOT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_task_main_workspace_id` (`workspace_id`),
  CONSTRAINT `fk_task_main_workspace_id` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=29 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `task_main`
--

LOCK TABLES `task_main` WRITE;
/*!40000 ALTER TABLE `task_main` DISABLE KEYS */;
/*!40000 ALTER TABLE `task_main` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `task_run`
--

DROP TABLE IF EXISTS `task_run`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
  `workspace_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_task_run_workspace_id` (`workspace_id`),
  CONSTRAINT `fk_task_run_workspace_id` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=782 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `task_run`
--

LOCK TABLES `task_run` WRITE;
/*!40000 ALTER TABLE `task_run` DISABLE KEYS */;
/*!40000 ALTER TABLE `task_run` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user`
--

DROP TABLE IF EXISTS `user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_name` varchar(100) NOT NULL,
  `user_password` char(48) NOT NULL,
  `user_description` varchar(200) DEFAULT NULL,
  `user_role` varchar(40) NOT NULL,
  `state` varchar(40) NOT NULL,
  `sort_order` int(11) NOT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id_uindex` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user`
--

LOCK TABLES `user` WRITE;
/*!40000 ALTER TABLE `user` DISABLE KEYS */;
INSERT INTO `user` VALUES (1,'nemo','648ce596dba3b408b523d3d1189b15070123456789abcdef','默认超级管理员','superadmin','enable',100,'2023-02-26 11:43:20','2023-03-02 15:40:23');
/*!40000 ALTER TABLE `user` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_workspace`
--

DROP TABLE IF EXISTS `user_workspace`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user_workspace` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `workspace_id` int(11) NOT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_workspace_id_uindex` (`id`),
  KEY `fk_userid` (`user_id`),
  KEY `fk_workspaceid` (`workspace_id`),
  CONSTRAINT `fk_userid` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_workspaceid` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_workspace`
--

LOCK TABLES `user_workspace` WRITE;
/*!40000 ALTER TABLE `user_workspace` DISABLE KEYS */;
INSERT INTO `user_workspace` VALUES (1,1,1,'2023-03-01 23:05:39','2023-03-01 23:05:39');
/*!40000 ALTER TABLE `user_workspace` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `vulnerability`
--

DROP TABLE IF EXISTS `vulnerability`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `vulnerability` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `target` varchar(100) NOT NULL,
  `url` varchar(200) NOT NULL,
  `poc_file` varchar(200) NOT NULL,
  `source` varchar(40) NOT NULL,
  `extra` varchar(4000) DEFAULT NULL,
  `hash` char(32) NOT NULL,
  `workspace_id` int(11) NOT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_vul_workspace_id` (`workspace_id`),
  CONSTRAINT `fk_vul_workspace_id` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=22 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `vulnerability`
--

LOCK TABLES `vulnerability` WRITE;
/*!40000 ALTER TABLE `vulnerability` DISABLE KEYS */;
/*!40000 ALTER TABLE `vulnerability` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `workspace`
--

DROP TABLE IF EXISTS `workspace`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `workspace` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `workspace_name` varchar(100) NOT NULL,
  `workspace_guid` char(36) NOT NULL,
  `workspace_description` varchar(200) DEFAULT NULL,
  `state` varchar(20) NOT NULL,
  `sort_order` int(11) NOT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `workspace_id_uindex` (`id`),
  UNIQUE KEY `workspace_space_guid_uindex` (`workspace_guid`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `workspace`
--

LOCK TABLES `workspace` WRITE;
/*!40000 ALTER TABLE `workspace` DISABLE KEYS */;
INSERT INTO `workspace` VALUES (1,'默认','b0c79065-7ff7-32ae-cc18-864ccd8f7717','默认工作空间','enable',100,'2023-02-26 11:40:00','2023-02-26 11:40:05');
/*!40000 ALTER TABLE `workspace` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

--
-- Table structure for table `ip_http`
--

DROP TABLE IF EXISTS `ip_http`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ip_http` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `r_id` int(11) unsigned NOT NULL,
  `source` varchar(40) NOT NULL,
  `tag` varchar(40) NOT NULL,
  `content` varchar(16000) NOT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `ip_http_id_uindex` (`id`),
  KEY `fk_ip_http_rid` (`r_id`),
  CONSTRAINT `fk_ip_http_rid` FOREIGN KEY (`r_id`) REFERENCES `port` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `domain_http`
--

DROP TABLE IF EXISTS `domain_http`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `domain_http` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `r_id` int(10) unsigned NOT NULL,
  `port` int(11) NOT NULL,
  `source` varchar(40) NOT NULL,
  `tag` varchar(40) NOT NULL,
  `content` varchar(16000) NOT NULL,
  `create_datetime` datetime NOT NULL,
  `update_datetime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `domain_http_id_uindex` (`id`),
  KEY `fk_domain_http_rid` (`r_id`),
  CONSTRAINT `fk_domain_http_rid` FOREIGN KEY (`r_id`) REFERENCES `domain` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2023-03-10 16:00:14
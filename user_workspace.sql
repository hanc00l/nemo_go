/* v2.8->v2.9，对数据库增加workspace、user及user_workspace字段，并对已有表进行相应的调整 */

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

/* 增加workspace_id字段  */

alter table ip add workspace_id int null after status;
alter table ip add constraint fk_ip_workspace_id foreign key (workspace_id) references workspace (id) on delete cascade;

alter table domain add workspace_id int null after org_id;
alter table domain add constraint fk_domain_workspace_id foreign key (workspace_id) references workspace (id) on delete cascade;

alter table organization add workspace_id int null after sort_order;
alter table organization add constraint fk_org_workspace_id foreign key (workspace_id) references workspace (id)on delete cascade;

alter table task_cron add workspace_id int null;
alter table task_cron add constraint fk_task_cron_workspace_id foreign key (workspace_id) references workspace (id) on delete cascade;

alter table task_main add workspace_id int null after cron_id;
alter table task_main add constraint fk_task_main_workspace_id foreign key (workspace_id) references workspace (id) on delete cascade;

alter table task_run add workspace_id int null;
alter table task_run add constraint fk_task_run_workspace_id foreign key (workspace_id) references workspace (id) on delete cascade;

alter table vulnerability add workspace_id int null after hash;
alter table vulnerability add constraint fk_vul_workspace_id foreign key (workspace_id) references workspace (id) on delete cascade;

alter table key_word add workspace_id int null after count;
alter table key_word add constraint fk_key_word_workspace_id foreign key (workspace_id) references workspace (id) on delete cascade;

/* 设置已有的资源为默认的工作空间*/

update ip set workspace_id=1 where 1=1;
alter table ip modify workspace_id int not null;

update domain set workspace_id=1 where 1=1;
alter table domain modify workspace_id int not null;

update key_word set workspace_id=1 where 1=1; 
alter table key_word modify workspace_id int not null;

update organization set workspace_id=1 where 1=1;
alter table organization modify workspace_id int not null;

update task_cron set workspace_id=1 where 1=1;
alter table task_cron modify workspace_id int not null;

update task_main set workspace_id=1 where 1=1;
alter table task_main modify workspace_id int not null;

update task_run set workspace_id=1 where 1=1;
alter table task_run modify workspace_id int not null;

update vulnerability set workspace_id=1 where 1=1;
alter table vulnerability modify workspace_id int not null;

/* 去除ip与域名的唯一性 */
drop index index_ip_ip on ip;
drop index index_ip_ip_int on ip;
drop index index_domain_domain on domain;

/* ip与domain增加置顶字段 */
alter table ip add pin_index int default 0 not null after workspace_id;
alter table domain add pin_index int default 0 not null after workspace_id;

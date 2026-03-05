-- create database for xxl-job
CREATE database if NOT EXISTS `xxl_job` default character set utf8mb4 collate utf8mb4_unicode_ci;
use `xxl_job`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for xxl_job_group
-- ----------------------------
DROP TABLE IF EXISTS `xxl_job_group`;
CREATE TABLE `xxl_job_group` (
  `id`           INT         NOT NULL AUTO_INCREMENT,
  `app_name`     VARCHAR(64) NOT NULL             COMMENT '执行器AppName',
  `title`        VARCHAR(12) NOT NULL             COMMENT '执行器名称',
  `address_type` TINYINT     NOT NULL DEFAULT '0' COMMENT '执行器地址类型：0=自动注册、1=手动录入',
  `address_list` TEXT                             COMMENT '执行器地址列表，多地址逗号分隔',
  `update_time`  DATETIME    DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_app_name` (`app_name`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for xxl_job_lock
-- ----------------------------
DROP TABLE IF EXISTS `xxl_job_lock`;
CREATE TABLE `xxl_job_lock` (
  `lock_name` VARCHAR(50) NOT NULL COMMENT '锁名称',
  PRIMARY KEY (`lock_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Records of xxl_job_lock
-- ----------------------------
BEGIN;
INSERT INTO `xxl_job_lock` (`lock_name`) VALUES ('schedule_lock');
COMMIT;

-- ----------------------------
-- Table structure for xxl_job_info
-- ----------------------------
DROP TABLE IF EXISTS `xxl_job_info`;
CREATE TABLE `xxl_job_info` (
  `id`                         INT          NOT NULL AUTO_INCREMENT,
  `job_group`                  INT          NOT NULL COMMENT '执行器主键ID',
  `job_desc`                   VARCHAR(255) NOT NULL,
  `add_time`                   DATETIME     DEFAULT NULL,
  `update_time`                DATETIME     DEFAULT NULL,
  `author`                     VARCHAR(64)  DEFAULT NULL                  COMMENT '作者',
  `alarm_email`                VARCHAR(255) DEFAULT NULL                  COMMENT '报警邮件',
  `schedule_type`              VARCHAR(50)  NOT NULL DEFAULT 'NONE'       COMMENT '调度类型',
  `schedule_conf`              VARCHAR(128) DEFAULT NULL                  COMMENT '调度配置，值含义取决于调度类型',
  `misfire_strategy`           VARCHAR(50)  NOT NULL DEFAULT 'DO_NOTHING' COMMENT '调度过期策略',
  `executor_route_strategy`    VARCHAR(50)  DEFAULT NULL                  COMMENT '执行器路由策略',
  `executor_handler`           VARCHAR(255) DEFAULT NULL                  COMMENT '执行器任务handler',
  `executor_param`             VARCHAR(512) DEFAULT NULL                  COMMENT '执行器任务参数',
  `executor_block_strategy`    VARCHAR(50)  DEFAULT NULL                  COMMENT '阻塞处理策略',
  `executor_timeout`           INT          NOT NULL DEFAULT '0'          COMMENT '任务执行超时时间，单位秒',
  `executor_fail_retry_count`  INT          NOT NULL DEFAULT '0'          COMMENT '失败重试次数',
  `glue_type`                  VARCHAR(50)  NOT NULL                      COMMENT 'GLUE类型',
  `glue_source`                MEDIUMTEXT                                 COMMENT 'GLUE源代码',
  `glue_remark`                VARCHAR(128) DEFAULT NULL                  COMMENT 'GLUE备注',
  `glue_updatetime`            DATETIME     DEFAULT NULL                  COMMENT 'GLUE更新时间',
  `child_jobid`                VARCHAR(255) DEFAULT NULL                  COMMENT '子任务ID，多个逗号分隔',
  `trigger_status`             TINYINT      NOT NULL DEFAULT '0'          COMMENT '调度状态：0-停止，1-运行',
  `trigger_last_time`          BIGINT       NOT NULL DEFAULT '0'          COMMENT '上次调度时间',
  `trigger_next_time`          BIGINT       NOT NULL DEFAULT '0'          COMMENT '下次调度时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_handler` (`job_group`, `executor_handler`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


-- ----------------------------
-- Table structure for xxl_job_log
-- ----------------------------
DROP TABLE IF EXISTS `xxl_job_log`;
CREATE TABLE `xxl_job_log` (
  `id`                        BIGINT       NOT NULL AUTO_INCREMENT,
  `job_group`                 INT          NOT NULL             COMMENT '执行器主键ID',
  `job_id`                    INT          NOT NULL             COMMENT '任务，主键ID',
  `executor_address`          VARCHAR(255) DEFAULT NULL         COMMENT '执行器地址，本次执行的地址',
  `executor_handler`          VARCHAR(255) DEFAULT NULL         COMMENT '执行器任务handler',
  `executor_param`            VARCHAR(512) DEFAULT NULL         COMMENT '执行器任务参数',
  `executor_sharding_param`   VARCHAR(20)  DEFAULT NULL         COMMENT '执行器任务分片参数，格式如 1/2',
  `executor_fail_retry_count` INT          NOT NULL DEFAULT '0' COMMENT '失败重试次数',
  `trigger_time`              DATETIME     DEFAULT NULL         COMMENT '调度-时间',
  `trigger_code`              INT          NOT NULL             COMMENT '调度-结果',
  `trigger_msg`               TEXT                              COMMENT '调度-日志',
  `handle_time`               DATETIME     DEFAULT NULL         COMMENT '执行-时间',
  `handle_code`               INT          NOT NULL             COMMENT '执行-状态',
  `handle_msg`                TEXT                              COMMENT '执行-日志',
  `alarm_status`              TINYINT      NOT NULL DEFAULT '0' COMMENT '告警状态：0-默认、1-无需告警、2-告警成功、3-告警失败',
  PRIMARY KEY (`id`),
  KEY `I_trigger_time` (`trigger_time`),
  KEY `I_handle_code` (`handle_code`),
  KEY `I_jobid_jobgroup` (`job_id`,`job_group`),
  KEY `I_job_id` (`job_id`)
) ENGINE=InnoDB AUTO_INCREMENT=364481 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Records of xxl_job_log
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for xxl_job_log_report
-- ----------------------------
DROP TABLE IF EXISTS `xxl_job_log_report`;
CREATE TABLE `xxl_job_log_report` (
  `id`            INT      NOT NULL AUTO_INCREMENT,
  `trigger_day`   DATETIME DEFAULT NULL         COMMENT '调度-时间',
  `running_count` INT      NOT NULL DEFAULT '0' COMMENT '运行中-日志数量',
  `suc_count`     INT      NOT NULL DEFAULT '0' COMMENT '执行成功-日志数量',
  `fail_count`    INT      NOT NULL DEFAULT '0' COMMENT '执行失败-日志数量',
  `update_time`   DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `i_trigger_day` (`trigger_day`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Records of xxl_job_log_report
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for xxl_job_logglue
-- ----------------------------
DROP TABLE IF EXISTS `xxl_job_logglue`;
CREATE TABLE `xxl_job_logglue` (
  `id`          INT          NOT NULL AUTO_INCREMENT,
  `job_id`      INT          NOT NULL     COMMENT '任务，主键ID',
  `glue_type`   VARCHAR(50)  DEFAULT NULL COMMENT 'GLUE类型',
  `glue_source` MEDIUMTEXT                COMMENT 'GLUE源代码',
  `glue_remark` VARCHAR(128) NOT NULL     COMMENT 'GLUE备注',
  `add_time`    DATETIME     DEFAULT NULL,
  `update_time` DATETIME     DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Records of xxl_job_logglue
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for xxl_job_registry
-- ----------------------------
DROP TABLE IF EXISTS `xxl_job_registry`;
CREATE TABLE `xxl_job_registry` (
  `id`             INT          NOT NULL AUTO_INCREMENT,
  `registry_group` VARCHAR(50)  NOT NULL,
  `registry_key`   VARCHAR(255) NOT NULL,
  `registry_value` VARCHAR(255) NOT NULL,
  `update_time`    DATETIME     DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `i_g_k_v` (`registry_group`,`registry_key`,`registry_value`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1150 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Records of xxl_job_registry
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for xxl_job_user
-- ----------------------------
DROP TABLE IF EXISTS `xxl_job_user`;
CREATE TABLE `xxl_job_user` (
  `id`       int          NOT NULL AUTO_INCREMENT,
  `username` varchar(50)  NOT NULL       COMMENT '账号',
  `password` varchar(100) NOT NULL       COMMENT '密码加密信息',
  `token`    varchar(100) DEFAULT NULL   COMMENT '登录token',
  `role`     tinyint      NOT NULL       COMMENT '角色：0-普通用户、1-管理员',
  `permission` varchar(255) DEFAULT NULL COMMENT '权限：执行器ID列表，多个逗号分割',
  PRIMARY KEY (`id`),
  UNIQUE KEY `i_username` (`username`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Records of xxl_job_user
-- password:
-- echo 'admin123' | shasum -a 256
-- ----------------------------
BEGIN;
INSERT INTO `xxl_job_user`
(`id`, `username`, `password`, `token`, `role`, `permission`)
VALUES
(1, 'admin', '240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9', '7c511e26b9724d6fafe6fc2de631a2ff', 1, NULL);
COMMIT;

SET FOREIGN_KEY_CHECKS = 1;

-- insert job_group and job_info for
INSERT INTO xxl_job_group
(id, app_name, title, address_type, address_list)
VALUES
(100,'confx','confx',0,'');

INSERT INTO xxl_job_info
(
  job_group,
  job_desc,
  author,
  glue_type,
  schedule_type,
  schedule_conf,
  executor_handler,
  trigger_status,
  glue_updatetime,
  executor_block_strategy,
  executor_route_strategy
)
VALUES
(
  100,
  'confx_test',
  'test',
  'BEAN',
  'CRON',
  '0/5 * * * * ?', -- per 5 second
  'confx_test',
  1,
  now(),
  'SERIAL_EXECUTION',
  'ROUND'
);
/*
Navicat MySQL Data Transfer

Source Server         : paas
Source Server Version : 50511
Source Host           : 192.168.1.102:3306
Source Database       : paas

Target Server Type    : MYSQL
Target Server Version : 50511
File Encoding         : 65001

Date: 2017-04-02 18:01:05
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for app
-- ----------------------------
DROP TABLE IF EXISTS `app`;
CREATE TABLE `app` (
  `name` varchar(256) NOT NULL,
  `region` varchar(256) DEFAULT NULL,
  `memory` varchar(11) DEFAULT NULL,
  `cpu` varchar(11) DEFAULT NULL,
  `instance_count` int(11) DEFAULT NULL,
  `envs` varchar(1024) DEFAULT NULL,
  `ports` varchar(1024) DEFAULT NULL,
  `image` varchar(1024) DEFAULT NULL,
  `command` varchar(1024) DEFAULT NULL,
  `status` int(1) DEFAULT NULL,
  `user_name` varchar(256) DEFAULT NULL,
  `remark` varchar(1024) DEFAULT NULL,
  `create_at` datetime NOT NULL,
  `revise_at` datetime NOT NULL,
  PRIMARY KEY (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

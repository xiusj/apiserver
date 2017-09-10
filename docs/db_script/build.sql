/*
Navicat MySQL Data Transfer

Source Server         : paas
Source Server Version : 50511
Source Host           : 192.168.1.102:3306
Source Database       : paas

Target Server Type    : MYSQL
Target Server Version : 50511
File Encoding         : 65001

Date: 2017-04-02 18:13:22
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for build
-- ----------------------------
DROP TABLE IF EXISTS `build`;
CREATE TABLE `build` (
  `id` int(11) NOT NULL,
  `appName` varchar(256) CHARACTER SET latin1 NOT NULL COMMENT '构建应用的名称',
  `version` varchar(256) CHARACTER SET latin1 NOT NULL COMMENT '构建应用的镜像版本',
  `remark` varchar(256) CHARACTER SET latin1 NOT NULL COMMENT '备注：说明该镜像是做什么用的\r\n',
  `base_image` varchar(256) CHARACTER SET latin1 NOT NULL COMMENT '基础镜像',
  `image` varchar(256) CHARACTER SET latin1 NOT NULL COMMENT '构建完成后生成的镜像名称\r\n',
  `tarball` varchar(256) CHARACTER SET latin1 DEFAULT NULL COMMENT '应用的tar文件',
  `registry` varchar(256) CHARACTER SET latin1 DEFAULT NULL COMMENT '经常推送的仓库地址',
  `repository` varchar(256) CHARACTER SET latin1 DEFAULT NULL COMMENT '代码库的仓库地址',
  `branch` varchar(256) CHARACTER SET latin1 DEFAULT NULL COMMENT '代码库的代码分支',
  `user_id` int(11) DEFAULT NULL COMMENT '当前用户的user id\r\n',
  `user_name` varchar(256) CHARACTER SET latin1 DEFAULT NULL COMMENT '当前用户的名字',
  `create_at` datetime DEFAULT NULL COMMENT '镜像的构建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

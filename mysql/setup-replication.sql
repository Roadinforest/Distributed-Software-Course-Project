-- MySQL 主从复制配置脚本
-- 此脚本在主库执行，创建复制用户

-- 创建复制用户
CREATE USER IF NOT EXISTS 'repl_user'@'%' IDENTIFIED BY 'repl123';
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'repl_user'@'%';

-- 刷新权限
FLUSH PRIVILEGES;

-- 显示主库状态
SHOW MASTER STATUS;

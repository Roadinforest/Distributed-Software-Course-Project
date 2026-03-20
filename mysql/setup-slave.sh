#!/bin/bash
# 从库复制配置脚本

set -e

echo "Waiting for master to be ready..."
sleep 10

echo "Configuring slave replication..."

# 等待主库完全就绪
until mysql -h mysql-master -uroot -proot123 -e "SELECT 1" &>/dev/null; do
    echo "Master not ready, waiting..."
    sleep 2
done

# 配置主从复制
mysql -h mysql-slave -uroot -proot123 <<EOF
STOP SLAVE;

-- 使用GTID自动定位
CHANGE MASTER TO
    MASTER_HOST='mysql-master',
    MASTER_USER='repl_user',
    MASTER_PASSWORD='repl123',
    MASTER_AUTO_POSITION=1;

START SLAVE;

-- 检查复制状态
SHOW SLAVE STATUS\G
EOF

echo "Slave replication configured successfully!"

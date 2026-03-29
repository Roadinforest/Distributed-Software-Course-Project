-- 订单数据库初始化脚本
CREATE DATABASE IF NOT EXISTS order_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE order_db;

-- 订单表
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY COMMENT '订单ID（雪花算法生成）',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    product_id BIGINT NOT NULL COMMENT '商品ID',
    quantity INT DEFAULT 1 COMMENT '购买数量',
    status TINYINT DEFAULT 0 COMMENT '订单状态：0=待处理，1=已创建，2=已取消，3=创建失败',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_user_id (user_id),
    INDEX idx_product_id (product_id),
    INDEX idx_user_product (user_id, product_id),
    UNIQUE KEY uk_user_product (user_id, product_id) COMMENT '同一用户同一商品唯一订单'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- 演示数据：初始化一些秒杀商品的库存
-- 注意：实际秒杀时，库存应该从商品服务同步
INSERT INTO orders (id, user_id, product_id, quantity, status) VALUES
    (1, 1001, 1, 1, 1),
    (2, 1002, 1, 1, 1)
ON DUPLICATE KEY UPDATE id=id;

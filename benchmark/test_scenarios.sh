#!/bin/bash
# Kong + Consul 流量治理测试场景
# 测试熔断、限流、降级等功能

set -e

KONG_PROXY="${KONG_PROXY:-http://localhost:8000}"
CONSUL_HTTP="${CONSUL_HTTP:-http://localhost:8500}"
USERS_SVC="/api/v1/users"
PRODUCTS_SVC="/api/v1/products"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}   Kong流量治理测试 - 服务注册发现与配置管理  ${NC}"
echo -e "${BLUE}================================================${NC}"

# 1. 检查Consul服务注册
echo -e "\n${GREEN}[1] 检查Consul服务注册发现${NC}"
echo "----------------------------------------"
echo "查看已注册的服务:"
curl -s "$CONSUL_HTTP/v1/catalog/services" | python3 -m json.tool 2>/dev/null || curl -s "$CONSUL_HTTP/v1/catalog/services"

echo -e "\n查看user-service实例:"
curl -s "$CONSUL_HTTP/v1/health/service/user-service?passing=true" | python3 -m json.tool 2>/dev/null || curl -s "$CONSUL_HTTP/v1/health/service/user-service?passing=true"

# 2. 查看Kong路由配置
echo -e "\n${GREEN}[2] 查看Kong路由配置${NC}"
echo "----------------------------------------"
echo "查看已配置的服务:"
curl -s "$KONG_PROXY:8001/services" | python3 -m json.tool 2>/dev/null || curl -s "$KONG_PROXY:8001/services"

echo -e "\n查看路由:"
curl -s "$KONG_PROXY:8001/routes" | python3 -m json.tool 2>/dev/null || curl -s "$KONG_PROXY:8001/routes"

# 3. Consul动态配置测试
echo -e "\n${GREEN}[3] Consul动态配置测试${NC}"
echo "----------------------------------------"
echo "写入测试配置到Consul KV:"
curl -s -X PUT "$CONSUL_HTTP/v1/kv/config/service/rate-limit" \
    -d '{"requests_per_second":100,"burst_size":200}'

echo -e "\n读取配置:"
curl -s "$CONSUL_HTTP/v1/kv/config/service/rate-limit" | python3 -m json.tool 2>/dev/null || curl -s "$CONSUL_HTTP/v1/kv/config/service/rate-limit"

# 4. 通过Kong网关调用服务
echo -e "\n${GREEN}[4] 通过Kong网关调用服务测试${NC}"
echo "----------------------------------------"
echo "调用用户服务(POST /api/v1/users/login):"
curl -s -X POST "$KONG_PROXY$USERS_SVC/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"test123"}' | python3 -m json.tool 2>/dev/null || \
    curl -s -X POST "$KONG_PROXY$USERS_SVC/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"test123"}'

echo -e "\n调用商品服务(GET /api/v1/products):"
curl -s "$KONG_PROXY$PRODUCTS_SVC" | python3 -m json.tool 2>/dev/null || \
    curl -s "$KONG_PROXY$PRODUCTS_SVC"

# 5. 健康检查
echo -e "\n${GREEN}[5] Kong健康检查状态${NC}"
echo "----------------------------------------"
echo "user-service上游健康状态:"
curl -s "$KONG_PROXY:8001/upstreams/user-service-upstream/health" | python3 -m json.tool 2>/dev/null || \
    curl -s "$KONG_PROXY:8001/upstreams/user-service-upstream/health"

echo -e "\n${GREEN}测试完成!${NC}"
echo ""
echo "后续可以使用压测工具验证流量治理效果:"
echo "  ./run_hey_test.sh -t rate-limit -n 5000 -c 200"
echo "  ./run_wrk_test.sh -d 30s -t 12 -c 400"
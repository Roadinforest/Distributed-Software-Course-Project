#!/bin/bash
# 压力测试脚本 - 使用hey进行HTTP压测
# 替代JMeter的Go编写的高性能压测工具

set -e

# 配置
GATEWAY_URL="${GATEWAY_URL:-http://localhost:8000}"
KONG_URL="${KONG_URL:-http://localhost:8000}"
NGINX_URL="${NGINX_URL:-http://localhost:80}"
USERS_SERVICE="${USERS_SERVICE:-/api/v1/users}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_usage() {
    echo "Usage: ./run_hey_test.sh [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -n NUM        Total number of requests (default: 10000)"
    echo "  -c NUM        Concurrent connections (default: 100)"
    echo "  -m METHOD     HTTP method (default: GET)"
    echo "  -h HOST       Target host header"
    echo "  -t TYPE       Test type: rate-limit, circuit-breaker, normal (default: normal)"
    echo "  --gateway URL Use specific gateway URL"
    echo "  --help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./run_hey_test.sh -n 10000 -c 100                    # Normal load test"
    echo "  ./run_hey_test.sh -t rate-limit -n 5000 -c 200      # Test rate limiting"
    echo "  ./run_hey_test.sh --gateway http://kong:8000         # Test via Kong"
}

# 解析参数
NUM_REQUESTS=10000
CONCURRENT=100
METHOD="GET"
HOST=""
TEST_TYPE="normal"

while [[ $# -gt 0 ]]; do
    case $1 in
        -n)
            NUM_REQUESTS="$2"
            shift 2
            ;;
        -c)
            CONCURRENT="$2"
            shift 2
            ;;
        -m)
            METHOD="$2"
            shift 2
            ;;
        -h)
            HOST="$2"
            shift 2
            ;;
        -t)
            TEST_TYPE="$2"
            shift 2
            ;;
        --gateway)
            GATEWAY_URL="$2"
            shift 2
            ;;
        --help)
            echo_usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo_usage
            exit 1
            ;;
    esac
done

# 检查hey是否安装
check_hey() {
    if ! command -v hey &> /dev/null; then
        echo -e "${YELLOW}hey not found, installing...${NC}"
        go install github.com/rakyll/hey@latest
    fi
}

# 测试健康检查
test_health() {
    echo -e "${GREEN}=== Testing Health Check ===${NC}"
    URL="${GATEWAY_URL}/healthz"
    hey -n 100 -c 10 "$URL"
}

# 测试限流
test_rate_limit() {
    echo -e "${GREEN}=== Testing Rate Limiting ===${NC}"
    echo -e "${YELLOW}Sending 1000 requests with 200 concurrent connections${NC}"
    echo -e "${YELLOW}Expected: Some requests should return 429 Too Many Requests${NC}"

    URL="${GATEWAY_URL}${USERS_SERVICE}/login"

    # 使用POST测试，带body
    hey -n 1000 -c 200 -m POST -T "application/json" -d '{"username":"test","password":"test"}' "$URL"

    echo ""
    echo "Rate limit test completed. Check for 429 responses."
}

# 测试熔断
test_circuit_breaker() {
    echo -e "${GREEN}=== Testing Circuit Breaker ===${NC}"
    echo -e "${YELLOW}Simulating high error rate to trigger circuit breaker${NC}"

    URL="${GATEWAY_URL}${USERS_SERVICE}/profile"

    # 快速发送大量请求
    hey -n 5000 -c 100 "$URL"

    echo ""
    echo "Circuit breaker test completed. Check service availability."
}

# 普通负载测试
test_normal() {
    echo -e "${GREEN}=== Normal Load Test ===${NC}"
    echo "Requests: $NUM_REQUESTS"
    echo "Concurrent: $CONCURRENT"
    echo "URL: ${GATEWAY_URL}${USERS_SERVICE}/login"

    URL="${GATEWAY_URL}${USERS_SERVICE}/login"

    # POST请求测试登录
    hey -n "$NUM_REQUESTS" -c "$CONCURRENT" -m POST -T "application/json" -d '{"username":"testuser","password":"test123"}' "$URL"
}

# 主测试流程
main() {
    echo -e "${YELLOW}======================================${NC}"
    echo -e "${YELLOW}   Pressure Test - hey benchmark     ${NC}"
    echo -e "${YELLOW}======================================${NC}"
    echo ""

    check_hey

    case $TEST_TYPE in
        rate-limit)
            test_rate_limit
            ;;
        circuit-breaker)
            test_circuit_breaker
            ;;
        health)
            test_health
            ;;
        normal|*)
            test_normal
            ;;
    esac

    echo ""
    echo -e "${GREEN}Test completed!${NC}"
}

main
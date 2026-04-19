#!/bin/bash
# wrk压力测试脚本
# wrk是Go编写的高性能HTTP压测工具，支持Lua脚本

set -e

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8000}"
USERS_SERVICE="/api/v1/users"
PRODUCT_SERVICE="/api/v1/products"

echo_usage() {
    echo "Usage: ./run_wrk_test.sh [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -d DURATION   Test duration (default: 30s)"
    echo "  -t THREADS     Number of threads (default: 12)"
    echo "  -c CONNECTIONS Number of connections (default: 400)"
    echo "  --url URL      Target URL"
    echo "  --help         Show help"
}

DURATION="30s"
THREADS=12
CONNECTIONS=400
URL="${GATEWAY_URL}${USERS_SERVICE}/login"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

while [[ $# -gt 0 ]]; do
    case $1 in
        -d)
            DURATION="$2"
            shift 2
            ;;
        -t)
            THREADS="$2"
            shift 2
            ;;
        -c)
            CONNECTIONS="$2"
            shift 2
            ;;
        --url)
            URL="$2"
            shift 2
            ;;
        --help)
            echo_usage
            exit 0
            ;;
        *)
            echo "Unknown: $1"
            exit 1
            ;;
    esac
done

check_wrk() {
    if ! command -v wrk &> /dev/null; then
        echo "wrk not found. Installing..."
        # Ubuntu/Debian
        sudo apt-get update && sudo apt-get install -y wrk || {
            echo "Please install wrk manually: https://github.com/wg/wrk"
            exit 1
        }
    fi
}

# 简单POST测试
run_post_test() {
    echo "=== POST Login Test ==="
    echo "URL: $URL"
    echo "Duration: $DURATION, Threads: $THREADS, Connections: $CONNECTIONS"

    wrk -t"$THREADS" -c"$CONNECTIONS" -d"$DURATION" \
        -H "Content-Type: application/json" \
        -H "X-Gateway: Kong" \
        -s <(cat <<'LUA'
wrk.method = "POST"
wrk.body   = '{"username":"testuser","password":"test123"}'
wrk.headers["Content-Type"] = "application/json"
LUA
) "$URL"
}

# GET测试
run_get_test() {
    local target="${GATEWAY_URL}${PRODUCT_SERVICE}"
    echo "=== GET Products Test ==="
    echo "URL: $target"

    wrk -t"$THREADS" -c"$CONNECTIONS" -d"$DURATION" \
        -H "X-Gateway: Kong" \
        "$target"
}

# 限流测试
run_rate_limit_test() {
    echo "=== Rate Limit Test ==="
    echo "Sending high concurrent requests to trigger rate limiting..."

    wrk -t20 -c500 -d20s \
        -H "Content-Type: application/json" \
        -s <(cat <<'LUA'
wrk.method = "POST"
wrk.body   = '{"username":"loadtest","password":"password"}'
LUA
) "$URL"
}

main() {
    echo "========================================"
    echo "   wrk HTTP Benchmark (Go压测工具)"
    echo "========================================"
    echo ""

    check_wrk

    echo "1. POST Login Test"
    run_post_test
    echo ""

    echo "2. GET Products Test"
    run_get_test
    echo ""

    echo "3. Rate Limit Test (高并发测试)"
    run_rate_limit_test

    echo ""
    echo "Test completed!"
}

main
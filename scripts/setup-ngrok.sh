#!/bin/bash

# Script tự động lấy ngrok URLs và cập nhật .env file
# Yêu cầu: ngrok đang chạy và jq đã được cài đặt

set -e

# Màu sắc cho output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔍 Đang lấy ngrok URLs...${NC}"

# Kiểm tra ngrok có đang chạy không
if ! curl -s http://localhost:4040/api/tunnels > /dev/null 2>&1; then
    echo -e "${RED}❌ Ngrok không đang chạy. Hãy khởi động ngrok trước:${NC}"
    echo -e "   ${YELLOW}ngrok http 5173${NC} (cho frontend)"
    echo -e "   ${YELLOW}ngrok http 3000${NC} (cho backend)"
    echo -e "   ${YELLOW}hoặc: ngrok start --all${NC} (nếu dùng config file)"
    exit 1
fi

# Kiểm tra jq có được cài đặt không
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}⚠️  jq chưa được cài đặt. Đang cài đặt...${NC}"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install jq
    else
        echo -e "${RED}Vui lòng cài đặt jq: https://stedolan.github.io/jq/download/${NC}"
        exit 1
    fi
fi

# Lấy tất cả tunnels
TUNNELS=$(curl -s http://localhost:4040/api/tunnels)

# Lấy frontend URL (port 5173)
FRONTEND_URL=$(echo $TUNNELS | jq -r '.tunnels[] | select(.config.addr == "localhost:5173" or .config.addr == "127.0.0.1:5173") | .public_url' | head -n 1)

# Lấy backend URL (port 3000)
BACKEND_URL=$(echo $TUNNELS | jq -r '.tunnels[] | select(.config.addr == "localhost:3000" or .config.addr == "127.0.0.1:3000") | .public_url' | head -n 1)

# Kiểm tra nếu không tìm thấy
if [ -z "$FRONTEND_URL" ]; then
    echo -e "${YELLOW}⚠️  Không tìm thấy frontend tunnel (port 5173)${NC}"
    echo -e "   Hãy đảm bảo đã chạy: ${YELLOW}ngrok http 5173${NC}"
fi

if [ -z "$BACKEND_URL" ]; then
    echo -e "${YELLOW}⚠️  Không tìm thấy backend tunnel (port 3000)${NC}"
    echo -e "   Hãy đảm bảo đã chạy: ${YELLOW}ngrok http 3000${NC}"
fi

if [ -z "$FRONTEND_URL" ] || [ -z "$BACKEND_URL" ]; then
    echo -e "${RED}❌ Không thể cập nhật .env vì thiếu tunnels${NC}"
    exit 1
fi

# Hiển thị URLs
echo -e "${GREEN}✅ Frontend URL: ${FRONTEND_URL}${NC}"
echo -e "${GREEN}✅ Backend URL: ${BACKEND_URL}${NC}"

# Tìm file .env
ENV_FILE=".env"
if [ ! -f "$ENV_FILE" ]; then
    ENV_FILE="../.env"
fi

if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}❌ Không tìm thấy file .env${NC}"
    exit 1
fi

echo -e "${GREEN}📝 Đang cập nhật file: $ENV_FILE${NC}"

# Backup file .env
cp "$ENV_FILE" "${ENV_FILE}.bak.$(date +%Y%m%d_%H%M%S)"
echo -e "${GREEN}💾 Đã backup .env file${NC}"

# Cập nhật VNPAY_RETURN_URL
if grep -q "VNPAY_RETURN_URL=" "$ENV_FILE"; then
    # macOS sử dụng sed khác với Linux
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|VNPAY_RETURN_URL=.*|VNPAY_RETURN_URL=${FRONTEND_URL}/payment/vnpay/return|" "$ENV_FILE"
    else
        sed -i "s|VNPAY_RETURN_URL=.*|VNPAY_RETURN_URL=${FRONTEND_URL}/payment/vnpay/return|" "$ENV_FILE"
    fi
    echo -e "${GREEN}✅ Đã cập nhật VNPAY_RETURN_URL${NC}"
else
    echo "VNPAY_RETURN_URL=${FRONTEND_URL}/payment/vnpay/return" >> "$ENV_FILE"
    echo -e "${GREEN}✅ Đã thêm VNPAY_RETURN_URL${NC}"
fi

# Cập nhật VNPAY_IPN_URL
if grep -q "VNPAY_IPN_URL=" "$ENV_FILE"; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|VNPAY_IPN_URL=.*|VNPAY_IPN_URL=${BACKEND_URL}/api/payment/vnpay/ipn|" "$ENV_FILE"
    else
        sed -i "s|VNPAY_IPN_URL=.*|VNPAY_IPN_URL=${BACKEND_URL}/api/payment/vnpay/ipn|" "$ENV_FILE"
    fi
    echo -e "${GREEN}✅ Đã cập nhật VNPAY_IPN_URL${NC}"
else
    echo "VNPAY_IPN_URL=${BACKEND_URL}/api/payment/vnpay/ipn" >> "$ENV_FILE"
    echo -e "${GREEN}✅ Đã thêm VNPAY_IPN_URL${NC}"
fi

echo -e "${GREEN}✨ Hoàn tất! Hãy khởi động lại backend để áp dụng thay đổi.${NC}"


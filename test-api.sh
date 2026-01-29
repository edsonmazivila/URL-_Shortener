#!/bin/bash

CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

BASE_URL="http://localhost:8080"

echo -e "${CYAN}URL Shortener API Test Suite${NC}"
echo "========================================"
echo ""

echo -e "${YELLOW}1. Testing Health Check Endpoint${NC}"
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n1)
BODY=$(echo "$HEALTH_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
    echo "Response: $BODY"
else
    echo -e "${RED}✗ Health check failed (HTTP $HTTP_CODE)${NC}"
    echo "Response: $BODY"
    exit 1
fi

echo ""
echo -e "${YELLOW}2. Creating Short URL (GitHub)${NC}"
CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/urls" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com"}')
HTTP_CODE=$(echo "$CREATE_RESPONSE" | tail -n1)
BODY=$(echo "$CREATE_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 201 ]; then
    echo -e "${GREEN}✓ Short URL created successfully${NC}"
    echo "Response: $BODY"
    SHORT_CODE=$(echo "$BODY" | grep -o '"short_code":"[^"]*' | sed 's/"short_code":"//')
    SHORT_URL=$(echo "$BODY" | grep -o '"short_url":"[^"]*' | sed 's/"short_url":"//')
    echo "Short Code: $SHORT_CODE"
else
    echo -e "${RED}✗ Failed to create short URL (HTTP $HTTP_CODE)${NC}"
    echo "Response: $BODY"
    exit 1
fi

echo ""
echo -e "${YELLOW}3. Creating Short URL with Custom Code${NC}"
CUSTOM_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/urls" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://golang.org", "custom_code": "golang"}')
HTTP_CODE=$(echo "$CUSTOM_RESPONSE" | tail -n1)
BODY=$(echo "$CUSTOM_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 201 ]; then
    echo -e "${GREEN}✓ Short URL with custom code created${NC}"
    echo "Response: $BODY"
else
    echo -e "${RED}✗ Failed to create custom short URL (HTTP $HTTP_CODE)${NC}"
    echo "Response: $BODY"
fi

echo ""
echo -e "${YELLOW}4. Creating Short URL with Expiration (1 hour)${NC}"
EXPIRE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/urls" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "ttl": 3600}')
HTTP_CODE=$(echo "$EXPIRE_RESPONSE" | tail -n1)
BODY=$(echo "$EXPIRE_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 201 ]; then
    echo -e "${GREEN}✓ Expiring short URL created${NC}"
    echo "Response: $BODY"
else
    echo -e "${RED}✗ Failed to create expiring URL (HTTP $HTTP_CODE)${NC}"
    echo "Response: $BODY"
fi

echo ""
echo -e "${YELLOW}5. Getting URL Metadata${NC}"
METADATA_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/urls/$SHORT_CODE")
HTTP_CODE=$(echo "$METADATA_RESPONSE" | tail -n1)
BODY=$(echo "$METADATA_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✓ Retrieved URL metadata${NC}"
    echo "Response: $BODY"
else
    echo -e "${RED}✗ Failed to retrieve metadata (HTTP $HTTP_CODE)${NC}"
    echo "Response: $BODY"
fi

echo ""
echo -e "${YELLOW}6. Testing Redirect${NC}"
REDIRECT_RESPONSE=$(curl -s -w "\n%{http_code}" -L "$BASE_URL/$SHORT_CODE" -o /dev/null)
HTTP_CODE=$(echo "$REDIRECT_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✓ Redirect successful${NC}"
else
    echo -e "${RED}✗ Redirect failed (HTTP $HTTP_CODE)${NC}"
fi

echo ""
echo -e "${YELLOW}7. Verifying Access Count Increment${NC}"
METADATA_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/urls/$SHORT_CODE")
HTTP_CODE=$(echo "$METADATA_RESPONSE" | tail -n1)
BODY=$(echo "$METADATA_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    ACCESS_COUNT=$(echo "$BODY" | grep -o '"access_count":[0-9]*' | sed 's/"access_count"://')
    if [ "$ACCESS_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓ Access count incremented (count: $ACCESS_COUNT)${NC}"
    else
        echo -e "${YELLOW}⚠ Access count not incremented${NC}"
    fi
else
    echo -e "${RED}✗ Failed to verify access count${NC}"
fi

echo ""
echo -e "${YELLOW}8. Listing URLs${NC}"
LIST_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/urls?limit=5")
HTTP_CODE=$(echo "$LIST_RESPONSE" | tail -n1)
BODY=$(echo "$LIST_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✓ Retrieved URL list${NC}"
    TOTAL=$(echo "$BODY" | grep -o '"total":[0-9]*' | sed 's/"total"://')
    echo "Total URLs: $TOTAL"
else
    echo -e "${RED}✗ Failed to list URLs (HTTP $HTTP_CODE)${NC}"
    echo "Response: $BODY"
fi

echo ""
echo -e "${YELLOW}9. Testing Invalid URL${NC}"
INVALID_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/urls" \
  -H "Content-Type: application/json" \
  -d '{"url": "not-a-valid-url"}')
HTTP_CODE=$(echo "$INVALID_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" -eq 400 ]; then
    echo -e "${GREEN}✓ Invalid URL correctly rejected${NC}"
else
    echo -e "${RED}✗ Invalid URL not rejected (HTTP $HTTP_CODE)${NC}"
fi

echo ""
echo -e "${YELLOW}10. Testing Duplicate Custom Code${NC}"
DUPLICATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/urls" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.org", "custom_code": "golang"}')
HTTP_CODE=$(echo "$DUPLICATE_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" -eq 409 ]; then
    echo -e "${GREEN}✓ Duplicate custom code correctly rejected${NC}"
else
    echo -e "${RED}✗ Duplicate not handled properly (HTTP $HTTP_CODE)${NC}"
fi

echo ""
echo -e "${YELLOW}11. Testing Non-existent URL${NC}"
NOTFOUND_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/urls/nonexistent123")
HTTP_CODE=$(echo "$NOTFOUND_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" -eq 404 ]; then
    echo -e "${GREEN}✓ Non-existent URL correctly returns 404${NC}"
else
    echo -e "${RED}✗ Wrong status code for non-existent URL (HTTP $HTTP_CODE)${NC}"
fi

echo ""
echo "========================================"
echo -e "${CYAN}Test Suite Completed!${NC}"
echo ""
echo "Summary of created URLs:"
echo "  - $SHORT_URL → https://github.com"
echo "  - $BASE_URL/golang → https://golang.org"
echo ""
echo "To delete a URL:"
echo "  curl -X DELETE $BASE_URL/api/urls/$SHORT_CODE"

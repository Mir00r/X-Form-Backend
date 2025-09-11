#!/bin/bash

echo "🧪 Testing Enhanced API Gateway with Swagger Documentation"
echo "========================================================="

BASE_URL="http://localhost:8085"

echo ""
echo "1. Testing Health Endpoint:"
echo "GET $BASE_URL/health"
curl -s -w "\nHTTP Status: %{http_code}\n" "$BASE_URL/health"

echo ""
echo "2. Testing Metrics Endpoint:"
echo "GET $BASE_URL/metrics"
curl -s -w "\nHTTP Status: %{http_code}\n" "$BASE_URL/metrics"

echo ""
echo "3. Testing Gateway Info Endpoint:"
echo "GET $BASE_URL/"
curl -s -w "\nHTTP Status: %{http_code}\n" "$BASE_URL/"

echo ""
echo "4. Testing API v1 Health Endpoint:"
echo "GET $BASE_URL/api/v1/health"
curl -s -w "\nHTTP Status: %{http_code}\n" "$BASE_URL/api/v1/health"

echo ""
echo "5. Testing Swagger Documentation:"
echo "GET $BASE_URL/swagger/index.html"
SWAGGER_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/swagger/index.html")
echo "HTTP Status: $SWAGGER_RESPONSE"

if [ "$SWAGGER_RESPONSE" = "200" ]; then
    echo "✅ Swagger documentation is accessible!"
    echo "📚 Visit: $BASE_URL/swagger/index.html"
else
    echo "❌ Swagger documentation is not accessible"
fi

echo ""
echo "6. Testing Swagger JSON spec:"
echo "GET $BASE_URL/swagger/doc.json"
SWAGGER_JSON_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/swagger/doc.json")
echo "HTTP Status: $SWAGGER_JSON_RESPONSE"

echo ""
echo "========================================================="
echo "🎉 API Gateway Testing Complete!"
echo "📚 Swagger UI: $BASE_URL/swagger/index.html"
echo "📄 OpenAPI Spec: $BASE_URL/swagger/doc.json"

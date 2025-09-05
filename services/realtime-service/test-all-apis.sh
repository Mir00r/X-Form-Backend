#!/bin/bash
# X-Form Realtime Service - API Testing Script

echo "🚀 Testing X-Form Realtime Service APIs"
echo "========================================"

BASE_URL="http://localhost:8002"

echo ""
echo "📊 1. Testing Health Endpoints..."
echo "--------------------------------"

echo "Basic Health Check:"
curl -s "$BASE_URL/health" | jq '.' || echo "❌ Health check failed"

echo ""
echo "Detailed Health Check:"
curl -s "$BASE_URL/health/detailed" | jq '.' || echo "❌ Detailed health check failed"

echo ""
echo "Liveness Probe:"
curl -s "$BASE_URL/health/live" | jq '.' || echo "❌ Liveness probe failed"

echo ""
echo "Readiness Probe:"
curl -s "$BASE_URL/health/ready" | jq '.' || echo "❌ Readiness probe failed"

echo ""
echo "🔌 2. Testing WebSocket Management..."
echo "-----------------------------------"

echo "WebSocket Connection Info:"
curl -s "$BASE_URL/api/websocket/info" | jq '.' || echo "❌ WebSocket info failed"

echo ""
echo "Active Connections:"
curl -s "$BASE_URL/api/websocket/connections" | jq '.' || echo "❌ Active connections failed"

echo ""
echo "⚡ 3. Testing Real-time Endpoints..."
echo "----------------------------------"

echo "Notify Form Update:"
curl -s -X POST "$BASE_URL/api/realtime/notify/form123" \
  -H "Content-Type: application/json" \
  -d '{"type": "form_updated", "data": {"title": "Test Form"}}' | jq '.' || echo "❌ Form notification failed"

echo ""
echo "Broadcast Response:"
curl -s -X POST "$BASE_URL/api/realtime/response/form123" \
  -H "Content-Type: application/json" \
  -d '{"responseId": "resp123", "data": {"question1": "Test Answer"}}' | jq '.' || echo "❌ Response broadcast failed"

echo ""
echo "Update Form Status:"
curl -s -X POST "$BASE_URL/api/realtime/status/form123" \
  -H "Content-Type: application/json" \
  -d '{"status": "published", "message": "Form is now live"}' | jq '.' || echo "❌ Status update failed"

echo ""
echo "Global Broadcast:"
curl -s -X POST "$BASE_URL/api/realtime/broadcast" \
  -H "Content-Type: application/json" \
  -d '{"type": "system_announcement", "message": "System maintenance in 1 hour"}' | jq '.' || echo "❌ Global broadcast failed"

echo ""
echo "Get Metrics:"
curl -s "$BASE_URL/api/realtime/metrics" | jq '.' || echo "❌ Metrics failed"

echo ""
echo "📖 4. Testing Documentation..."
echo "-----------------------------"

echo "Swagger UI (returns HTML):"
curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api-docs"
if [ $? -eq 0 ]; then
    echo "✅ Swagger UI accessible"
else
    echo "❌ Swagger UI not accessible"
fi

echo ""
echo "OpenAPI JSON:"
curl -s "$BASE_URL/api-docs.json" | jq '.info.title' || echo "❌ OpenAPI JSON failed"

echo ""
echo "🎯 Testing Complete!"
echo "==================="
echo ""
echo "🌐 Access Points:"
echo "- Service: $BASE_URL"
echo "- Swagger UI: $BASE_URL/api-docs"
echo "- WebSocket Demo: $BASE_URL/demo/websocket-test.html"
echo ""
echo "💡 To test WebSocket functionality:"
echo "   1. Open: $BASE_URL/demo/websocket-test.html"
echo "   2. Click 'Connect' button"
echo "   3. Try subscribing to forms and sending responses"

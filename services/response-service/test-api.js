/**
 * API Test Script
 * Tests all the enhanced Response Service endpoints
 */

const axios = require('axios');

const BASE_URL = 'http://localhost:3002/api/v1';

async function testAPI() {
  console.log('🚀 Testing Enhanced Response Service API...\n');

  try {
    // Test 1: Health Check
    console.log('1. Testing Health Check...');
    const healthResponse = await axios.get(`${BASE_URL}/health`);
    console.log('✅ Health Check:', healthResponse.data);
    console.log('');

    // Test 2: Get all responses
    console.log('2. Testing Get All Responses...');
    const responsesResponse = await axios.get(`${BASE_URL}/responses`);
    console.log('✅ Get Responses:', {
      status: responsesResponse.status,
      count: responsesResponse.data.data?.length || 0,
      success: responsesResponse.data.success
    });
    console.log('');

    // Test 3: Create a response
    console.log('3. Testing Create Response...');
    const newResponse = {
      formId: 'test-form-123',
      respondentId: 'user-456',
      responses: {
        question1: 'Test Answer 1',
        question2: 'Test Answer 2'
      },
      metadata: {
        browser: 'test-browser',
        platform: 'test-platform'
      }
    };

    const createResponse = await axios.post(`${BASE_URL}/responses`, newResponse);
    console.log('✅ Create Response:', {
      status: createResponse.status,
      responseId: createResponse.data.data?.id,
      success: createResponse.data.success
    });
    console.log('');

    // Test 4: Analytics
    console.log('4. Testing Analytics...');
    const analyticsResponse = await axios.get(`${BASE_URL}/analytics/summary`);
    console.log('✅ Analytics:', {
      status: analyticsResponse.status,
      success: analyticsResponse.data.success
    });
    console.log('');

    console.log('🎉 All API tests completed successfully!');
    console.log('\n📖 Swagger Documentation: http://localhost:3002/api-docs');

  } catch (error) {
    console.error('❌ API Test Failed:', {
      endpoint: error.config?.url,
      status: error.response?.status,
      message: error.response?.data?.message || error.message
    });
  }
}

// Run tests
testAPI();

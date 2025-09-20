"""
Test configuration for Analytics Service
"""
import pytest
import os
import sys
from unittest.mock import AsyncMock, MagicMock

# Add the app directory to the Python path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'app'))

# Mock external dependencies for testing
class MockRedis:
    def __init__(self, *args, **kwargs):
        self.data = {}
    
    def get(self, key):
        return self.data.get(key)
    
    def setex(self, key, ttl, value):
        self.data[key] = value
        return True
    
    def delete(self, key):
        return self.data.pop(key, None) is not None
    
    def keys(self, pattern):
        return [k for k in self.data.keys() if pattern.replace('*', '') in k]
    
    def ping(self):
        return True
    
    def info(self, section=None):
        return {
            'connected_clients': 1,
            'used_memory_human': '1MB',
            'uptime_in_seconds': 3600,
            'keyspace_hits': 100,
            'keyspace_misses': 10
        }

# Mock BigQuery client
class MockBigQueryClient:
    def __init__(self, *args, **kwargs):
        pass
    
    def dataset(self, dataset_id):
        return MagicMock()
    
    def get_dataset(self, dataset_ref):
        return MagicMock()
    
    def create_dataset(self, dataset):
        return MagicMock()
    
    def get_table(self, table_ref):
        return MagicMock()
    
    def create_table(self, table):
        return MagicMock()
    
    def query(self, query, job_config=None):
        job = MagicMock()
        job.result.return_value = []
        return job
    
    def insert_rows_json(self, table, rows):
        return []

# Set up test environment
os.environ.update({
    'ENVIRONMENT': 'test',
    'BIGQUERY_PROJECT_ID': 'test-project',
    'BIGQUERY_DATASET_ID': 'test_dataset',
    'REDIS_HOST': 'localhost',
    'REDIS_PORT': '6379',
    'JWT_SECRET': 'test-secret',
    'CACHE_TTL': '300',
    'LOG_LEVEL': 'DEBUG'
})

# Mock modules
import unittest.mock
sys.modules['redis'] = unittest.mock.MagicMock()
sys.modules['google.cloud.bigquery'] = unittest.mock.MagicMock()
sys.modules['google.cloud.exceptions'] = unittest.mock.MagicMock()
sys.modules['plotly.graph_objects'] = unittest.mock.MagicMock()
sys.modules['plotly.express'] = unittest.mock.MagicMock()
sys.modules['plotly.utils'] = unittest.mock.MagicMock()
sys.modules['pandas'] = unittest.mock.MagicMock()
sys.modules['jwt'] = unittest.mock.MagicMock()
sys.modules['fastapi'] = unittest.mock.MagicMock()
sys.modules['uvicorn'] = unittest.mock.MagicMock()

@pytest.fixture
def mock_redis():
    return MockRedis()

@pytest.fixture
def mock_bigquery():
    return MockBigQueryClient()

@pytest.fixture
def sample_form_data():
    return {
        'form_id': 'test-form-123',
        'title': 'Test Form',
        'total_responses': 100,
        'completed_responses': 85,
        'partial_responses': 15,
        'completion_rate': 85.0
    }

@pytest.fixture
def sample_question_data():
    return {
        'question_id': 'q1',
        'question_type': 'multiple_choice',
        'total_responses': 100,
        'answered_responses': 90,
        'distribution': [
            {'value': 'Option A', 'count': 45, 'percentage': 50.0},
            {'value': 'Option B', 'count': 30, 'percentage': 33.3},
            {'value': 'Option C', 'count': 15, 'percentage': 16.7}
        ]
    }

/**
 * Health Controller for Response Service
 * Provides comprehensive health checks for the service and its dependencies
 */

const { createSuccessResponse, createErrorResponse } = require('../dto/response-dtos');
const logger = require('../utils/logger');

// Service start time for uptime calculation
const serviceStartTime = Date.now();

/**
 * Check database connectivity
 */
async function checkDatabase() {
  const startTime = Date.now();
  
  try {
    // Mock database check - replace with actual database ping
    await new Promise(resolve => setTimeout(resolve, Math.random() * 50)); // Simulate DB latency
    
    const responseTime = Date.now() - startTime;
    
    return {
      status: 'healthy',
      responseTime,
      details: {
        type: 'mock-database',
        connection: 'active',
        lastCheck: new Date().toISOString()
      }
    };
  } catch (error) {
    const responseTime = Date.now() - startTime;
    
    return {
      status: 'unhealthy',
      responseTime,
      error: error.message,
      details: {
        type: 'mock-database',
        connection: 'failed',
        lastCheck: new Date().toISOString()
      }
    };
  }
}

/**
 * Check form service connectivity
 */
async function checkFormService() {
  const startTime = Date.now();
  
  try {
    // Mock external service check
    await new Promise(resolve => setTimeout(resolve, Math.random() * 100)); // Simulate API call
    
    const responseTime = Date.now() - startTime;
    
    // Simulate occasional service failures
    if (Math.random() > 0.95) {
      throw new Error('Form service temporarily unavailable');
    }
    
    return {
      status: 'healthy',
      responseTime,
      details: {
        service: 'form-service',
        endpoint: process.env.FORM_SERVICE_URL || 'http://localhost:3001',
        lastCheck: new Date().toISOString()
      }
    };
  } catch (error) {
    const responseTime = Date.now() - startTime;
    
    return {
      status: 'unhealthy',
      responseTime,
      error: error.message,
      details: {
        service: 'form-service',
        endpoint: process.env.FORM_SERVICE_URL || 'http://localhost:3001',
        lastCheck: new Date().toISOString()
      }
    };
  }
}

/**
 * Check file storage service
 */
async function checkFileStorage() {
  const startTime = Date.now();
  
  try {
    // Mock file storage check
    await new Promise(resolve => setTimeout(resolve, Math.random() * 30));
    
    const responseTime = Date.now() - startTime;
    
    return {
      status: 'healthy',
      responseTime,
      details: {
        service: 'file-storage',
        provider: process.env.FILE_STORAGE_PROVIDER || 'local',
        lastCheck: new Date().toISOString()
      }
    };
  } catch (error) {
    const responseTime = Date.now() - startTime;
    
    return {
      status: 'unhealthy',
      responseTime,
      error: error.message,
      details: {
        service: 'file-storage',
        provider: process.env.FILE_STORAGE_PROVIDER || 'local',
        lastCheck: new Date().toISOString()
      }
    };
  }
}

/**
 * Check system resources
 */
function checkSystemResources() {
  const memoryUsage = process.memoryUsage();
  const cpuUsage = process.cpuUsage();
  
  // Calculate memory usage percentage (assuming 512MB container limit)
  const memoryLimitMB = parseInt(process.env.MEMORY_LIMIT_MB) || 512;
  const memoryUsedMB = memoryUsage.heapUsed / 1024 / 1024;
  const memoryPercentage = (memoryUsedMB / memoryLimitMB) * 100;
  
  // Determine health based on resource usage
  const isHealthy = memoryPercentage < 90; // Consider unhealthy if using more than 90% memory
  
  return {
    status: isHealthy ? 'healthy' : 'unhealthy',
    details: {
      memory: {
        used: Math.round(memoryUsedMB * 100) / 100,
        limit: memoryLimitMB,
        percentage: Math.round(memoryPercentage * 100) / 100
      },
      cpu: {
        user: cpuUsage.user,
        system: cpuUsage.system
      },
      uptime: process.uptime(),
      lastCheck: new Date().toISOString()
    }
  };
}

/**
 * Get comprehensive health status
 */
const getHealth = async (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const startTime = Date.now();
  
  try {
    logger.debug('Health check requested', { correlationId });
    
    // Run all health checks in parallel
    const [databaseHealth, formServiceHealth, fileStorageHealth, systemHealth] = await Promise.all([
      checkDatabase(),
      checkFormService(),
      checkFileStorage(),
      Promise.resolve(checkSystemResources())
    ]);
    
    // Determine overall health status
    const allChecks = [databaseHealth, formServiceHealth, fileStorageHealth, systemHealth];
    const healthyChecks = allChecks.filter(check => check.status === 'healthy').length;
    const totalChecks = allChecks.length;
    
    let overallStatus;
    if (healthyChecks === totalChecks) {
      overallStatus = 'healthy';
    } else if (healthyChecks >= totalChecks * 0.5) {
      overallStatus = 'degraded';
    } else {
      overallStatus = 'unhealthy';
    }
    
    const healthData = {
      status: overallStatus,
      timestamp: new Date().toISOString(),
      version: process.env.SERVICE_VERSION || '1.0.0',
      uptime: formatUptime(Date.now() - serviceStartTime),
      checks: {
        database: databaseHealth,
        externalServices: {
          formService: formServiceHealth,
          fileStorage: fileStorageHealth
        },
        system: systemHealth
      },
      summary: {
        total: totalChecks,
        healthy: healthyChecks,
        unhealthy: totalChecks - healthyChecks
      }
    };
    
    const duration = Date.now() - startTime;
    const statusCode = overallStatus === 'healthy' ? 200 : 503;
    
    logger.info('Health check completed', {
      correlationId,
      status: overallStatus,
      duration,
      healthyChecks,
      totalChecks
    });
    
    res.status(statusCode).json(
      createSuccessResponse(
        healthData,
        `Service is ${overallStatus}`,
        correlationId
      )
    );
    
  } catch (error) {
    const duration = Date.now() - startTime;
    
    logger.error('Health check failed', {
      correlationId,
      error: error.message,
      duration
    });
    
    res.status(503).json(
      createErrorResponse(
        'HEALTH_CHECK_FAILED',
        'Unable to perform health check',
        { error: error.message },
        correlationId
      )
    );
  }
};

/**
 * Get readiness status (for Kubernetes readiness probe)
 */
const getReadiness = async (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  
  try {
    logger.debug('Readiness check requested', { correlationId });
    
    // Check critical dependencies for readiness
    const [databaseHealth, formServiceHealth] = await Promise.all([
      checkDatabase(),
      checkFormService()
    ]);
    
    const isReady = databaseHealth.status === 'healthy' && formServiceHealth.status === 'healthy';
    const statusCode = isReady ? 200 : 503;
    
    const readinessData = {
      ready: isReady,
      timestamp: new Date().toISOString(),
      checks: {
        database: databaseHealth.status,
        formService: formServiceHealth.status
      }
    };
    
    logger.debug('Readiness check completed', {
      correlationId,
      ready: isReady
    });
    
    if (isReady) {
      res.status(statusCode).json(
        createSuccessResponse(
          readinessData,
          'Service is ready',
          correlationId
        )
      );
    } else {
      res.status(statusCode).json(
        createErrorResponse(
          'SERVICE_NOT_READY',
          'Service is not ready to accept traffic',
          readinessData,
          correlationId
        )
      );
    }
    
  } catch (error) {
    logger.error('Readiness check failed', {
      correlationId,
      error: error.message
    });
    
    res.status(503).json(
      createErrorResponse(
        'READINESS_CHECK_FAILED',
        'Unable to perform readiness check',
        { error: error.message },
        correlationId
      )
    );
  }
};

/**
 * Get liveness status (for Kubernetes liveness probe)
 */
const getLiveness = async (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  
  try {
    logger.debug('Liveness check requested', { correlationId });
    
    // Simple liveness check - service is alive if it can respond
    const livenessData = {
      alive: true,
      timestamp: new Date().toISOString(),
      uptime: process.uptime(),
      pid: process.pid
    };
    
    logger.debug('Liveness check completed', {
      correlationId,
      alive: true
    });
    
    res.json(
      createSuccessResponse(
        livenessData,
        'Service is alive',
        correlationId
      )
    );
    
  } catch (error) {
    logger.error('Liveness check failed', {
      correlationId,
      error: error.message
    });
    
    res.status(503).json(
      createErrorResponse(
        'LIVENESS_CHECK_FAILED',
        'Service liveness check failed',
        { error: error.message },
        correlationId
      )
    );
  }
};

/**
 * Format uptime in human-readable format
 */
function formatUptime(uptimeMs) {
  const seconds = Math.floor(uptimeMs / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  
  if (days > 0) {
    return `${days}d ${hours % 24}h ${minutes % 60}m ${seconds % 60}s`;
  } else if (hours > 0) {
    return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
  } else if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  } else {
    return `${seconds}s`;
  }
}

module.exports = {
  getHealth,
  getReadiness,
  getLiveness
};

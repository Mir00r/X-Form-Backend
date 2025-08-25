// Comprehensive Health Check Service for Microservices
// Monitors service health and dependencies following best practices

import { Pool } from 'pg';
import { CircuitBreakerFactory } from '../resilience/circuit-breaker';
import { logger } from '../logging/structured-logger';
import {
  HealthCheckResponseDTO,
  DependencyHealthDTO,
  ServiceMetricsDTO,
} from '../../interface/dto/auth-dtos';

export interface HealthCheckDependency {
  name: string;
  type: 'DATABASE' | 'CACHE' | 'EMAIL' | 'EXTERNAL_API';
  check: () => Promise<{ healthy: boolean; responseTime?: number; error?: string }>;
  timeout?: number;
  critical?: boolean;
}

export class HealthCheckService {
  private dependencies: HealthCheckDependency[] = [];
  private startTime = Date.now();
  private requestMetrics = {
    total: 0,
    errors: 0,
    totalResponseTime: 0,
  };

  constructor(private dbPool?: Pool) {
    this.registerDefaultDependencies();
  }

  registerDependency(dependency: HealthCheckDependency): void {
    this.dependencies.push(dependency);
  }

  async performHealthCheck(): Promise<HealthCheckResponseDTO> {
    const start = Date.now();
    const dependencyResults: DependencyHealthDTO[] = [];
    let overallStatus: 'HEALTHY' | 'UNHEALTHY' | 'DEGRADED' = 'HEALTHY';

    // Check all dependencies in parallel
    const dependencyChecks = this.dependencies.map(async (dep): Promise<DependencyHealthDTO> => {
      const depStart = Date.now();
      try {
        const result = await this.checkDependencyWithTimeout(dep);
        const responseTime = Date.now() - depStart;

        const health: DependencyHealthDTO = {
          name: dep.name,
          type: dep.type,
          status: result.healthy ? 'HEALTHY' : 'UNHEALTHY',
          responseTime,
          lastChecked: new Date().toISOString(),
          error: result.error,
        };

        // Determine overall status
        if (!result.healthy) {
          if (dep.critical !== false) {
            overallStatus = 'UNHEALTHY';
          } else if (overallStatus === 'HEALTHY') {
            overallStatus = 'DEGRADED';
          }
        }

        return health;
      } catch (error) {
        const health: DependencyHealthDTO = {
          name: dep.name,
          type: dep.type,
          status: 'UNHEALTHY',
          responseTime: Date.now() - depStart,
          lastChecked: new Date().toISOString(),
          error: error instanceof Error ? error.message : 'Unknown error',
        };

        if (dep.critical !== false) {
          overallStatus = 'UNHEALTHY';
        } else if (overallStatus === 'HEALTHY') {
          overallStatus = 'DEGRADED';
        }

        return health;
      }
    });

    const results = await Promise.all(dependencyChecks);
    dependencyResults.push(...results);

    const metrics = this.getServiceMetrics();
    const uptime = (Date.now() - this.startTime) / 1000;

    const healthCheck: HealthCheckResponseDTO = {
      service: 'auth-service',
      version: process.env.npm_package_version || '1.0.0',
      status: overallStatus,
      uptime,
      timestamp: new Date().toISOString(),
      environment: process.env.NODE_ENV || 'development',
      dependencies: dependencyResults,
      metrics,
    };

    // Log health check results
    logger.info('Health check completed', {
      operation: 'health_check',
      responseTime: Date.now() - start,
      metadata: {
        status: overallStatus,
        dependencyCount: dependencyResults.length,
        unhealthyDependencies: dependencyResults.filter(d => d.status === 'UNHEALTHY').length,
      },
    });

    return healthCheck;
  }

  async isHealthy(): Promise<boolean> {
    try {
      const health = await this.performHealthCheck();
      return health.status !== 'UNHEALTHY';
    } catch (error) {
      logger.error('Health check failed', error as Error);
      return false;
    }
  }

  async getReadiness(): Promise<boolean> {
    // Check only critical dependencies for readiness
    const criticalDeps = this.dependencies.filter(dep => dep.critical !== false);
    
    for (const dep of criticalDeps) {
      try {
        const result = await this.checkDependencyWithTimeout(dep);
        if (!result.healthy) {
          return false;
        }
      } catch (error) {
        return false;
      }
    }

    return true;
  }

  async getLiveness(): Promise<boolean> {
    // Simple liveness check - ensure service is responsive
    try {
      const memory = process.memoryUsage();
      const memoryUsagePercent = (memory.heapUsed / memory.heapTotal) * 100;
      
      // Consider unhealthy if memory usage is above 90%
      if (memoryUsagePercent > 90) {
        logger.warn('High memory usage detected', {
          operation: 'liveness_check',
          metadata: { memoryUsagePercent },
        });
        return false;
      }

      return true;
    } catch (error) {
      return false;
    }
  }

  private async checkDependencyWithTimeout(
    dep: HealthCheckDependency
  ): Promise<{ healthy: boolean; responseTime?: number; error?: string }> {
    const timeout = dep.timeout || 5000;
    
    return Promise.race([
      dep.check(),
      new Promise<{ healthy: boolean; error: string }>((_, reject) =>
        setTimeout(() => reject(new Error(`Health check timeout for ${dep.name}`)), timeout)
      ),
    ]);
  }

  private registerDefaultDependencies(): void {
    // Database health check
    if (this.dbPool) {
      this.registerDependency({
        name: 'postgresql',
        type: 'DATABASE',
        critical: true,
        timeout: 3000,
        check: async () => {
          try {
            const start = Date.now();
            const client = await this.dbPool!.connect();
            await client.query('SELECT 1');
            client.release();
            return {
              healthy: true,
              responseTime: Date.now() - start,
            };
          } catch (error) {
            return {
              healthy: false,
              error: error instanceof Error ? error.message : 'Database connection failed',
            };
          }
        },
      });
    }

    // Circuit breaker health checks
    this.registerDependency({
      name: 'circuit-breakers',
      type: 'EXTERNAL_API',
      critical: false,
      check: async () => {
        try {
          const metrics = CircuitBreakerFactory.getAllMetrics();
          const openCircuits = Object.values(metrics).filter(m => m.state === 'OPEN').length;
          
          return {
            healthy: openCircuits === 0,
            error: openCircuits > 0 ? `${openCircuits} circuit breakers are open` : undefined,
          };
        } catch (error) {
          return {
            healthy: false,
            error: 'Circuit breaker health check failed',
          };
        }
      },
    });

    // Email service health check (mock for now)
    this.registerDependency({
      name: 'email-service',
      type: 'EMAIL',
      critical: false,
      timeout: 5000,
      check: async () => {
        try {
          // In a real implementation, this would test email service connectivity
          // For now, we'll simulate a health check
          const isHealthy = process.env.EMAIL_SERVICE_ENABLED !== 'false';
          return {
            healthy: isHealthy,
            error: !isHealthy ? 'Email service is disabled' : undefined,
          };
        } catch (error) {
          return {
            healthy: false,
            error: 'Email service health check failed',
          };
        }
      },
    });
  }

  private getServiceMetrics(): ServiceMetricsDTO {
    const memory = process.memoryUsage();
    const cpuUsage = process.cpuUsage();
    
    // Calculate CPU usage percentage (this is a simplified calculation)
    const cpuPercent = ((cpuUsage.user + cpuUsage.system) / 1000000) / (process.uptime() * 1000) * 100;

    const metrics: ServiceMetricsDTO = {
      requestCount: this.requestMetrics.total,
      errorRate: this.requestMetrics.total > 0 ? this.requestMetrics.errors / this.requestMetrics.total : 0,
      averageResponseTime: this.requestMetrics.total > 0 ? 
        this.requestMetrics.totalResponseTime / this.requestMetrics.total : 0,
      activeConnections: this.dbPool?.totalCount || 0,
      memoryUsage: {
        used: Math.round(memory.heapUsed / 1024 / 1024), // MB
        free: Math.round((memory.heapTotal - memory.heapUsed) / 1024 / 1024), // MB
        total: Math.round(memory.heapTotal / 1024 / 1024), // MB
        percentage: Math.round((memory.heapUsed / memory.heapTotal) * 100),
      },
      cpuUsage: Math.round(cpuPercent * 100) / 100,
    };

    return metrics;
  }

  // Methods to update metrics (to be called by middleware)
  incrementRequestCount(): void {
    this.requestMetrics.total++;
  }

  incrementErrorCount(): void {
    this.requestMetrics.errors++;
  }

  addResponseTime(time: number): void {
    this.requestMetrics.totalResponseTime += time;
  }

  getMetricsSummary(): {
    requestCount: number;
    errorRate: number;
    averageResponseTime: number;
  } {
    return {
      requestCount: this.requestMetrics.total,
      errorRate: this.requestMetrics.total > 0 ? this.requestMetrics.errors / this.requestMetrics.total : 0,
      averageResponseTime: this.requestMetrics.total > 0 ? 
        this.requestMetrics.totalResponseTime / this.requestMetrics.total : 0,
    };
  }
}

// Singleton instance
let healthCheckService: HealthCheckService;

export const createHealthCheckService = (dbPool?: Pool): HealthCheckService => {
  if (!healthCheckService) {
    healthCheckService = new HealthCheckService(dbPool);
  }
  return healthCheckService;
};

export const getHealthCheckService = (): HealthCheckService => {
  if (!healthCheckService) {
    throw new Error('Health check service not initialized. Call createHealthCheckService first.');
  }
  return healthCheckService;
};

// Express middleware for metrics collection
export const metricsMiddleware = (req: any, res: any, next: any): void => {
  const start = Date.now();
  
  // Increment request count
  healthCheckService?.incrementRequestCount();

  // Override res.end to capture response time and errors
  const originalEnd = res.end;
  res.end = function(chunk: any, encoding: any) {
    res.end = originalEnd;
    res.end(chunk, encoding);
    
    const responseTime = Date.now() - start;
    healthCheckService?.addResponseTime(responseTime);
    
    // Increment error count for 4xx and 5xx responses
    if (res.statusCode >= 400) {
      healthCheckService?.incrementErrorCount();
    }
  };

  next();
};

// Circuit Breaker Implementation for Fault Tolerance
// Following microservices best practices for resilience

export enum CircuitBreakerState {
  CLOSED = 'CLOSED',
  OPEN = 'OPEN',
  HALF_OPEN = 'HALF_OPEN',
}

export interface CircuitBreakerConfig {
  failureThreshold: number;
  resetTimeout: number;
  monitoringPeriod: number;
  expectedErrors: string[];
  name: string;
}

export interface CircuitBreakerMetrics {
  totalRequests: number;
  failedRequests: number;
  successfulRequests: number;
  averageResponseTime: number;
  lastFailureTime?: Date;
  lastSuccessTime?: Date;
  state: CircuitBreakerState;
  stateChangedAt: Date;
}

export class CircuitBreaker {
  private state: CircuitBreakerState = CircuitBreakerState.CLOSED;
  private failureCount = 0;
  private nextAttempt = Date.now();
  private metrics: CircuitBreakerMetrics;
  private readonly config: CircuitBreakerConfig;

  constructor(config: CircuitBreakerConfig) {
    this.config = config;
    this.metrics = {
      totalRequests: 0,
      failedRequests: 0,
      successfulRequests: 0,
      averageResponseTime: 0,
      state: this.state,
      stateChangedAt: new Date(),
    };
  }

  async execute<T>(operation: () => Promise<T>): Promise<T> {
    if (this.state === CircuitBreakerState.OPEN) {
      if (this.nextAttempt > Date.now()) {
        throw new CircuitBreakerError(
          `Circuit breaker is OPEN for ${this.config.name}. Next attempt at ${new Date(this.nextAttempt).toISOString()}`
        );
      }
      this.setState(CircuitBreakerState.HALF_OPEN);
    }

    const startTime = Date.now();
    this.metrics.totalRequests++;

    try {
      const result = await operation();
      this.onSuccess(Date.now() - startTime);
      return result;
    } catch (error) {
      this.onFailure(error as Error, Date.now() - startTime);
      throw error;
    }
  }

  private onSuccess(responseTime: number): void {
    this.failureCount = 0;
    this.metrics.successfulRequests++;
    this.metrics.lastSuccessTime = new Date();
    this.updateAverageResponseTime(responseTime);

    if (this.state === CircuitBreakerState.HALF_OPEN) {
      this.setState(CircuitBreakerState.CLOSED);
    }
  }

  private onFailure(error: Error, responseTime: number): void {
    this.metrics.failedRequests++;
    this.metrics.lastFailureTime = new Date();
    this.updateAverageResponseTime(responseTime);

    if (this.isExpectedError(error)) {
      // Don't count expected errors towards circuit breaker failures
      return;
    }

    this.failureCount++;

    if (this.failureCount >= this.config.failureThreshold) {
      this.setState(CircuitBreakerState.OPEN);
      this.nextAttempt = Date.now() + this.config.resetTimeout;
    }
  }

  private isExpectedError(error: Error): boolean {
    return this.config.expectedErrors.some(expectedError =>
      error.name === expectedError || error.message.includes(expectedError)
    );
  }

  private setState(newState: CircuitBreakerState): void {
    if (this.state !== newState) {
      const previousState = this.state;
      this.state = newState;
      this.metrics.state = newState;
      this.metrics.stateChangedAt = new Date();
      
      console.log(
        `Circuit breaker ${this.config.name} state changed from ${previousState} to ${newState}`
      );
    }
  }

  private updateAverageResponseTime(responseTime: number): void {
    const totalRequests = this.metrics.totalRequests;
    this.metrics.averageResponseTime = 
      ((this.metrics.averageResponseTime * (totalRequests - 1)) + responseTime) / totalRequests;
  }

  getMetrics(): CircuitBreakerMetrics {
    return { ...this.metrics };
  }

  getState(): CircuitBreakerState {
    return this.state;
  }

  reset(): void {
    this.state = CircuitBreakerState.CLOSED;
    this.failureCount = 0;
    this.nextAttempt = Date.now();
    this.metrics = {
      totalRequests: 0,
      failedRequests: 0,
      successfulRequests: 0,
      averageResponseTime: 0,
      state: this.state,
      stateChangedAt: new Date(),
    };
  }
}

export class CircuitBreakerError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'CircuitBreakerError';
  }
}

// Circuit Breaker Factory for creating configured instances
export class CircuitBreakerFactory {
  private static instances = new Map<string, CircuitBreaker>();

  static create(name: string, config?: Partial<CircuitBreakerConfig>): CircuitBreaker {
    if (this.instances.has(name)) {
      return this.instances.get(name)!;
    }

    const defaultConfig: CircuitBreakerConfig = {
      failureThreshold: 5,
      resetTimeout: 60000, // 1 minute
      monitoringPeriod: 30000, // 30 seconds
      expectedErrors: ['ValidationError', 'AuthenticationError'],
      name,
    };

    const finalConfig = { ...defaultConfig, ...config, name };
    const circuitBreaker = new CircuitBreaker(finalConfig);
    
    this.instances.set(name, circuitBreaker);
    return circuitBreaker;
  }

  static get(name: string): CircuitBreaker | undefined {
    return this.instances.get(name);
  }

  static getAllMetrics(): Record<string, CircuitBreakerMetrics> {
    const metrics: Record<string, CircuitBreakerMetrics> = {};
    this.instances.forEach((breaker, name) => {
      metrics[name] = breaker.getMetrics();
    });
    return metrics;
  }

  static reset(name?: string): void {
    if (name) {
      const breaker = this.instances.get(name);
      if (breaker) {
        breaker.reset();
      }
    } else {
      this.instances.forEach(breaker => breaker.reset());
    }
  }
}

// Decorators for easy circuit breaker application
export function withCircuitBreaker(name: string, config?: Partial<CircuitBreakerConfig>) {
  return function (target: any, propertyName: string, descriptor: PropertyDescriptor) {
    const method = descriptor.value;
    const circuitBreaker = CircuitBreakerFactory.create(name, config);

    descriptor.value = async function (...args: any[]) {
      return circuitBreaker.execute(() => method.apply(this, args));
    };

    return descriptor;
  };
}

// Circuit breakers for common services
export const databaseCircuitBreaker = CircuitBreakerFactory.create('database', {
  failureThreshold: 3,
  resetTimeout: 30000,
  expectedErrors: ['ConnectionError', 'TimeoutError'],
});

export const emailCircuitBreaker = CircuitBreakerFactory.create('email', {
  failureThreshold: 5,
  resetTimeout: 120000,
  expectedErrors: ['TemplateError', 'ValidationError'],
});

export const tokenServiceCircuitBreaker = CircuitBreakerFactory.create('tokenService', {
  failureThreshold: 10,
  resetTimeout: 60000,
  expectedErrors: ['TokenExpiredError', 'InvalidTokenError'],
});

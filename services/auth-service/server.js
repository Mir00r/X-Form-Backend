require('dotenv').config();

const app = require('./app');
const { initializeDatabase } = require('./config/database');

// Configuration
const PORT = process.env.AUTH_SERVICE_PORT || 3001;
const HOST = process.env.HOST || '0.0.0.0';

// Validate required environment variables
const requiredEnvVars = [
  'DATABASE_URL',
  'JWT_SECRET',
  'JWT_REFRESH_SECRET'
];

const missingEnvVars = requiredEnvVars.filter(envVar => !process.env[envVar]);

if (missingEnvVars.length > 0) {
  console.error('Missing required environment variables:', missingEnvVars);
  process.exit(1);
}

// Optional but recommended environment variables
const optionalEnvVars = {
  'GOOGLE_CLIENT_ID': 'Google OAuth will be disabled',
  'GOOGLE_CLIENT_SECRET': 'Google OAuth will be disabled',
  'FRONTEND_URL': 'Will default to http://localhost:3000',
  'REDIS_URL': 'Rate limiting will use memory store'
};

Object.entries(optionalEnvVars).forEach(([envVar, warning]) => {
  if (!process.env[envVar]) {
    console.warn(`Warning: ${envVar} not set. ${warning}`);
  }
});

// Start server
async function startServer() {
  try {
    // Initialize database connection
    console.log('Initializing database connection...');
    await initializeDatabase();
    console.log('Database connection established successfully');

    // Start HTTP server
    const server = app.listen(PORT, HOST, () => {
      console.log(`🚀 Auth Service running on ${HOST}:${PORT}`);
      console.log(`📖 API Documentation: http://${HOST}:${PORT}/auth/health`);
      console.log(`🌍 Environment: ${process.env.NODE_ENV || 'development'}`);
      console.log(`📊 Health Check: http://${HOST}:${PORT}/health`);
      
      // Log configured features
      const features = [];
      if (process.env.GOOGLE_CLIENT_ID) features.push('Google OAuth');
      if (process.env.REDIS_URL) features.push('Redis Rate Limiting');
      else features.push('Memory Rate Limiting');
      
      console.log(`🔧 Enabled features: ${features.join(', ')}`);
    });

    // Handle server errors
    server.on('error', (error) => {
      if (error.code === 'EADDRINUSE') {
        console.error(`❌ Port ${PORT} is already in use`);
        process.exit(1);
      } else {
        console.error('❌ Server error:', error);
        process.exit(1);
      }
    });

    // Graceful shutdown
    const gracefulShutdown = async (signal) => {
      console.log(`\n📴 ${signal} received, starting graceful shutdown...`);
      
      server.close(async () => {
        console.log('✅ HTTP server closed');
        
        try {
          // Close database connections
          const { getPool } = require('./config/database');
          const pool = getPool();
          if (pool) {
            await pool.end();
            console.log('✅ Database connections closed');
          }
          
          console.log('✅ Graceful shutdown completed');
          process.exit(0);
        } catch (error) {
          console.error('❌ Error during shutdown:', error);
          process.exit(1);
        }
      });

      // Force shutdown after 30 seconds
      setTimeout(() => {
        console.error('❌ Forced shutdown due to timeout');
        process.exit(1);
      }, 30000);
    };

    // Register shutdown handlers
    process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
    process.on('SIGINT', () => gracefulShutdown('SIGINT'));

  } catch (error) {
    console.error('❌ Failed to start server:', error);
    process.exit(1);
  }
}

// Start the server
startServer();

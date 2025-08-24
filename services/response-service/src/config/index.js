const config = {
  server: {
    port: process.env.RESPONSE_SERVICE_PORT || 3002,
    host: process.env.HOST || '0.0.0.0',
    environment: process.env.NODE_ENV || 'development',
  },
  
  firebase: {
    projectId: process.env.FIREBASE_PROJECT_ID || 'xform-backend-dev',
    keyFile: process.env.FIREBASE_KEY_FILE || './firebase-key.json',
    databaseURL: process.env.FIREBASE_DATABASE_URL,
  },
  
  jwt: {
    secret: process.env.JWT_SECRET || 'your-jwt-secret-key',
    expiresIn: process.env.JWT_EXPIRES_IN || '24h',
  },
  
  kafka: {
    clientId: 'response-service',
    brokers: process.env.KAFKA_BROKERS ? 
      process.env.KAFKA_BROKERS.split(',') : 
      ['localhost:9092'],
    topics: {
      formUpdated: 'form.updated',
      formDeleted: 'form.deleted',
      responseSubmitted: 'response.submitted',
    },
  },
  
  googleSheets: {
    credentials: {
      type: process.env.GOOGLE_ACCOUNT_TYPE || 'service_account',
      project_id: process.env.GOOGLE_PROJECT_ID,
      private_key_id: process.env.GOOGLE_PRIVATE_KEY_ID,
      private_key: process.env.GOOGLE_PRIVATE_KEY?.replace(/\\n/g, '\n'),
      client_email: process.env.GOOGLE_CLIENT_EMAIL,
      client_id: process.env.GOOGLE_CLIENT_ID,
      auth_uri: 'https://accounts.google.com/o/oauth2/auth',
      token_uri: 'https://oauth2.googleapis.com/token',
      auth_provider_x509_cert_url: 'https://www.googleapis.com/oauth2/v1/certs',
      client_x509_cert_url: process.env.GOOGLE_CLIENT_CERT_URL,
    },
    scopes: [
      'https://www.googleapis.com/auth/spreadsheets',
      'https://www.googleapis.com/auth/drive.file',
    ],
  },
  
  formService: {
    baseUrl: process.env.FORM_SERVICE_URL || 'http://localhost:8082',
    timeout: parseInt(process.env.FORM_SERVICE_TIMEOUT) || 5000,
  },
  
  rateLimit: {
    windowMs: parseInt(process.env.RATE_LIMIT_WINDOW_MS) || 15 * 60 * 1000, // 15 minutes
    max: parseInt(process.env.RATE_LIMIT_MAX) || 100, // limit each IP to 100 requests per windowMs
    message: 'Too many requests from this IP, please try again later.',
  },
  
  websocket: {
    cors: {
      origin: process.env.WEBSOCKET_CORS_ORIGIN || '*',
      methods: ['GET', 'POST'],
      credentials: true,
    },
  },
  
  logging: {
    level: process.env.LOG_LEVEL || 'info',
    format: process.env.LOG_FORMAT || 'combined',
  },
  
  export: {
    maxRecords: parseInt(process.env.MAX_EXPORT_RECORDS) || 10000,
    csvDelimiter: process.env.CSV_DELIMITER || ',',
    csvEncoding: process.env.CSV_ENCODING || 'utf8',
  },
  
  validation: {
    maxResponseSize: parseInt(process.env.MAX_RESPONSE_SIZE) || 1024 * 1024, // 1MB
    maxFileSize: parseInt(process.env.MAX_FILE_SIZE) || 10 * 1024 * 1024, // 10MB
    allowedFileTypes: process.env.ALLOWED_FILE_TYPES?.split(',') || [
      'image/jpeg',
      'image/png',
      'image/gif',
      'application/pdf',
      'text/plain',
      'application/msword',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    ],
  },
};

module.exports = config;

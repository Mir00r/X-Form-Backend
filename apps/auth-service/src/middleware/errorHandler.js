const errorHandler = (err, req, res, next) => {
  console.error('Error:', err);

  // Default error response
  let status = err.status || err.statusCode || 500;
  let message = err.message || 'Internal server error';
  let error = process.env.NODE_ENV === 'production' ? 'Internal server error' : err.message;

  // Database errors
  if (err.code === '23505') { // PostgreSQL unique violation
    status = 409;
    message = 'Resource already exists';
    error = 'Duplicate entry';
  }

  if (err.code === '23503') { // PostgreSQL foreign key violation
    status = 400;
    message = 'Invalid reference';
    error = 'Referenced resource does not exist';
  }

  if (err.code === '23502') { // PostgreSQL not null violation
    status = 400;
    message = 'Missing required field';
    error = 'Required field cannot be null';
  }

  // JWT errors
  if (err.name === 'JsonWebTokenError') {
    status = 401;
    message = 'Invalid token';
    error = 'Authentication failed';
  }

  if (err.name === 'TokenExpiredError') {
    status = 401;
    message = 'Token expired';
    error = 'Please login again';
  }

  // Validation errors
  if (err.name === 'ValidationError') {
    status = 400;
    message = 'Validation failed';
    error = err.message;
  }

  res.status(status).json({
    error: error,
    message: message,
    ...(process.env.NODE_ENV === 'development' && { stack: err.stack })
  });
};

module.exports = {
  errorHandler
};

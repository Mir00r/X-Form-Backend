/**
 * Analytics Controller for Response Service
 * Provides comprehensive analytics and insights for form responses
 */

const { createSuccessResponse, createErrorResponse } = require('../dto/response-dtos');
const { NotFoundError } = require('../middleware/errorHandler');
const logger = require('../utils/logger');

// Mock database operations (replace with actual database integration)
const mockDatabase = {
  responses: new Map(),
  forms: new Map()
};

/**
 * Get analytics for a specific form
 */
const getFormAnalytics = async (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const { formId } = req.params;
  const startTime = Date.now();
  
  try {
    logger.info('Generating form analytics', {
      correlationId,
      formId
    });

    // Verify form exists
    const form = mockDatabase.forms.get(formId);
    if (!form) {
      throw new NotFoundError('Form not found');
    }

    // Get all responses for this form
    const formResponses = Array.from(mockDatabase.responses.values())
      .filter(response => response.formId === formId);

    // Calculate basic statistics
    const totalResponses = formResponses.length;
    const completedResponses = formResponses.filter(r => r.status === 'completed').length;
    const draftResponses = formResponses.filter(r => r.status === 'draft').length;
    const partialResponses = formResponses.filter(r => r.status === 'partial').length;
    const archivedResponses = formResponses.filter(r => r.status === 'archived').length;

    // Calculate completion rate
    const completionRate = totalResponses > 0 ? (completedResponses / totalResponses) * 100 : 0;

    // Calculate average completion time (for responses that have time data)
    const responsesWithTime = formResponses.filter(r => r.metadata?.timeSpent);
    const averageCompletionTime = responsesWithTime.length > 0
      ? responsesWithTime.reduce((sum, r) => sum + r.metadata.timeSpent, 0) / responsesWithTime.length
      : null;

    // Get responses by date for trend analysis
    const responsesByDate = getResponsesByDate(formResponses);

    // Get last response timestamp
    const lastResponse = formResponses.length > 0
      ? formResponses.reduce((latest, response) => 
          new Date(response.submittedAt) > new Date(latest.submittedAt) ? response : latest
        ).submittedAt
      : null;

    // Generate question-level analytics
    const questionAnalytics = generateQuestionAnalytics(formResponses);

    // Generate respondent analytics
    const respondentAnalytics = generateRespondentAnalytics(formResponses);

    // Generate device/browser analytics from metadata
    const deviceAnalytics = generateDeviceAnalytics(formResponses);

    // Generate geographic analytics (if IP data available)
    const geographicAnalytics = generateGeographicAnalytics(formResponses);

    const analyticsData = {
      formId,
      formTitle: form.title,
      totalResponses,
      completedResponses,
      draftResponses,
      partialResponses,
      archivedResponses,
      completionRate: Math.round(completionRate * 100) / 100,
      averageCompletionTime,
      responsesByDate,
      lastResponse,
      questionAnalytics,
      respondentAnalytics,
      deviceAnalytics,
      geographicAnalytics,
      trends: {
        daily: calculateDailyTrend(responsesByDate),
        weekly: calculateWeeklyTrend(responsesByDate),
        monthly: calculateMonthlyTrend(responsesByDate)
      },
      generatedAt: new Date().toISOString()
    };

    const duration = Date.now() - startTime;

    logger.logBusiness('ANALYTICS_GENERATED', 'form', formId, {
      totalResponses,
      completionRate,
      generationTime: duration
    }, { correlationId });

    logger.info('Form analytics generated successfully', {
      correlationId,
      formId,
      totalResponses,
      duration
    });

    res.json(
      createSuccessResponse(
        analyticsData,
        'Analytics retrieved successfully',
        correlationId
      )
    );

  } catch (error) {
    const duration = Date.now() - startTime;
    
    logger.error('Failed to generate form analytics', {
      correlationId,
      formId,
      error: error.message,
      duration
    });

    throw error;
  }
};

/**
 * Get responses grouped by date for trend analysis
 */
function getResponsesByDate(responses) {
  const responsesByDate = {};
  
  responses.forEach(response => {
    const date = response.submittedAt.split('T')[0]; // Get YYYY-MM-DD
    
    if (!responsesByDate[date]) {
      responsesByDate[date] = {
        date,
        count: 0,
        completed: 0,
        draft: 0,
        partial: 0
      };
    }
    
    responsesByDate[date].count++;
    responsesByDate[date][response.status]++;
  });
  
  // Convert to array and sort by date
  return Object.values(responsesByDate).sort((a, b) => a.date.localeCompare(b.date));
}

/**
 * Generate analytics for individual questions
 */
function generateQuestionAnalytics(responses) {
  const questionStats = {};
  
  responses.forEach(response => {
    response.responses.forEach(questionResponse => {
      const { questionId, questionType, value } = questionResponse;
      
      if (!questionStats[questionId]) {
        questionStats[questionId] = {
          questionId,
          questionType,
          totalResponses: 0,
          uniqueValues: new Set(),
          valueDistribution: {},
          averageLength: 0,
          emptyResponses: 0
        };
      }
      
      const stats = questionStats[questionId];
      stats.totalResponses++;
      
      if (!value || value === '' || (Array.isArray(value) && value.length === 0)) {
        stats.emptyResponses++;
      } else {
        const valueStr = typeof value === 'object' ? JSON.stringify(value) : String(value);
        stats.uniqueValues.add(valueStr);
        
        // Count value distribution for categorical data
        if (questionType === 'radio' || questionType === 'select' || questionType === 'checkbox') {
          const key = Array.isArray(value) ? value.join(', ') : String(value);
          stats.valueDistribution[key] = (stats.valueDistribution[key] || 0) + 1;
        }
        
        // Calculate average length for text responses
        if (questionType === 'text' || questionType === 'textarea') {
          const currentAvg = stats.averageLength;
          const currentCount = stats.totalResponses - stats.emptyResponses;
          stats.averageLength = (currentAvg * (currentCount - 1) + valueStr.length) / currentCount;
        }
      }
    });
  });
  
  // Convert Sets to arrays and finalize calculations
  Object.values(questionStats).forEach(stats => {
    stats.uniqueValues = Array.from(stats.uniqueValues);
    stats.responseRate = ((stats.totalResponses - stats.emptyResponses) / stats.totalResponses) * 100;
    stats.averageLength = Math.round(stats.averageLength * 100) / 100;
  });
  
  return Object.values(questionStats);
}

/**
 * Generate respondent analytics
 */
function generateRespondentAnalytics(responses) {
  const uniqueEmails = new Set();
  const anonymousResponses = responses.filter(r => !r.respondentEmail).length;
  let totalTimeSpent = 0;
  let responsesWithTime = 0;
  
  responses.forEach(response => {
    if (response.respondentEmail) {
      uniqueEmails.add(response.respondentEmail);
    }
    
    if (response.metadata?.timeSpent) {
      totalTimeSpent += response.metadata.timeSpent;
      responsesWithTime++;
    }
  });
  
  return {
    totalRespondents: uniqueEmails.size + anonymousResponses,
    authenticatedRespondents: uniqueEmails.size,
    anonymousResponses,
    averageTimeSpent: responsesWithTime > 0 ? Math.round(totalTimeSpent / responsesWithTime) : null,
    returnRespondents: responses.length - uniqueEmails.size - anonymousResponses // Approximate
  };
}

/**
 * Generate device and browser analytics from user agent data
 */
function generateDeviceAnalytics(responses) {
  const devices = {};
  const browsers = {};
  const operatingSystems = {};
  
  responses.forEach(response => {
    const userAgent = response.metadata?.userAgent;
    if (!userAgent) return;
    
    // Simple user agent parsing (in production, use a proper library like 'ua-parser-js')
    const ua = userAgent.toLowerCase();
    
    // Device type detection
    let deviceType = 'desktop';
    if (ua.includes('mobile') || ua.includes('android')) {
      deviceType = 'mobile';
    } else if (ua.includes('tablet') || ua.includes('ipad')) {
      deviceType = 'tablet';
    }
    devices[deviceType] = (devices[deviceType] || 0) + 1;
    
    // Browser detection
    let browser = 'unknown';
    if (ua.includes('chrome')) browser = 'chrome';
    else if (ua.includes('firefox')) browser = 'firefox';
    else if (ua.includes('safari')) browser = 'safari';
    else if (ua.includes('edge')) browser = 'edge';
    browsers[browser] = (browsers[browser] || 0) + 1;
    
    // OS detection
    let os = 'unknown';
    if (ua.includes('windows')) os = 'windows';
    else if (ua.includes('mac')) os = 'macos';
    else if (ua.includes('linux')) os = 'linux';
    else if (ua.includes('android')) os = 'android';
    else if (ua.includes('ios')) os = 'ios';
    operatingSystems[os] = (operatingSystems[os] || 0) + 1;
  });
  
  return {
    devices,
    browsers,
    operatingSystems
  };
}

/**
 * Generate geographic analytics from IP data
 */
function generateGeographicAnalytics(responses) {
  const countries = {};
  const cities = {};
  
  responses.forEach(response => {
    const ipAddress = response.metadata?.ipAddress;
    if (!ipAddress) return;
    
    // Mock geographic data (in production, use a GeoIP service)
    const mockCountries = ['United States', 'Canada', 'United Kingdom', 'Germany', 'France', 'Australia'];
    const mockCities = ['New York', 'Toronto', 'London', 'Berlin', 'Paris', 'Sydney'];
    
    const country = mockCountries[Math.floor(Math.random() * mockCountries.length)];
    const city = mockCities[Math.floor(Math.random() * mockCities.length)];
    
    countries[country] = (countries[country] || 0) + 1;
    cities[city] = (cities[city] || 0) + 1;
  });
  
  return {
    countries,
    cities
  };
}

/**
 * Calculate daily response trend
 */
function calculateDailyTrend(responsesByDate) {
  if (responsesByDate.length < 2) return null;
  
  const recent = responsesByDate.slice(-7); // Last 7 days
  const previous = responsesByDate.slice(-14, -7); // Previous 7 days
  
  const recentAvg = recent.reduce((sum, day) => sum + day.count, 0) / recent.length;
  const previousAvg = previous.length > 0 
    ? previous.reduce((sum, day) => sum + day.count, 0) / previous.length 
    : 0;
  
  const change = previousAvg > 0 ? ((recentAvg - previousAvg) / previousAvg) * 100 : 0;
  
  return {
    current: Math.round(recentAvg * 100) / 100,
    previous: Math.round(previousAvg * 100) / 100,
    change: Math.round(change * 100) / 100,
    trend: change > 5 ? 'increasing' : change < -5 ? 'decreasing' : 'stable'
  };
}

/**
 * Calculate weekly response trend
 */
function calculateWeeklyTrend(responsesByDate) {
  // Group by week and calculate trend similar to daily
  // Simplified implementation
  return calculateDailyTrend(responsesByDate);
}

/**
 * Calculate monthly response trend
 */
function calculateMonthlyTrend(responsesByDate) {
  // Group by month and calculate trend similar to daily
  // Simplified implementation
  return calculateDailyTrend(responsesByDate);
}

/**
 * Get dashboard analytics (summary across all forms for admin users)
 */
const getDashboardAnalytics = async (req, res) => {
  const correlationId = req.headers['x-correlation-id'] || req.correlationId;
  const startTime = Date.now();
  
  try {
    logger.info('Generating dashboard analytics', { correlationId });

    const allResponses = Array.from(mockDatabase.responses.values());
    const allForms = Array.from(mockDatabase.forms.values());

    // Calculate overall statistics
    const totalForms = allForms.length;
    const totalResponses = allResponses.length;
    const activeForms = allForms.filter(f => f.status === 'active').length;
    
    // Forms with most responses
    const formResponseCounts = {};
    allResponses.forEach(response => {
      formResponseCounts[response.formId] = (formResponseCounts[response.formId] || 0) + 1;
    });
    
    const topForms = Object.entries(formResponseCounts)
      .map(([formId, count]) => ({
        formId,
        formTitle: mockDatabase.forms.get(formId)?.title || 'Unknown',
        responseCount: count
      }))
      .sort((a, b) => b.responseCount - a.responseCount)
      .slice(0, 10);

    // Recent activity
    const recentResponses = allResponses
      .sort((a, b) => new Date(b.submittedAt) - new Date(a.submittedAt))
      .slice(0, 20)
      .map(response => ({
        id: response.id,
        formId: response.formId,
        formTitle: response.formTitle,
        submittedAt: response.submittedAt,
        status: response.status
      }));

    // Response trends
    const responsesByDate = getResponsesByDate(allResponses);

    const dashboardData = {
      summary: {
        totalForms,
        totalResponses,
        activeForms,
        avgResponsesPerForm: totalForms > 0 ? Math.round(totalResponses / totalForms * 100) / 100 : 0
      },
      topForms,
      recentResponses,
      trends: {
        daily: calculateDailyTrend(responsesByDate),
        responsesByDate: responsesByDate.slice(-30) // Last 30 days
      },
      generatedAt: new Date().toISOString()
    };

    const duration = Date.now() - startTime;

    logger.info('Dashboard analytics generated successfully', {
      correlationId,
      totalResponses,
      totalForms,
      duration
    });

    res.json(
      createSuccessResponse(
        dashboardData,
        'Dashboard analytics retrieved successfully',
        correlationId
      )
    );

  } catch (error) {
    const duration = Date.now() - startTime;
    
    logger.error('Failed to generate dashboard analytics', {
      correlationId,
      error: error.message,
      duration
    });

    throw error;
  }
};

module.exports = {
  getFormAnalytics,
  getDashboardAnalytics
};

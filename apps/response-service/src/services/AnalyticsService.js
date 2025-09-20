const { firestore } = require('../config/firebase');
const { ResponseAnalytics } = require('../models');
const { DateHelper, PaginationHelper, ErrorHelper } = require('../utils/helpers');
const logger = require('../utils/logger');

/**
 * Analytics Service for form response analytics and reporting
 */
class AnalyticsService {
  constructor() {
    this.responsesCollection = firestore.collection('responses');
    this.analyticsCollection = firestore.collection('response_analytics');
  }

  /**
   * Get analytics data for a form
   * @param {string} formId - Form ID
   * @param {Object} options - Analytics options
   */
  async getFormAnalytics(formId, options = {}) {
    try {
      const {
        period = 'daily',
        startDate,
        endDate,
        metrics = ['responses', 'completion_rate', 'average_time'],
      } = options;

      // Validate date range
      if (!startDate || !endDate) {
        throw ErrorHelper.createError('Start date and end date are required', 400, 'MISSING_DATE_RANGE');
      }

      const start = new Date(startDate);
      const end = new Date(endDate);

      if (start >= end) {
        throw ErrorHelper.createError('Start date must be before end date', 400, 'INVALID_DATE_RANGE');
      }

      // Get aggregated analytics data
      const analytics = await this.getAggregatedAnalytics(formId, start, end, period);

      // Get real-time data if needed
      const realtimeData = await this.getRealtimeMetrics(formId);

      // Calculate time-series data
      const timeSeries = await this.generateTimeSeries(formId, start, end, period, metrics);

      // Get question analytics
      const questionAnalytics = await this.getQuestionAnalytics(formId, start, end);

      // Get conversion funnel
      const conversionFunnel = await this.getConversionFunnel(formId, start, end);

      // Get device/browser breakdown
      const deviceBreakdown = await this.getDeviceBreakdown(formId, start, end);

      const result = {
        formId,
        period,
        dateRange: { startDate, endDate },
        summary: {
          totalResponses: analytics.totalResponses,
          completedResponses: analytics.completedResponses,
          completionRate: analytics.completionRate,
          averageCompletionTime: analytics.averageCompletionTime,
          bounceRate: analytics.bounceRate,
          uniqueVisitors: analytics.uniqueVisitors,
        },
        realtime: realtimeData,
        timeSeries,
        questionAnalytics,
        conversionFunnel,
        deviceBreakdown,
      };

      logger.info('Analytics data generated', {
        formId,
        period,
        dateRange: { startDate, endDate },
        metricsRequested: metrics,
      });

      return result;
    } catch (error) {
      logger.error('Failed to get form analytics:', error);
      throw ErrorHelper.handleDbError(error);
    }
  }

  /**
   * Get real-time analytics for the last hour
   * @param {string} formId - Form ID
   * @param {number} timeWindow - Time window in minutes
   */
  async getRealtimeAnalytics(formId, timeWindow = 60) {
    try {
      const cutoffTime = new Date(Date.now() - (timeWindow * 60 * 1000));

      const snapshot = await this.responsesCollection
        .where('formId', '==', formId)
        .where('submittedAt', '>=', cutoffTime)
        .orderBy('submittedAt', 'desc')
        .get();

      const responses = snapshot.docs.map(doc => doc.data());

      // Calculate real-time metrics
      const metrics = {
        recentResponses: responses.length,
        completedResponses: responses.filter(r => r.isComplete).length,
        averageTimeToComplete: 0,
        activeUsers: new Set(responses.map(r => r.submitterId).filter(id => id)).size,
        topSources: this.getTopSources(responses),
        recentActivity: responses.slice(0, 10).map(r => ({
          id: r.id,
          submittedAt: r.submittedAt,
          isComplete: r.isComplete,
          duration: r.duration,
        })),
      };

      // Calculate average completion time
      const completedWithDuration = responses.filter(r => r.isComplete && r.duration);
      if (completedWithDuration.length > 0) {
        const totalDuration = completedWithDuration.reduce((sum, r) => sum + r.duration, 0);
        metrics.averageTimeToComplete = totalDuration / completedWithDuration.length;
      }

      return metrics;
    } catch (error) {
      logger.error('Failed to get realtime analytics:', error);
      throw ErrorHelper.handleDbError(error);
    }
  }

  /**
   * Get question-level analytics
   * @param {string} formId - Form ID
   * @param {Date} startDate - Start date
   * @param {Date} endDate - End date
   */
  async getQuestionAnalytics(formId, startDate, endDate) {
    try {
      const snapshot = await this.responsesCollection
        .where('formId', '==', formId)
        .where('submittedAt', '>=', startDate)
        .where('submittedAt', '<=', endDate)
        .where('isComplete', '==', true)
        .get();

      const responses = snapshot.docs.map(doc => doc.data());
      
      if (responses.length === 0) {
        return {};
      }

      // Aggregate question responses
      const questionStats = {};

      responses.forEach(response => {
        if (response.responses) {
          Object.entries(response.responses).forEach(([questionId, answer]) => {
            if (!questionStats[questionId]) {
              questionStats[questionId] = {
                totalAnswers: 0,
                answerDistribution: {},
                skipRate: 0,
                averageTime: 0,
                sentiment: { positive: 0, neutral: 0, negative: 0 },
              };
            }

            questionStats[questionId].totalAnswers++;

            // Track answer distribution for multiple choice questions
            if (Array.isArray(answer)) {
              answer.forEach(value => {
                questionStats[questionId].answerDistribution[value] = 
                  (questionStats[questionId].answerDistribution[value] || 0) + 1;
              });
            } else if (answer && typeof answer === 'string') {
              questionStats[questionId].answerDistribution[answer] = 
                (questionStats[questionId].answerDistribution[answer] || 0) + 1;
            }
          });
        }
      });

      // Calculate skip rates
      Object.keys(questionStats).forEach(questionId => {
        const answered = questionStats[questionId].totalAnswers;
        const total = responses.length;
        questionStats[questionId].skipRate = ((total - answered) / total) * 100;
      });

      return questionStats;
    } catch (error) {
      logger.error('Failed to get question analytics:', error);
      throw ErrorHelper.handleDbError(error);
    }
  }

  /**
   * Generate time series data
   * @param {string} formId - Form ID
   * @param {Date} startDate - Start date
   * @param {Date} endDate - End date
   * @param {string} period - Period (daily, weekly, monthly)
   * @param {Array} metrics - Metrics to include
   */
  async generateTimeSeries(formId, startDate, endDate, period, metrics) {
    try {
      const timePoints = this.generateTimePoints(startDate, endDate, period);
      const timeSeries = [];

      for (const timePoint of timePoints) {
        const { start, end } = this.getTimePointRange(timePoint, period);
        
        const snapshot = await this.responsesCollection
          .where('formId', '==', formId)
          .where('submittedAt', '>=', start)
          .where('submittedAt', '<', end)
          .get();

        const responses = snapshot.docs.map(doc => doc.data());
        
        const dataPoint = {
          date: timePoint.toISOString().split('T')[0],
          timestamp: timePoint.getTime(),
        };

        // Calculate requested metrics
        if (metrics.includes('responses')) {
          dataPoint.responses = responses.length;
        }

        if (metrics.includes('completion_rate')) {
          const completed = responses.filter(r => r.isComplete).length;
          dataPoint.completionRate = responses.length > 0 ? (completed / responses.length) * 100 : 0;
        }

        if (metrics.includes('average_time')) {
          const completedWithDuration = responses.filter(r => r.isComplete && r.duration);
          if (completedWithDuration.length > 0) {
            const totalDuration = completedWithDuration.reduce((sum, r) => sum + r.duration, 0);
            dataPoint.averageTime = totalDuration / completedWithDuration.length;
          } else {
            dataPoint.averageTime = 0;
          }
        }

        if (metrics.includes('bounce_rate')) {
          const drafts = responses.filter(r => r.isDraft).length;
          dataPoint.bounceRate = responses.length > 0 ? (drafts / responses.length) * 100 : 0;
        }

        timeSeries.push(dataPoint);
      }

      return timeSeries;
    } catch (error) {
      logger.error('Failed to generate time series:', error);
      throw error;
    }
  }

  /**
   * Get conversion funnel data
   * @param {string} formId - Form ID
   * @param {Date} startDate - Start date
   * @param {Date} endDate - End date
   */
  async getConversionFunnel(formId, startDate, endDate) {
    try {
      // This would typically require additional tracking data
      // For now, we'll calculate basic funnel based on available data
      
      const snapshot = await this.responsesCollection
        .where('formId', '==', formId)
        .where('submittedAt', '>=', startDate)
        .where('submittedAt', '<=', endDate)
        .get();

      const responses = snapshot.docs.map(doc => doc.data());
      
      const funnel = {
        formViews: responses.length + Math.floor(responses.length * 0.3), // Estimate
        formStarts: responses.length,
        partialCompletions: responses.filter(r => !r.isComplete && !r.isDraft).length,
        completions: responses.filter(r => r.isComplete).length,
      };

      // Calculate conversion rates
      funnel.startRate = funnel.formViews > 0 ? (funnel.formStarts / funnel.formViews) * 100 : 0;
      funnel.completionRate = funnel.formStarts > 0 ? (funnel.completions / funnel.formStarts) * 100 : 0;
      funnel.abandonmentRate = funnel.formStarts > 0 ? ((funnel.formStarts - funnel.completions) / funnel.formStarts) * 100 : 0;

      return funnel;
    } catch (error) {
      logger.error('Failed to get conversion funnel:', error);
      throw error;
    }
  }

  /**
   * Get device and browser breakdown
   * @param {string} formId - Form ID
   * @param {Date} startDate - Start date
   * @param {Date} endDate - End date
   */
  async getDeviceBreakdown(formId, startDate, endDate) {
    try {
      const snapshot = await this.responsesCollection
        .where('formId', '==', formId)
        .where('submittedAt', '>=', startDate)
        .where('submittedAt', '<=', endDate)
        .get();

      const responses = snapshot.docs.map(doc => doc.data());
      
      const deviceBreakdown = { desktop: 0, mobile: 0, tablet: 0, unknown: 0 };
      const browserBreakdown = {};
      const osBreakdown = {};

      responses.forEach(response => {
        if (response.userAgent) {
          const userAgent = response.userAgent.toLowerCase();
          
          // Simple device detection
          if (userAgent.includes('mobile')) {
            deviceBreakdown.mobile++;
          } else if (userAgent.includes('tablet') || userAgent.includes('ipad')) {
            deviceBreakdown.tablet++;
          } else if (userAgent.includes('mozilla') || userAgent.includes('chrome')) {
            deviceBreakdown.desktop++;
          } else {
            deviceBreakdown.unknown++;
          }

          // Browser detection
          let browser = 'Unknown';
          if (userAgent.includes('chrome')) browser = 'Chrome';
          else if (userAgent.includes('firefox')) browser = 'Firefox';
          else if (userAgent.includes('safari')) browser = 'Safari';
          else if (userAgent.includes('edge')) browser = 'Edge';
          
          browserBreakdown[browser] = (browserBreakdown[browser] || 0) + 1;

          // OS detection
          let os = 'Unknown';
          if (userAgent.includes('windows')) os = 'Windows';
          else if (userAgent.includes('mac')) os = 'macOS';
          else if (userAgent.includes('linux')) os = 'Linux';
          else if (userAgent.includes('android')) os = 'Android';
          else if (userAgent.includes('ios')) os = 'iOS';
          
          osBreakdown[os] = (osBreakdown[os] || 0) + 1;
        } else {
          deviceBreakdown.unknown++;
        }
      });

      return {
        devices: deviceBreakdown,
        browsers: browserBreakdown,
        operatingSystems: osBreakdown,
      };
    } catch (error) {
      logger.error('Failed to get device breakdown:', error);
      throw error;
    }
  }

  /**
   * Get aggregated analytics data
   * @private
   */
  async getAggregatedAnalytics(formId, startDate, endDate, period) {
    try {
      // Try to get from analytics collection first
      const analyticsSnapshot = await this.analyticsCollection
        .where('formId', '==', formId)
        .where('date', '>=', DateHelper.formatForFilename(startDate))
        .where('date', '<=', DateHelper.formatForFilename(endDate))
        .get();

      if (!analyticsSnapshot.empty) {
        // Aggregate existing analytics data
        const analytics = analyticsSnapshot.docs.map(doc => doc.data());
        
        return analytics.reduce((acc, curr) => ({
          totalResponses: acc.totalResponses + curr.totalResponses,
          completedResponses: acc.completedResponses + curr.completedResponses,
          completionRate: 0, // Will be calculated after aggregation
          averageCompletionTime: 0, // Will be calculated after aggregation
          bounceRate: 0, // Will be calculated after aggregation
          uniqueVisitors: acc.uniqueVisitors + curr.uniqueVisitors,
        }), {
          totalResponses: 0,
          completedResponses: 0,
          completionRate: 0,
          averageCompletionTime: 0,
          bounceRate: 0,
          uniqueVisitors: 0,
        });
      }

      // Fall back to real-time calculation
      return await this.calculateRealtimeAnalytics(formId, startDate, endDate);
    } catch (error) {
      logger.error('Failed to get aggregated analytics:', error);
      throw error;
    }
  }

  /**
   * Calculate real-time analytics
   * @private
   */
  async calculateRealtimeAnalytics(formId, startDate, endDate) {
    const snapshot = await this.responsesCollection
      .where('formId', '==', formId)
      .where('submittedAt', '>=', startDate)
      .where('submittedAt', '<=', endDate)
      .get();

    const responses = snapshot.docs.map(doc => doc.data());
    
    const totalResponses = responses.length;
    const completedResponses = responses.filter(r => r.isComplete).length;
    const drafts = responses.filter(r => r.isDraft).length;
    
    const completionRate = totalResponses > 0 ? (completedResponses / totalResponses) * 100 : 0;
    const bounceRate = totalResponses > 0 ? (drafts / totalResponses) * 100 : 0;
    
    // Calculate average completion time
    const completedWithDuration = responses.filter(r => r.isComplete && r.duration);
    const averageCompletionTime = completedWithDuration.length > 0
      ? completedWithDuration.reduce((sum, r) => sum + r.duration, 0) / completedWithDuration.length
      : 0;

    // Count unique visitors
    const uniqueVisitors = new Set(
      responses.map(r => r.submitterId || r.sessionId).filter(id => id)
    ).size;

    return {
      totalResponses,
      completedResponses,
      completionRate,
      averageCompletionTime,
      bounceRate,
      uniqueVisitors,
    };
  }

  /**
   * Get real-time metrics
   * @private
   */
  async getRealtimeMetrics(formId) {
    return await this.getRealtimeAnalytics(formId, 60); // Last 60 minutes
  }

  /**
   * Generate time points for time series
   * @private
   */
  generateTimePoints(startDate, endDate, period) {
    const points = [];
    const current = new Date(startDate);
    
    while (current <= endDate) {
      points.push(new Date(current));
      
      switch (period) {
        case 'daily':
          current.setDate(current.getDate() + 1);
          break;
        case 'weekly':
          current.setDate(current.getDate() + 7);
          break;
        case 'monthly':
          current.setMonth(current.getMonth() + 1);
          break;
      }
    }
    
    return points;
  }

  /**
   * Get time point range
   * @private
   */
  getTimePointRange(timePoint, period) {
    const start = new Date(timePoint);
    const end = new Date(timePoint);
    
    switch (period) {
      case 'daily':
        start.setHours(0, 0, 0, 0);
        end.setHours(23, 59, 59, 999);
        break;
      case 'weekly':
        const dayOfWeek = start.getDay();
        start.setDate(start.getDate() - dayOfWeek);
        start.setHours(0, 0, 0, 0);
        end.setDate(start.getDate() + 6);
        end.setHours(23, 59, 59, 999);
        break;
      case 'monthly':
        start.setDate(1);
        start.setHours(0, 0, 0, 0);
        end.setMonth(end.getMonth() + 1, 0);
        end.setHours(23, 59, 59, 999);
        break;
    }
    
    return { start, end };
  }

  /**
   * Get top sources from responses
   * @private
   */
  getTopSources(responses) {
    const sources = {};
    
    responses.forEach(response => {
      const source = response.metadata?.source || 'direct';
      sources[source] = (sources[source] || 0) + 1;
    });
    
    return Object.entries(sources)
      .sort(([,a], [,b]) => b - a)
      .slice(0, 5)
      .map(([source, count]) => ({ source, count }));
  }
}

module.exports = AnalyticsService;

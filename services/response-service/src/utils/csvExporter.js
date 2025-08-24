const createCsvWriter = require('csv-writer').createObjectCsvWriter;
const fastCsv = require('fast-csv');
const path = require('path');
const fs = require('fs');
const { promisify } = require('util');
const { DateHelper } = require('./helpers');
const logger = require('./logger');

const writeFile = promisify(fs.writeFile);
const mkdir = promisify(fs.mkdir);

/**
 * CSV Export utilities for form responses
 */
class CsvExporter {
  constructor() {
    this.exportDir = path.join(process.cwd(), 'exports');
    this.ensureExportDir();
  }

  /**
   * Ensure export directory exists
   */
  async ensureExportDir() {
    try {
      await mkdir(this.exportDir, { recursive: true });
    } catch (error) {
      logger.error('Failed to create export directory:', error);
    }
  }

  /**
   * Export form responses to CSV
   * @param {Array} responses - Array of response objects
   * @param {Object} form - Form metadata
   * @param {Object} options - Export options
   */
  async exportResponses(responses, form, options = {}) {
    try {
      const {
        filename = this.generateFilename(form.title || 'responses'),
        includeMetadata = true,
        customHeaders = null,
      } = options;

      const filePath = path.join(this.exportDir, filename);
      const headers = this.generateHeaders(form, responses, customHeaders);
      const records = this.transformResponses(responses, form, includeMetadata);

      const csvWriter = createCsvWriter({
        path: filePath,
        header: headers,
        encoding: 'utf8',
      });

      await csvWriter.writeRecords(records);

      logger.info(`CSV export completed: ${filename}`, {
        formId: form.id,
        responseCount: responses.length,
        filePath,
      });

      return {
        filename,
        filePath,
        recordCount: records.length,
        size: await this.getFileSize(filePath),
      };
    } catch (error) {
      logger.error('CSV export failed:', error);
      throw new Error(`CSV export failed: ${error.message}`);
    }
  }

  /**
   * Export responses as stream for large datasets
   * @param {AsyncIterable} responseStream - Stream of responses
   * @param {Object} form - Form metadata
   * @param {Object} options - Export options
   */
  async exportResponsesStream(responseStream, form, options = {}) {
    try {
      const {
        filename = this.generateFilename(form.title || 'responses'),
        includeMetadata = true,
      } = options;

      const filePath = path.join(this.exportDir, filename);
      const headers = this.generateHeadersFromForm(form);

      return new Promise((resolve, reject) => {
        const csvStream = fastCsv.format({ headers: true });
        const writeStream = fs.createWriteStream(filePath);

        csvStream.pipe(writeStream);

        let recordCount = 0;

        // Write header row
        const headerRow = {};
        headers.forEach(header => {
          headerRow[header.id] = header.title;
        });

        responseStream.on('data', (response) => {
          try {
            const record = this.transformResponse(response, form, includeMetadata);
            csvStream.write(record);
            recordCount++;
          } catch (error) {
            logger.error('Error processing response for CSV:', error);
          }
        });

        responseStream.on('end', async () => {
          csvStream.end();
          const size = await this.getFileSize(filePath);
          
          resolve({
            filename,
            filePath,
            recordCount,
            size,
          });
        });

        responseStream.on('error', reject);
        writeStream.on('error', reject);
      });
    } catch (error) {
      logger.error('CSV stream export failed:', error);
      throw new Error(`CSV stream export failed: ${error.message}`);
    }
  }

  /**
   * Generate headers for CSV based on form structure
   * @param {Object} form - Form object
   * @param {Array} responses - Sample responses for dynamic headers
   * @param {Array} customHeaders - Custom header configuration
   */
  generateHeaders(form, responses = [], customHeaders = null) {
    if (customHeaders) {
      return customHeaders;
    }

    const headers = [];

    // Add metadata headers
    headers.push(
      { id: 'responseId', title: 'Response ID' },
      { id: 'submittedAt', title: 'Submitted At' },
      { id: 'submitterEmail', title: 'Submitter Email' },
      { id: 'submitterName', title: 'Submitter Name' },
      { id: 'ipAddress', title: 'IP Address' },
      { id: 'userAgent', title: 'User Agent' },
      { id: 'duration', title: 'Duration (seconds)' }
    );

    // Add question headers from form structure
    if (form.questions && Array.isArray(form.questions)) {
      form.questions.forEach(question => {
        const baseHeader = {
          id: `question_${question.id}`,
          title: this.sanitizeHeaderTitle(question.question || question.title || question.label),
        };

        headers.push(baseHeader);

        // Add additional columns for specific question types
        if (question.type === 'multiple_choice' && question.allowOther) {
          headers.push({
            id: `question_${question.id}_other`,
            title: `${baseHeader.title} (Other)`,
          });
        }

        if (question.type === 'file_upload') {
          headers.push({
            id: `question_${question.id}_filename`,
            title: `${baseHeader.title} (Filename)`,
          });
          headers.push({
            id: `question_${question.id}_filesize`,
            title: `${baseHeader.title} (File Size)`,
          });
        }
      });
    }

    // Add dynamic headers from responses if form structure is incomplete
    if (responses.length > 0) {
      const dynamicHeaders = this.extractDynamicHeaders(responses, headers);
      headers.push(...dynamicHeaders);
    }

    return headers;
  }

  /**
   * Generate headers from form structure only
   * @param {Object} form - Form object
   */
  generateHeadersFromForm(form) {
    return this.generateHeaders(form, [], null);
  }

  /**
   * Transform responses to CSV records
   * @param {Array} responses - Array of response objects
   * @param {Object} form - Form metadata
   * @param {boolean} includeMetadata - Whether to include metadata
   */
  transformResponses(responses, form, includeMetadata = true) {
    return responses.map(response => 
      this.transformResponse(response, form, includeMetadata)
    );
  }

  /**
   * Transform single response to CSV record
   * @param {Object} response - Response object
   * @param {Object} form - Form metadata
   * @param {boolean} includeMetadata - Whether to include metadata
   */
  transformResponse(response, form, includeMetadata = true) {
    const record = {};

    // Add metadata
    if (includeMetadata) {
      record.responseId = response.id;
      record.submittedAt = response.submittedAt ? 
        new Date(response.submittedAt.toDate()).toISOString() : '';
      record.submitterEmail = response.submitterEmail || '';
      record.submitterName = response.submitterName || '';
      record.ipAddress = response.metadata?.ipAddress || '';
      record.userAgent = response.metadata?.userAgent || '';
      record.duration = response.metadata?.duration || '';
    }

    // Add question responses
    if (response.responses && typeof response.responses === 'object') {
      Object.entries(response.responses).forEach(([questionId, answer]) => {
        const fieldKey = `question_${questionId}`;
        
        if (Array.isArray(answer)) {
          record[fieldKey] = answer.join('; ');
        } else if (typeof answer === 'object' && answer !== null) {
          // Handle complex objects (e.g., file uploads, other responses)
          if (answer.value !== undefined) {
            record[fieldKey] = answer.value;
          }
          if (answer.other) {
            record[`${fieldKey}_other`] = answer.other;
          }
          if (answer.filename) {
            record[`${fieldKey}_filename`] = answer.filename;
            record[`${fieldKey}_filesize`] = answer.filesize || '';
          }
          if (answer.text) {
            record[fieldKey] = answer.text;
          }
        } else {
          record[fieldKey] = answer || '';
        }
      });
    }

    return record;
  }

  /**
   * Extract dynamic headers from responses
   * @param {Array} responses - Array of responses
   * @param {Array} existingHeaders - Existing headers to avoid duplicates
   */
  extractDynamicHeaders(responses, existingHeaders = []) {
    const existingIds = new Set(existingHeaders.map(h => h.id));
    const dynamicHeaders = [];

    responses.forEach(response => {
      if (response.responses && typeof response.responses === 'object') {
        Object.keys(response.responses).forEach(questionId => {
          const fieldKey = `question_${questionId}`;
          if (!existingIds.has(fieldKey)) {
            dynamicHeaders.push({
              id: fieldKey,
              title: `Question ${questionId}`,
            });
            existingIds.add(fieldKey);
          }
        });
      }
    });

    return dynamicHeaders;
  }

  /**
   * Sanitize header title for CSV
   * @param {string} title - Original title
   */
  sanitizeHeaderTitle(title) {
    if (!title) return 'Untitled';
    return title
      .replace(/[^\w\s-]/g, '') // Remove special characters
      .replace(/\s+/g, ' ') // Normalize whitespace
      .trim()
      .substring(0, 100); // Limit length
  }

  /**
   * Generate filename for export
   * @param {string} baseName - Base name for file
   */
  generateFilename(baseName) {
    const sanitizedName = baseName
      .replace(/[^\w\s-]/g, '')
      .replace(/\s+/g, '_')
      .toLowerCase();
    
    const timestamp = DateHelper.formatForFilename();
    
    return `${sanitizedName}_${timestamp}.csv`;
  }

  /**
   * Get file size
   * @param {string} filePath - Path to file
   */
  async getFileSize(filePath) {
    try {
      const stats = await promisify(fs.stat)(filePath);
      return stats.size;
    } catch (error) {
      logger.error('Failed to get file size:', error);
      return 0;
    }
  }

  /**
   * Clean up old export files
   * @param {number} maxAgeHours - Maximum age in hours
   */
  async cleanupOldExports(maxAgeHours = 24) {
    try {
      const files = await promisify(fs.readdir)(this.exportDir);
      const cutoffTime = Date.now() - (maxAgeHours * 60 * 60 * 1000);

      for (const file of files) {
        const filePath = path.join(this.exportDir, file);
        const stats = await promisify(fs.stat)(filePath);
        
        if (stats.mtime.getTime() < cutoffTime) {
          await promisify(fs.unlink)(filePath);
          logger.info(`Cleaned up old export file: ${file}`);
        }
      }
    } catch (error) {
      logger.error('Failed to cleanup old exports:', error);
    }
  }
}

/**
 * CSV Import utilities for bulk operations
 */
class CsvImporter {
  /**
   * Parse CSV file and validate structure
   * @param {string} filePath - Path to CSV file
   * @param {Object} options - Import options
   */
  async importCsv(filePath, options = {}) {
    try {
      const {
        headers = true,
        delimiter = ',',
        maxRows = 10000,
        validation = null,
      } = options;

      const records = [];
      let rowCount = 0;

      return new Promise((resolve, reject) => {
        fs.createReadStream(filePath)
          .pipe(fastCsv.parse({ headers, delimiter }))
          .on('data', (row) => {
            if (rowCount >= maxRows) {
              return reject(new Error(`Maximum row limit (${maxRows}) exceeded`));
            }

            if (validation && !validation(row, rowCount)) {
              return reject(new Error(`Validation failed at row ${rowCount + 1}`));
            }

            records.push(row);
            rowCount++;
          })
          .on('end', () => {
            resolve({
              records,
              rowCount,
              headers: headers ? Object.keys(records[0] || {}) : null,
            });
          })
          .on('error', reject);
      });
    } catch (error) {
      logger.error('CSV import failed:', error);
      throw new Error(`CSV import failed: ${error.message}`);
    }
  }
}

module.exports = {
  CsvExporter,
  CsvImporter,
};

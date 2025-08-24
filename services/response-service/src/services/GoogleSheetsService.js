const { google } = require('googleapis');
const { firestore } = require('../config/firebase');
const { FormIntegration } = require('../models');
const { ErrorHelper, DateHelper } = require('../utils/helpers');
const { CsvExporter } = require('../utils/csvExporter');
const logger = require('../utils/logger');
const config = require('../config');

/**
 * Google Sheets Integration Service
 */
class GoogleSheetsService {
  constructor() {
    this.auth = new google.auth.GoogleAuth({
      keyFile: config.googleSheets.keyFile,
      scopes: ['https://www.googleapis.com/auth/spreadsheets'],
    });
    this.sheets = google.sheets({ version: 'v4', auth: this.auth });
    this.integrationsCollection = firestore.collection('form_integrations');
  }

  /**
   * Create a new Google Sheets integration
   * @param {string} formId - Form ID
   * @param {Object} config - Integration configuration
   */
  async createIntegration(formId, integrationConfig) {
    try {
      // Validate spreadsheet access
      await this.validateSpreadsheetAccess(integrationConfig.spreadsheetId);

      const integration = new FormIntegration({
        formId,
        type: 'google_sheets',
        config: integrationConfig,
        isActive: true,
      });

      const validation = integration.validate();
      if (!validation.isValid) {
        throw ErrorHelper.createError('Integration validation failed', 400, 'VALIDATION_ERROR');
      }

      // Save integration
      const docRef = await this.integrationsCollection.add(integration.toFirestore());
      integration.id = docRef.id;

      logger.info('Google Sheets integration created', {
        integrationId: integration.id,
        formId,
        spreadsheetId: integrationConfig.spreadsheetId,
      });

      return integration;
    } catch (error) {
      logger.error('Failed to create Google Sheets integration:', error);
      throw ErrorHelper.handleExternalServiceError(error, 'Google Sheets');
    }
  }

  /**
   * Export responses to Google Sheets
   * @param {string} integrationId - Integration ID
   * @param {Array} responses - Array of responses
   * @param {Object} form - Form metadata
   */
  async exportResponses(integrationId, responses, form) {
    try {
      // Get integration configuration
      const integrationDoc = await this.integrationsCollection.doc(integrationId).get();
      
      if (!integrationDoc.exists) {
        throw ErrorHelper.createError('Integration not found', 404, 'INTEGRATION_NOT_FOUND');
      }

      const integration = FormIntegration.fromFirestore(integrationDoc);
      const { spreadsheetId, worksheetName, includeTimestamp, fieldMapping } = integration.config;

      // Prepare data for export
      const exportData = await this.prepareExportData(responses, form, integration.config);

      // Clear existing data if configured
      if (integration.config.clearExisting) {
        await this.clearWorksheet(spreadsheetId, worksheetName);
      }

      // Write headers
      await this.writeHeaders(spreadsheetId, worksheetName, exportData.headers);

      // Write data
      const result = await this.writeData(spreadsheetId, worksheetName, exportData.rows);

      // Update integration status
      integration.updateSyncStatus('success');
      await this.integrationsCollection.doc(integrationId).update(integration.toFirestore());

      logger.info('Responses exported to Google Sheets', {
        integrationId,
        spreadsheetId,
        worksheetName,
        responseCount: responses.length,
        rowsWritten: result.updatedRows,
      });

      return {
        success: true,
        spreadsheetId,
        worksheetName,
        rowsWritten: result.updatedRows,
        responseCount: responses.length,
      };
    } catch (error) {
      // Update integration status with error
      try {
        const integrationDoc = await this.integrationsCollection.doc(integrationId).get();
        if (integrationDoc.exists) {
          const integration = FormIntegration.fromFirestore(integrationDoc);
          integration.updateSyncStatus('failed', error);
          await this.integrationsCollection.doc(integrationId).update(integration.toFirestore());
        }
      } catch (updateError) {
        logger.error('Failed to update integration status:', updateError);
      }

      logger.error('Failed to export to Google Sheets:', error);
      throw ErrorHelper.handleExternalServiceError(error, 'Google Sheets');
    }
  }

  /**
   * Sync single response to Google Sheets
   * @param {string} integrationId - Integration ID
   * @param {Object} response - Response object
   * @param {Object} form - Form metadata
   */
  async syncResponse(integrationId, response, form) {
    try {
      const integrationDoc = await this.integrationsCollection.doc(integrationId).get();
      
      if (!integrationDoc.exists) {
        throw ErrorHelper.createError('Integration not found', 404, 'INTEGRATION_NOT_FOUND');
      }

      const integration = FormIntegration.fromFirestore(integrationDoc);
      const { spreadsheetId, worksheetName } = integration.config;

      // Prepare single response data
      const exportData = await this.prepareExportData([response], form, integration.config);
      
      if (exportData.rows.length === 0) {
        return { success: true, message: 'No data to sync' };
      }

      // Append single row
      const result = await this.appendData(spreadsheetId, worksheetName, [exportData.rows[0]]);

      logger.info('Response synced to Google Sheets', {
        integrationId,
        responseId: response.id,
        spreadsheetId,
        worksheetName,
      });

      return {
        success: true,
        responseId: response.id,
        rowsAdded: result.updatedRows,
      };
    } catch (error) {
      logger.error('Failed to sync response to Google Sheets:', error);
      throw ErrorHelper.handleExternalServiceError(error, 'Google Sheets');
    }
  }

  /**
   * Validate spreadsheet access
   * @param {string} spreadsheetId - Spreadsheet ID
   */
  async validateSpreadsheetAccess(spreadsheetId) {
    try {
      const response = await this.sheets.spreadsheets.get({
        spreadsheetId,
      });

      return {
        success: true,
        title: response.data.properties.title,
        sheets: response.data.sheets.map(sheet => ({
          title: sheet.properties.title,
          sheetId: sheet.properties.sheetId,
        })),
      };
    } catch (error) {
      if (error.code === 404) {
        throw ErrorHelper.createError('Spreadsheet not found', 404, 'SPREADSHEET_NOT_FOUND');
      }
      if (error.code === 403) {
        throw ErrorHelper.createError('Access denied to spreadsheet', 403, 'SPREADSHEET_ACCESS_DENIED');
      }
      throw error;
    }
  }

  /**
   * Create a new worksheet
   * @param {string} spreadsheetId - Spreadsheet ID
   * @param {string} worksheetName - Worksheet name
   */
  async createWorksheet(spreadsheetId, worksheetName) {
    try {
      const response = await this.sheets.spreadsheets.batchUpdate({
        spreadsheetId,
        requestBody: {
          requests: [
            {
              addSheet: {
                properties: {
                  title: worksheetName,
                },
              },
            },
          ],
        },
      });

      const newSheet = response.data.replies[0].addSheet;
      
      logger.info('Worksheet created', {
        spreadsheetId,
        worksheetName,
        sheetId: newSheet.properties.sheetId,
      });

      return newSheet;
    } catch (error) {
      throw ErrorHelper.handleExternalServiceError(error, 'Google Sheets');
    }
  }

  /**
   * Clear worksheet data
   * @param {string} spreadsheetId - Spreadsheet ID
   * @param {string} worksheetName - Worksheet name
   */
  async clearWorksheet(spreadsheetId, worksheetName) {
    try {
      await this.sheets.spreadsheets.values.clear({
        spreadsheetId,
        range: `${worksheetName}!A:Z`,
      });

      logger.info('Worksheet cleared', {
        spreadsheetId,
        worksheetName,
      });
    } catch (error) {
      throw ErrorHelper.handleExternalServiceError(error, 'Google Sheets');
    }
  }

  /**
   * Write headers to worksheet
   * @param {string} spreadsheetId - Spreadsheet ID
   * @param {string} worksheetName - Worksheet name
   * @param {Array} headers - Header row
   */
  async writeHeaders(spreadsheetId, worksheetName, headers) {
    try {
      const response = await this.sheets.spreadsheets.values.update({
        spreadsheetId,
        range: `${worksheetName}!A1:${this.columnToLetter(headers.length)}1`,
        valueInputOption: 'RAW',
        requestBody: {
          values: [headers],
        },
      });

      // Format header row
      await this.sheets.spreadsheets.batchUpdate({
        spreadsheetId,
        requestBody: {
          requests: [
            {
              repeatCell: {
                range: {
                  sheetId: await this.getSheetId(spreadsheetId, worksheetName),
                  startRowIndex: 0,
                  endRowIndex: 1,
                  startColumnIndex: 0,
                  endColumnIndex: headers.length,
                },
                cell: {
                  userEnteredFormat: {
                    backgroundColor: { red: 0.9, green: 0.9, blue: 0.9 },
                    textFormat: { bold: true },
                  },
                },
                fields: 'userEnteredFormat(backgroundColor,textFormat)',
              },
            },
          ],
        },
      });

      return response.data;
    } catch (error) {
      throw ErrorHelper.handleExternalServiceError(error, 'Google Sheets');
    }
  }

  /**
   * Write data to worksheet
   * @param {string} spreadsheetId - Spreadsheet ID
   * @param {string} worksheetName - Worksheet name
   * @param {Array} rows - Data rows
   */
  async writeData(spreadsheetId, worksheetName, rows) {
    try {
      if (rows.length === 0) {
        return { updatedRows: 0 };
      }

      const response = await this.sheets.spreadsheets.values.update({
        spreadsheetId,
        range: `${worksheetName}!A2:${this.columnToLetter(rows[0].length)}${rows.length + 1}`,
        valueInputOption: 'RAW',
        requestBody: {
          values: rows,
        },
      });

      return response.data;
    } catch (error) {
      throw ErrorHelper.handleExternalServiceError(error, 'Google Sheets');
    }
  }

  /**
   * Append data to worksheet
   * @param {string} spreadsheetId - Spreadsheet ID
   * @param {string} worksheetName - Worksheet name
   * @param {Array} rows - Data rows to append
   */
  async appendData(spreadsheetId, worksheetName, rows) {
    try {
      const response = await this.sheets.spreadsheets.values.append({
        spreadsheetId,
        range: `${worksheetName}!A:A`,
        valueInputOption: 'RAW',
        insertDataOption: 'INSERT_ROWS',
        requestBody: {
          values: rows,
        },
      });

      return response.data;
    } catch (error) {
      throw ErrorHelper.handleExternalServiceError(error, 'Google Sheets');
    }
  }

  /**
   * Prepare export data from responses
   * @param {Array} responses - Array of responses
   * @param {Object} form - Form metadata
   * @param {Object} config - Integration configuration
   */
  async prepareExportData(responses, form, config) {
    try {
      const csvExporter = new CsvExporter();
      
      // Generate headers based on form and configuration
      const headers = csvExporter.generateHeaders(form, responses);
      
      // Apply field mapping if configured
      let finalHeaders = headers.map(h => h.title);
      if (config.fieldMapping && Object.keys(config.fieldMapping).length > 0) {
        finalHeaders = headers.map(h => 
          config.fieldMapping[h.id] || h.title
        );
      }

      // Add timestamp column if configured
      if (config.includeTimestamp) {
        finalHeaders.unshift('Export Timestamp');
      }

      // Transform responses to rows
      const rows = responses.map(response => {
        const record = csvExporter.transformResponse(response, form, true);
        let row = headers.map(h => record[h.id] || '');
        
        // Add timestamp if configured
        if (config.includeTimestamp) {
          row.unshift(new Date().toISOString());
        }

        return row;
      });

      return { headers: finalHeaders, rows };
    } catch (error) {
      logger.error('Failed to prepare export data:', error);
      throw error;
    }
  }

  /**
   * Get sheet ID by name
   * @param {string} spreadsheetId - Spreadsheet ID
   * @param {string} worksheetName - Worksheet name
   */
  async getSheetId(spreadsheetId, worksheetName) {
    try {
      const response = await this.sheets.spreadsheets.get({
        spreadsheetId,
      });

      const sheet = response.data.sheets.find(s => 
        s.properties.title === worksheetName
      );

      if (!sheet) {
        throw ErrorHelper.createError('Worksheet not found', 404, 'WORKSHEET_NOT_FOUND');
      }

      return sheet.properties.sheetId;
    } catch (error) {
      throw ErrorHelper.handleExternalServiceError(error, 'Google Sheets');
    }
  }

  /**
   * Convert column number to letter (A, B, C, ..., AA, AB, etc.)
   * @param {number} column - Column number (1-based)
   */
  columnToLetter(column) {
    let result = '';
    while (column > 0) {
      const remainder = (column - 1) % 26;
      result = String.fromCharCode(65 + remainder) + result;
      column = Math.floor((column - 1) / 26);
    }
    return result;
  }

  /**
   * Test integration connectivity
   * @param {Object} config - Integration configuration
   */
  async testIntegration(config) {
    try {
      // Validate spreadsheet access
      const spreadsheetInfo = await this.validateSpreadsheetAccess(config.spreadsheetId);

      // Check if worksheet exists
      const worksheetExists = spreadsheetInfo.sheets.some(sheet => 
        sheet.title === config.worksheetName
      );

      if (!worksheetExists) {
        // Try to create the worksheet
        await this.createWorksheet(config.spreadsheetId, config.worksheetName);
      }

      // Write test data
      const testHeaders = ['Test Column 1', 'Test Column 2', 'Timestamp'];
      const testRow = ['Test Value 1', 'Test Value 2', new Date().toISOString()];

      await this.writeHeaders(config.spreadsheetId, config.worksheetName, testHeaders);
      await this.appendData(config.spreadsheetId, config.worksheetName, [testRow]);

      return {
        success: true,
        message: 'Integration test successful',
        spreadsheetTitle: spreadsheetInfo.title,
        worksheetName: config.worksheetName,
      };
    } catch (error) {
      return {
        success: false,
        message: error.message,
        error: error.code || 'UNKNOWN_ERROR',
      };
    }
  }
}

module.exports = GoogleSheetsService;

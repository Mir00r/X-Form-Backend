const admin = require('firebase-admin');
const config = require('./index');
const logger = require('../utils/logger');

let firestore = null;

/**
 * Initialize Firebase Admin SDK
 */
const initializeFirebase = () => {
  try {
    // Check if Firebase is already initialized
    if (admin.apps.length === 0) {
      let credential;
      
      if (config.firebase.keyFile && config.firebase.keyFile !== './firebase-key.json') {
        // Use service account key file
        credential = admin.credential.cert(require(config.firebase.keyFile));
      } else if (process.env.GOOGLE_APPLICATION_CREDENTIALS) {
        // Use Google Application Default Credentials
        credential = admin.credential.applicationDefault();
      } else if (config.googleSheets.credentials.private_key) {
        // Use environment variables
        credential = admin.credential.cert({
          projectId: config.firebase.projectId,
          clientEmail: config.googleSheets.credentials.client_email,
          privateKey: config.googleSheets.credentials.private_key,
        });
      } else {
        throw new Error('No valid Firebase credentials found');
      }

      admin.initializeApp({
        credential,
        projectId: config.firebase.projectId,
        databaseURL: config.firebase.databaseURL,
      });

      logger.info('Firebase Admin SDK initialized successfully');
    }

    firestore = admin.firestore();
    
    // Configure Firestore settings
    firestore.settings({
      timestampsInSnapshots: true,
    });

    return firestore;
  } catch (error) {
    logger.error('Failed to initialize Firebase:', error);
    throw error;
  }
};

/**
 * Get Firestore instance
 * @returns {admin.firestore.Firestore}
 */
const getFirestore = () => {
  if (!firestore) {
    initializeFirebase();
  }
  return firestore;
};

/**
 * Collection references
 */
const collections = {
  responses: 'responses',
  formIntegrations: 'form_integrations',
  responseMetadata: 'response_metadata',
};

/**
 * Get a collection reference
 * @param {string} collectionName 
 * @returns {admin.firestore.CollectionReference}
 */
const getCollection = (collectionName) => {
  const db = getFirestore();
  return db.collection(collectionName);
};

/**
 * Batch operations helper
 */
const batch = () => {
  const db = getFirestore();
  return db.batch();
};

/**
 * Transaction helper
 * @param {Function} updateFunction 
 * @returns {Promise}
 */
const runTransaction = (updateFunction) => {
  const db = getFirestore();
  return db.runTransaction(updateFunction);
};

/**
 * Timestamp helpers
 */
const timestamp = {
  now: () => admin.firestore.Timestamp.now(),
  fromDate: (date) => admin.firestore.Timestamp.fromDate(date),
  fromMillis: (milliseconds) => admin.firestore.Timestamp.fromMillis(milliseconds),
};

/**
 * Field value helpers
 */
const fieldValue = {
  serverTimestamp: () => admin.firestore.FieldValue.serverTimestamp(),
  increment: (value) => admin.firestore.FieldValue.increment(value),
  arrayUnion: (...elements) => admin.firestore.FieldValue.arrayUnion(...elements),
  arrayRemove: (...elements) => admin.firestore.FieldValue.arrayRemove(...elements),
  delete: () => admin.firestore.FieldValue.delete(),
};

module.exports = {
  initializeFirebase,
  getFirestore,
  getCollection,
  collections,
  batch,
  runTransaction,
  timestamp,
  fieldValue,
  admin,
};

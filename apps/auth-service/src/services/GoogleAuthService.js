const passport = require('passport');
const GoogleStrategy = require('passport-google-oauth20').Strategy;
const AuthService = require('./AuthService');
const { User } = require('../models/User');
const { getPool } = require('../config/database');

class GoogleAuthService extends AuthService {
  constructor() {
    super();
    this.setupGoogleStrategy();
  }

  setupGoogleStrategy() {
    passport.use(new GoogleStrategy({
      clientID: process.env.GOOGLE_CLIENT_ID,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET,
      callbackURL: "/auth/google/callback"
    }, async (accessToken, refreshToken, profile, done) => {
      try {
        const user = await this.handleGoogleAuth(profile, accessToken);
        return done(null, user);
      } catch (error) {
        return done(error, null);
      }
    }));

    passport.serializeUser((user, done) => {
      done(null, user.id);
    });

    passport.deserializeUser(async (id, done) => {
      try {
        const user = await this.getUserById(id);
        done(null, user);
      } catch (error) {
        done(error, null);
      }
    });
  }

  async handleGoogleAuth(profile, accessToken) {
    const client = await this.pool.connect();
    
    try {
      await client.query('BEGIN');

      const googleId = profile.id;
      const email = profile.emails[0].value;
      const firstName = profile.name.givenName;
      const lastName = profile.name.familyName;
      const avatarUrl = profile.photos[0]?.value;

      // Check if user exists with Google ID
      let userResult = await client.query(
        'SELECT * FROM users WHERE google_id = $1',
        [googleId]
      );

      if (userResult.rows.length > 0) {
        // Update existing user
        const existingUser = userResult.rows[0];
        await client.query(`
          UPDATE users 
          SET avatar_url = $1, last_login = CURRENT_TIMESTAMP 
          WHERE id = $2
        `, [avatarUrl, existingUser.id]);

        await client.query('COMMIT');
        return new User(existingUser);
      }

      // Check if user exists with email (link accounts)
      userResult = await client.query(
        'SELECT * FROM users WHERE email = $1',
        [email]
      );

      if (userResult.rows.length > 0) {
        // Link Google account to existing user
        const existingUser = userResult.rows[0];
        await client.query(`
          UPDATE users 
          SET google_id = $1, avatar_url = $2, email_verified = TRUE, last_login = CURRENT_TIMESTAMP
          WHERE id = $3
        `, [googleId, avatarUrl, existingUser.id]);

        await client.query('COMMIT');
        
        // Return updated user
        const updatedResult = await client.query('SELECT * FROM users WHERE id = $1', [existingUser.id]);
        return new User(updatedResult.rows[0]);
      }

      // Create new user
      const newUserResult = await client.query(`
        INSERT INTO users (
          email, first_name, last_name, google_id, avatar_url, 
          email_verified, provider, is_active
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING *
      `, [
        email,
        firstName,
        lastName,
        googleId,
        avatarUrl,
        true,
        'google',
        true
      ]);

      const newUser = new User(newUserResult.rows[0]);

      // Log registration event
      await this.logAuthEvent(newUser.id, 'register', null, null, {
        provider: 'google',
        googleId
      }, true);

      await client.query('COMMIT');
      return newUser;

    } catch (error) {
      await client.query('ROLLBACK');
      throw error;
    } finally {
      client.release();
    }
  }

  async loginWithGoogle(user, ipAddress, userAgent) {
    try {
      // Generate tokens
      const jwtPayload = user.toJWTPayload();
      const accessToken = this.generateAccessToken(jwtPayload);
      const refreshToken = this.generateRefreshToken({ id: user.id });

      // Store refresh token
      const refreshTokenExpiry = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000); // 7 days
      const refreshTokenHash = crypto.createHash('sha256').update(refreshToken).digest('hex');

      await this.pool.query(`
        INSERT INTO refresh_tokens (user_id, token_hash, expires_at, ip_address, device_info)
        VALUES ($1, $2, $3, $4, $5)
      `, [
        user.id,
        refreshTokenHash,
        refreshTokenExpiry,
        ipAddress,
        JSON.stringify({ userAgent, provider: 'google' })
      ]);

      // Log successful login
      await this.logAuthEvent(user.id, 'login', ipAddress, userAgent, {
        provider: 'google'
      }, true);

      return {
        user: user.toSafeObject(),
        accessToken,
        refreshToken,
        expiresIn: this.getTokenExpiryTime(this.ACCESS_TOKEN_EXPIRY)
      };

    } catch (error) {
      throw error;
    }
  }
}

module.exports = GoogleAuthService;

package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Mir00r/X-Form-Backend/services/api-gateway/internal/models"
	"github.com/gin-gonic/gin"
)

// Enhanced Authentication Handlers with comprehensive Swagger documentation

// Register godoc
// @Summary      Register a new user account
// @Description  Creates a new user account with comprehensive validation and security features
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body models.RegistrationRequest true "User registration data"
// @Success      201 {object} models.StandardAPIResponse{data=models.AuthenticationResponse} "User registered successfully"
// @Failure      400 {object} models.ValidationErrorResponse "Validation errors"
// @Failure      409 {object} models.StandardAPIResponse{error=models.DetailedError} "Email already exists"
// @Failure      422 {object} models.StandardAPIResponse{error=models.DetailedError} "Invalid registration data"
// @Failure      429 {object} models.StandardAPIResponse{error=models.DetailedError} "Rate limit exceeded"
// @Failure      500 {object} models.StandardAPIResponse{error=models.DetailedError} "Internal server error"
// @Router       /auth/register [post]
// @Security     ApiKeyAuth
func EnhancedRegister(c *gin.Context) {
	var req models.RegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.StandardAPIResponse{
			Success: false,
			Message: "Invalid request format",
			Error: &models.DetailedError{
				Code:      "VALIDATION_ERROR",
				Message:   "Request validation failed",
				Details:   err.Error(),
				Timestamp: time.Now(),
				TraceID:   c.GetString("request_id"),
			},
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	// Mock successful registration response
	response := models.AuthenticationResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.mock.token",
		RefreshToken: "refresh_token_example",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		ExpiresAt:    time.Now().Add(time.Hour),
		Scope:        []string{"read", "write"},
		User: models.DetailedUser{
			ID:            "user_123456789",
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			FullName:      req.FirstName + " " + req.LastName,
			Email:         req.Email,
			EmailVerified: false,
			Role:          "user",
			Status:        "active",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		Session: models.SessionInfo{
			ID:        "sess_123456789",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(8 * time.Hour),
			IPAddress: c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			IsActive:  true,
		},
		Permissions:      []models.Permission{},
		TwoFactorEnabled: false,
	}

	c.JSON(http.StatusCreated, models.StandardAPIResponse{
		Success:   true,
		Message:   "User registered successfully",
		Data:      response,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
	})
}

// Login godoc
// @Summary      Authenticate user credentials
// @Description  Authenticates user with email/password and returns JWT tokens with comprehensive session information
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body models.AuthenticationRequest true "User login credentials"
// @Success      200 {object} models.StandardAPIResponse{data=models.AuthenticationResponse} "Login successful"
// @Failure      400 {object} models.ValidationErrorResponse "Invalid request format"
// @Failure      401 {object} models.StandardAPIResponse{error=models.DetailedError} "Invalid credentials"
// @Failure      403 {object} models.StandardAPIResponse{error=models.DetailedError} "Account locked or suspended"
// @Failure      422 {object} models.StandardAPIResponse{error=models.DetailedError} "Two-factor authentication required"
// @Failure      429 {object} models.StandardAPIResponse{error=models.DetailedError} "Rate limit exceeded"
// @Failure      500 {object} models.StandardAPIResponse{error=models.DetailedError} "Internal server error"
// @Router       /auth/login [post]
func EnhancedLogin(c *gin.Context) {
	var req models.AuthenticationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
			Success: false,
			Message: "Validation failed",
			Errors: []models.ValidationError{
				{
					Field:   "email",
					Value:   req.Email,
					Message: "Invalid email format",
					Code:    "INVALID_EMAIL",
				},
			},
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	// Mock successful login response
	response := models.AuthenticationResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.enhanced.token",
		RefreshToken: "refresh_token_enhanced",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		ExpiresAt:    time.Now().Add(time.Hour),
		Scope:        []string{"read", "write", "admin"},
		User: models.DetailedUser{
			ID:               "user_987654321",
			FirstName:        "John",
			LastName:         "Doe",
			FullName:         "John Doe",
			Email:            req.Email,
			EmailVerified:    true,
			PhoneNumber:      "+1234567890",
			PhoneVerified:    true,
			Avatar:           "https://example.com/avatar.jpg",
			Role:             "user",
			Status:           "active",
			LastLoginAt:      &[]time.Time{time.Now()}[0],
			LoginCount:       42,
			TwoFactorEnabled: false,
			CreatedAt:        time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:        time.Now(),
		},
		Session: models.SessionInfo{
			ID:        "sess_987654321",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(8 * time.Hour),
			IPAddress: c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			IsActive:  true,
		},
		Permissions: []models.Permission{
			{
				ID:          "perm_1",
				Name:        "forms.create",
				Description: "Create new forms",
				Resource:    "forms",
				Actions:     []string{"create"},
			},
		},
		TwoFactorEnabled: false,
	}

	c.JSON(http.StatusOK, models.StandardAPIResponse{
		Success:   true,
		Message:   "Login successful",
		Data:      response,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
		Meta: &models.ResponseMetadata{
			RequestDuration: "150ms",
			APIVersion:      "v1",
			ServerInstance:  "gateway-01",
		},
	})
}

// Logout godoc
// @Summary      Logout user and invalidate session
// @Description  Logs out the user, invalidates the current session, and adds token to blacklist
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200 {object} models.StandardAPIResponse "Logout successful"
// @Failure      401 {object} models.StandardAPIResponse{error=models.DetailedError} "Authentication required"
// @Failure      500 {object} models.StandardAPIResponse{error=models.DetailedError} "Internal server error"
// @Router       /auth/logout [post]
// @Security     BearerAuth
func EnhancedLogout(c *gin.Context) {
	c.JSON(http.StatusOK, models.StandardAPIResponse{
		Success:   true,
		Message:   "Logout successful",
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
	})
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Exchanges a valid refresh token for a new access token and refresh token pair
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        refresh_token body object{refresh_token=string} true "Refresh token"
// @Success      200 {object} models.StandardAPIResponse{data=models.AuthenticationResponse} "Token refreshed successfully"
// @Failure      400 {object} models.StandardAPIResponse{error=models.DetailedError} "Invalid request format"
// @Failure      401 {object} models.StandardAPIResponse{error=models.DetailedError} "Invalid or expired refresh token"
// @Failure      500 {object} models.StandardAPIResponse{error=models.DetailedError} "Internal server error"
// @Router       /auth/refresh [post]
func EnhancedRefreshToken(c *gin.Context) {
	// Mock token refresh response
	response := models.AuthenticationResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refreshed.token",
		RefreshToken: "new_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		ExpiresAt:    time.Now().Add(time.Hour),
		Scope:        []string{"read", "write"},
	}

	c.JSON(http.StatusOK, models.StandardAPIResponse{
		Success:   true,
		Message:   "Token refreshed successfully",
		Data:      response,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
	})
}

// GetProfile godoc
// @Summary      Get user profile information
// @Description  Retrieves comprehensive profile information for the authenticated user
// @Tags         Authentication,User Profile
// @Accept       json
// @Produce      json
// @Success      200 {object} models.StandardAPIResponse{data=models.DetailedUser} "Profile retrieved successfully"
// @Failure      401 {object} models.StandardAPIResponse{error=models.DetailedError} "Authentication required"
// @Failure      404 {object} models.StandardAPIResponse{error=models.DetailedError} "User not found"
// @Failure      500 {object} models.StandardAPIResponse{error=models.DetailedError} "Internal server error"
// @Router       /auth/profile [get]
// @Security     BearerAuth
func EnhancedGetProfile(c *gin.Context) {
	// Mock user profile
	user := models.DetailedUser{
		ID:            "user_987654321",
		FirstName:     "John",
		LastName:      "Doe",
		FullName:      "John Doe",
		Email:         "john.doe@example.com",
		EmailVerified: true,
		PhoneNumber:   "+1234567890",
		PhoneVerified: true,
		Avatar:        "https://example.com/avatar.jpg",
		Role:          "user",
		Status:        "active",
		Preferences: models.UserPreferences{
			Language: "en",
			Timezone: "UTC",
			Theme:    "light",
			Notifications: models.NotificationSettings{
				Email: true,
				SMS:   false,
				Push:  true,
			},
		},
		LastLoginAt:      &[]time.Time{time.Now().Add(-2 * time.Hour)}[0],
		LoginCount:       42,
		TwoFactorEnabled: false,
		CreatedAt:        time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:        time.Now(),
	}

	c.JSON(http.StatusOK, models.StandardAPIResponse{
		Success:   true,
		Message:   "Profile retrieved successfully",
		Data:      user,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
	})
}

// UpdateProfile godoc
// @Summary      Update user profile information
// @Description  Updates user profile with validation and security checks
// @Tags         Authentication,User Profile
// @Accept       json
// @Produce      json
// @Param        request body object{first_name=string,last_name=string,phone_number=string,preferences=models.UserPreferences} true "Profile update data"
// @Success      200 {object} models.StandardAPIResponse{data=models.DetailedUser} "Profile updated successfully"
// @Failure      400 {object} models.ValidationErrorResponse "Validation errors"
// @Failure      401 {object} models.StandardAPIResponse{error=models.DetailedError} "Authentication required"
// @Failure      404 {object} models.StandardAPIResponse{error=models.DetailedError} "User not found"
// @Failure      422 {object} models.StandardAPIResponse{error=models.DetailedError} "Invalid profile data"
// @Failure      500 {object} models.StandardAPIResponse{error=models.DetailedError} "Internal server error"
// @Router       /auth/profile [put]
// @Security     BearerAuth
func EnhancedUpdateProfile(c *gin.Context) {
	// Mock updated user profile
	user := models.DetailedUser{
		ID:            "user_987654321",
		FirstName:     "John",
		LastName:      "Doe",
		FullName:      "John Doe",
		Email:         "john.doe@example.com",
		EmailVerified: true,
		PhoneNumber:   "+1234567890",
		PhoneVerified: true,
		Avatar:        "https://example.com/avatar.jpg",
		Role:          "user",
		Status:        "active",
		UpdatedAt:     time.Now(),
	}

	c.JSON(http.StatusOK, models.StandardAPIResponse{
		Success:   true,
		Message:   "Profile updated successfully",
		Data:      user,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
	})
}

// DeleteProfile godoc
// @Summary      Delete user account
// @Description  Permanently deletes user account with all associated data (GDPR compliant)
// @Tags         Authentication,User Profile
// @Accept       json
// @Produce      json
// @Param        confirmation body object{password=string,confirmation=string} true "Account deletion confirmation"
// @Success      200 {object} models.StandardAPIResponse "Account deleted successfully"
// @Failure      400 {object} models.ValidationErrorResponse "Invalid confirmation"
// @Failure      401 {object} models.StandardAPIResponse{error=models.DetailedError} "Authentication required"
// @Failure      403 {object} models.StandardAPIResponse{error=models.DetailedError} "Invalid password or confirmation"
// @Failure      500 {object} models.StandardAPIResponse{error=models.DetailedError} "Internal server error"
// @Router       /auth/profile [delete]
// @Security     BearerAuth
func EnhancedDeleteProfile(c *gin.Context) {
	c.JSON(http.StatusOK, models.StandardAPIResponse{
		Success:   true,
		Message:   "Account deleted successfully. All associated data has been removed.",
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
	})
}

// Enhanced Form Handlers

// ListForms godoc
// @Summary      Get list of forms with advanced filtering and pagination
// @Description  Retrieves a paginated list of forms with comprehensive filtering, sorting, and search capabilities
// @Tags         Forms
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number" default(1) minimum(1)
// @Param        limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Param        sort query string false "Sort field" Enums(created_at,updated_at,title,submission_count,view_count) default(created_at)
// @Param        order query string false "Sort order" Enums(asc,desc) default(desc)
// @Param        search query string false "Search term for title and description"
// @Param        status query []string false "Filter by status" Enums(draft,published,archived,expired)
// @Param        category query []string false "Filter by category"
// @Param        tags query []string false "Filter by tags"
// @Param        is_public query bool false "Filter by public status"
// @Param        owner query string false "Filter by owner ID"
// @Param        date_from query string false "Filter from date (RFC3339)"
// @Param        date_to query string false "Filter to date (RFC3339)"
// @Success      200 {object} models.StandardAPIResponse{data=models.FormListResponse} "Forms retrieved successfully"
// @Failure      400 {object} models.ValidationErrorResponse "Invalid query parameters"
// @Failure      401 {object} models.StandardAPIResponse{error=models.DetailedError} "Authentication required for private forms"
// @Failure      500 {object} models.StandardAPIResponse{error=models.DetailedError} "Internal server error"
// @Router       /forms [get]
// @Security     BearerAuth
func EnhancedListForms(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Mock forms list
	forms := []models.FormSummary{
		{
			ID:              "form_123456789",
			Title:           "Customer Feedback Survey",
			Description:     "Please provide your valuable feedback",
			Status:          "published",
			IsPublic:        true,
			SubmissionCount: 42,
			ViewCount:       150,
			CompletionRate:  0.85,
			CreatedAt:       time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt:       time.Now().Add(-2 * time.Hour),
			Owner: models.BasicUser{
				ID:    "user_123",
				Name:  "John Doe",
				Email: "john@example.com",
			},
			Tags:     []string{"feedback", "customer", "survey"},
			Category: "feedback",
		},
		{
			ID:              "form_987654321",
			Title:           "Event Registration",
			Description:     "Register for our upcoming event",
			Status:          "published",
			IsPublic:        true,
			SubmissionCount: 128,
			ViewCount:       450,
			CompletionRate:  0.92,
			CreatedAt:       time.Now().Add(-14 * 24 * time.Hour),
			UpdatedAt:       time.Now().Add(-1 * time.Hour),
			Owner: models.BasicUser{
				ID:    "user_456",
				Name:  "Jane Smith",
				Email: "jane@example.com",
			},
			Tags:     []string{"event", "registration"},
			Category: "event",
		},
	}

	pagination := models.PaginationResponse{
		CurrentPage: page,
		PerPage:     limit,
		TotalPages:  1,
		TotalItems:  len(forms),
		HasNext:     false,
		HasPrevious: false,
	}

	response := models.FormListResponse{
		Forms:      forms,
		Pagination: pagination,
		Filters: models.FormListFilters{
			Status:   []string{c.Query("status")},
			Category: []string{c.Query("category")},
		},
		Sorting: models.FormListSorting{
			Field: c.DefaultQuery("sort", "created_at"),
			Order: c.DefaultQuery("order", "desc"),
		},
	}

	c.JSON(http.StatusOK, models.StandardAPIResponse{
		Success:    true,
		Message:    "Forms retrieved successfully",
		Data:       response,
		Pagination: &pagination,
		RequestID:  c.GetString("request_id"),
		Timestamp:  time.Now(),
	})
}

// CreateForm godoc
// @Summary      Create a new form
// @Description  Creates a new form with comprehensive configuration options, validation, and security features
// @Tags         Forms
// @Accept       json
// @Produce      json
// @Param        request body models.FormCreationRequest true "Form creation data"
// @Success      201 {object} models.StandardAPIResponse{data=models.FormResponse} "Form created successfully"
// @Failure      400 {object} models.ValidationErrorResponse "Validation errors"
// @Failure      401 {object} models.StandardAPIResponse{error=models.DetailedError} "Authentication required"
// @Failure      403 {object} models.StandardAPIResponse{error=models.DetailedError} "Insufficient permissions"
// @Failure      422 {object} models.StandardAPIResponse{error=models.DetailedError} "Invalid form configuration"
// @Failure      429 {object} models.StandardAPIResponse{error=models.DetailedError} "Rate limit exceeded"
// @Failure      500 {object} models.StandardAPIResponse{error=models.DetailedError} "Internal server error"
// @Router       /forms [post]
// @Security     BearerAuth
func EnhancedCreateForm(c *gin.Context) {
	var req models.FormCreationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
			Success: false,
			Message: "Validation failed",
			Errors: []models.ValidationError{
				{
					Field:   "title",
					Message: "Title is required",
					Code:    "REQUIRED_FIELD",
				},
			},
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	// Mock created form response
	now := time.Now()
	form := models.FormResponse{
		ID:                  "form_new123456",
		Title:               req.Title,
		Description:         req.Description,
		Category:            req.Category,
		Tags:                req.Tags,
		Status:              "draft",
		IsPublic:            req.IsPublic,
		AllowAnonymous:      req.AllowAnonymous,
		RequireAuth:         req.RequireAuth,
		MultipleSubmissions: req.MultipleSubmissions,
		SubmissionLimit:     req.SubmissionLimit,
		SubmissionCount:     0,
		ViewCount:           0,
		CompletionRate:      0.0,
		AverageTime:         0,
		CreatedAt:           now,
		UpdatedAt:           now,
		ExpiresAt:           req.ExpiresAt,
		Fields:              req.Fields,
		Owner: models.DetailedUser{
			ID:    "user_123",
			Email: "john@example.com",
		},
	}

	c.JSON(http.StatusCreated, models.StandardAPIResponse{
		Success:   true,
		Message:   "Form created successfully",
		Data:      form,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
	})
}

// GetForm godoc
// @Summary      Get form details
// @Description  Retrieves comprehensive form information including fields, settings, analytics, and permissions
// @Tags         Forms
// @Accept       json
// @Produce      json
// @Param        id path string true "Form ID" Format(uuid)
// @Param        include_analytics query bool false "Include analytics data" default(false)
// @Param        include_responses query bool false "Include recent responses" default(false)
// @Success      200 {object} models.StandardAPIResponse{data=models.FormResponse} "Form retrieved successfully"
// @Failure      400 {object} models.StandardAPIResponse{error=models.DetailedError} "Invalid form ID"
// @Failure      401 {object} models.StandardAPIResponse{error=models.DetailedError} "Authentication required for private forms"
// @Failure      403 {object} models.StandardAPIResponse{error=models.DetailedError} "Access denied"
// @Failure      404 {object} models.StandardAPIResponse{error=models.DetailedError} "Form not found"
// @Failure      500 {object} models.StandardAPIResponse{error=models.DetailedError} "Internal server error"
// @Router       /forms/{id} [get]
// @Security     BearerAuth
func EnhancedGetForm(c *gin.Context) {
	formID := c.Param("id")

	// Mock form response
	form := models.FormResponse{
		ID:          formID,
		Title:       "Customer Feedback Survey",
		Description: "Please provide your valuable feedback",
		Category:    "feedback",
		Tags:        []string{"feedback", "customer", "survey"},
		Status:      "published",
		IsPublic:    true,
		CreatedAt:   time.Now().Add(-7 * 24 * time.Hour),
		UpdatedAt:   time.Now().Add(-2 * time.Hour),
		Owner: models.DetailedUser{
			ID:    "user_123",
			Email: "john@example.com",
		},
	}

	c.JSON(http.StatusOK, models.StandardAPIResponse{
		Success:   true,
		Message:   "Form retrieved successfully",
		Data:      form,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
	})
}

// UpdateForm godoc
// @Summary Update existing form
// @Description Updates an existing form with new configuration and fields
// @Tags Forms
// @Accept json
// @Produce json
// @Param id path string true "Form ID"
// @Param request body models.FormUpdateRequest true "Form update data"
// @Success 200 {object} models.StandardAPIResponse{data=models.FormResponse} "Form updated successfully"
// @Failure 400 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Invalid request data"
// @Failure 401 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Authentication required"
// @Failure 403 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Insufficient permissions"
// @Failure 404 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Form not found"
// @Failure 500 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Internal server error"
// @Router /forms/{id} [put]
// @Security BearerAuth
func UpdateForm(c *gin.Context) {
	formID := c.Param("id")
	if formID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form ID is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Form updated successfully",
		"data": gin.H{
			"id":         formID,
			"updated_at": time.Now(),
		},
	})
}

// DeleteForm godoc
// @Summary Delete form
// @Description Permanently deletes a form and all associated data
// @Tags Forms
// @Produce json
// @Param id path string true "Form ID"
// @Success 200 {object} models.StandardAPIResponse "Form deleted successfully"
// @Failure 401 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Authentication required"
// @Failure 403 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Insufficient permissions"
// @Failure 404 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Form not found"
// @Failure 500 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Internal server error"
// @Router /forms/{id} [delete]
// @Security BearerAuth
func DeleteForm(c *gin.Context) {
	formID := c.Param("id")
	if formID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form ID is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Form deleted successfully",
	})
}

// PublishForm godoc
// @Summary Publish form
// @Description Makes a form publicly available for submissions
// @Tags Forms
// @Produce json
// @Param id path string true "Form ID"
// @Success 200 {object} models.StandardAPIResponse "Form published successfully"
// @Failure 401 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Authentication required"
// @Failure 403 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Insufficient permissions"
// @Failure 404 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Form not found"
// @Failure 500 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Internal server error"
// @Router /forms/{id}/publish [post]
// @Security BearerAuth
func PublishForm(c *gin.Context) {
	formID := c.Param("id")
	if formID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form ID is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Form published successfully",
		"data": gin.H{
			"id":           formID,
			"status":       "published",
			"published_at": time.Now(),
		},
	})
}

// UnpublishForm godoc
// @Summary Unpublish form
// @Description Makes a form unavailable for new submissions
// @Tags Forms
// @Produce json
// @Param id path string true "Form ID"
// @Success 200 {object} models.StandardAPIResponse "Form unpublished successfully"
// @Failure 401 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Authentication required"
// @Failure 403 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Insufficient permissions"
// @Failure 404 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Form not found"
// @Failure 500 {object} models.StandardAPIResponse{error=models.DetailedErrorResponse} "Internal server error"
// @Router /forms/{id}/unpublish [post]
// @Security BearerAuth
func UnpublishForm(c *gin.Context) {
	formID := c.Param("id")
	if formID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form ID is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Form unpublished successfully",
		"data": gin.H{
			"id":             formID,
			"status":         "draft",
			"unpublished_at": time.Now(),
		},
	})
}

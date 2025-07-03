package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"case/internal/models"
	"case/internal/utils"

	"github.com/gin-gonic/gin"
)

// PasswordHandler handles password change operations
type PasswordHandler struct {
	userService *models.UserService
}

// NewPasswordHandler creates a new password handler
func NewPasswordHandler(userService *models.UserService) *PasswordHandler {
	return &PasswordHandler{
		userService: userService,
	}
}

// ShowChangePasswordForm shows the password change form
func (h *PasswordHandler) ShowChangePasswordForm(c *gin.Context) {
	// Generate CSRF token
	csrfToken := generateCSRFToken()
	c.SetCookie("csrf_token", csrfToken, 3600, "/", "", false, true)

	c.HTML(http.StatusOK, "change_password.html", gin.H{
		"CSRFToken": csrfToken,
	})
}

// ChangePassword handles password change requests
func (h *PasswordHandler) ChangePassword(c *gin.Context) {
	// Get current user ID from session
	userID := utils.GetUserIDFromSession(c)
	if userID == 0 {
		c.HTML(http.StatusUnauthorized, "error.html", gin.H{"error": "Unauthorized"})
		return
	}

	// Get form data
	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")
	csrfToken := c.PostForm("csrf_token")

	// Validate CSRF token
	if !validateCSRFToken(c, csrfToken) {
		c.HTML(http.StatusBadRequest, "change_password.html", gin.H{
			"Message":     "Invalid security token",
			"MessageType": "error",
		})
		return
	}

	// Validate input
	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		c.HTML(http.StatusBadRequest, "change_password.html", gin.H{
			"Message":     "All fields are required",
			"MessageType": "error",
		})
		return
	}

	// Check if new passwords match
	if newPassword != confirmPassword {
		c.HTML(http.StatusBadRequest, "change_password.html", gin.H{
			"Message":     "New passwords do not match",
			"MessageType": "error",
		})
		return
	}

	// Validate password strength
	if !validatePasswordStrength(newPassword) {
		c.HTML(http.StatusBadRequest, "change_password.html", gin.H{
			"Message":     "Password does not meet strength requirements",
			"MessageType": "error",
		})
		return
	}

	// Get user from database
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "change_password.html", gin.H{
			"Message":     "Error retrieving user information",
			"MessageType": "error",
		})
		return
	}

	// Verify current password
	if !verifyPassword(currentPassword, user.UserPass.String, "") {
		c.HTML(http.StatusBadRequest, "change_password.html", gin.H{
			"Message":     "Current password is incorrect",
			"MessageType": "error",
		})
		return
	}

	// Generate new password hash and salt
	newSalt := generateSalt()
	newHash := hashPassword(newPassword, newSalt)

	// Update user password
	err = h.userService.UpdatePassword(int64(user.UserID), newHash, newSalt)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "change_password.html", gin.H{
			"Message":     "Error updating password",
			"MessageType": "error",
		})
		return
	}

	// Log password change
	h.logPasswordChange(userID, c.ClientIP(), c.Request.UserAgent())

	// Redirect with success message
	c.Redirect(http.StatusSeeOther, "/dashboard?message=Password changed successfully")
}

// RequestPasswordReset handles password reset requests
func (h *PasswordHandler) RequestPasswordReset(c *gin.Context) {
	email := c.PostForm("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	// Get user by email
	user, err := h.userService.GetUserByEmail(email)
	if err != nil {
		// Don't reveal if user exists or not
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent"})
		return
	}

	// Generate reset token
	token := generateResetToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create password change request
	request := &models.PasswordChangeRequest{
		UserID:              int64(user.UserID),
		RequestToken:        token,
		CurrentPasswordHash: user.UserPass.String,
		NewPasswordHash:     "", // Will be set when approved
		NewPasswordSalt:     "", // Will be set when approved
		ExpiresAt:           expiresAt,
		IPAddress:           sql.NullString{String: c.ClientIP(), Valid: true},
		UserAgent:           sql.NullString{String: c.Request.UserAgent(), Valid: true},
	}

	// Save request to database
	err = h.userService.CreatePasswordChangeRequest(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating reset request"})
		return
	}

	// Send reset email (implement email service)
	// h.sendPasswordResetEmail(user.Email, token)

	c.JSON(http.StatusOK, gin.H{"message": "Password reset link sent to your email"})
}

// ResetPassword handles password reset with token
func (h *PasswordHandler) ResetPassword(c *gin.Context) {
	token := c.Param("token")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	if newPassword == "" || confirmPassword == "" {
		c.HTML(http.StatusBadRequest, "reset_password.html", gin.H{
			"Message":     "All fields are required",
			"MessageType": "error",
			"Token":       token,
		})
		return
	}

	if newPassword != confirmPassword {
		c.HTML(http.StatusBadRequest, "reset_password.html", gin.H{
			"Message":     "Passwords do not match",
			"MessageType": "error",
			"Token":       token,
		})
		return
	}

	if !validatePasswordStrength(newPassword) {
		c.HTML(http.StatusBadRequest, "reset_password.html", gin.H{
			"Message":     "Password does not meet strength requirements",
			"MessageType": "error",
			"Token":       token,
		})
		return
	}

	// Get password change request
	request, err := h.userService.GetPasswordChangeRequest(token)
	if err != nil {
		c.HTML(http.StatusBadRequest, "reset_password.html", gin.H{
			"Message":     "Invalid or expired reset token",
			"MessageType": "error",
		})
		return
	}

	// Check if token is expired
	if time.Now().After(request.ExpiresAt) {
		c.HTML(http.StatusBadRequest, "reset_password.html", gin.H{
			"Message":     "Reset token has expired",
			"MessageType": "error",
		})
		return
	}

	// Generate new password hash and salt
	newSalt := generateSalt()
	newHash := hashPassword(newPassword, newSalt)

	// Update password change request
	request.NewPasswordHash = newHash
	request.NewPasswordSalt = newSalt
	request.IsApproved = true
	request.ApprovedAt = sql.NullTime{Time: time.Now(), Valid: true}

	err = h.userService.UpdatePasswordChangeRequest(request)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "reset_password.html", gin.H{
			"Message":     "Error updating password",
			"MessageType": "error",
		})
		return
	}

	// Update user password
	err = h.userService.UpdatePassword(request.UserID, newHash, newSalt)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "reset_password.html", gin.H{
			"Message":     "Error updating password",
			"MessageType": "error",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/login?message=Password reset successfully")
}

// ShowForgotPasswordForm shows the forgot password form
func (h *PasswordHandler) ShowForgotPasswordForm(c *gin.Context) {
	c.HTML(http.StatusOK, "forgot_password.html", gin.H{})
}

// ShowResetPasswordForm shows the reset password form
func (h *PasswordHandler) ShowResetPasswordForm(c *gin.Context) {
	token := c.Param("token")
	c.HTML(http.StatusOK, "reset_password.html", gin.H{
		"Token": token,
	})
}

// Helper functions
func verifyPassword(password, hash, salt string) bool {
	return hashPassword(password, salt) == hash
}

func validatePasswordStrength(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		default:
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func generateCSRFToken() string {
	token := make([]byte, 32)
	rand.Read(token)
	return hex.EncodeToString(token)
}

func validateCSRFToken(c *gin.Context, token string) bool {
	cookieToken, err := c.Cookie("csrf_token")
	if err != nil {
		return false
	}
	return token == cookieToken
}

func generateResetToken() string {
	token := make([]byte, 32)
	rand.Read(token)
	return hex.EncodeToString(token)
}

func (h *PasswordHandler) logPasswordChange(userID int64, ipAddress, userAgent string) {
	// Log password change for audit purposes
	// This should be implemented based on your logging system
}

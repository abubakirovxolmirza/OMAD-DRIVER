package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"taxi-service/internal/config"
	"taxi-service/internal/database"
	"taxi-service/internal/middleware"
	"taxi-service/internal/models"
	"taxi-service/internal/utils"
)

// RegisterFiber godoc
// @Summary Register a new user
// @Description Register a new user with phone number, name and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) RegisterFiber(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.H{"error": err.Error()})
	}

	// Validate passwords match
	if req.Password != req.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.H{"error": "Passwords do not match"})
	}

	// Check if user already exists
	var existingID int64
	err := database.DB.QueryRow("SELECT id FROM users WHERE phone_number = $1", req.PhoneNumber).Scan(&existingID)
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.H{"error": "Phone number already registered"})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.H{"error": "Failed to process password"})
	}

	// Insert user
	var user models.User
	err = database.DB.QueryRow(`
		INSERT INTO users (phone_number, name, password, role, language)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, phone_number, name, role, language, avatar, is_blocked, created_at, updated_at
	`, req.PhoneNumber, req.Name, hashedPassword, models.RoleUser, models.LangUzLatin).Scan(
		&user.ID, &user.PhoneNumber, &user.Name, &user.Role, &user.Language,
		&user.Avatar, &user.IsBlocked, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.H{"error": "Failed to create user"})
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Role, h.cfg.JWT.Secret, h.cfg.JWT.ExpirationHours)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.H{"error": "Failed to generate token"})
	}

	return c.Status(fiber.StatusCreated).JSON(AuthResponse{
		Token: token,
		Role:  string(user.Role),
		User:  &user,
	})
}

// LoginFiber godoc
// @Summary Login user
// @Description Login with phone number and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) LoginFiber(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.H{"error": err.Error()})
	}

	// Get user by phone number
	var user models.User
	err := database.DB.QueryRow(`
		SELECT id, phone_number, name, password, role, language, avatar, is_blocked, created_at, updated_at
		FROM users WHERE phone_number = $1
	`, req.PhoneNumber).Scan(
		&user.ID, &user.PhoneNumber, &user.Name, &user.Password, &user.Role,
		&user.Language, &user.Avatar, &user.IsBlocked, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.H{"error": "Invalid credentials"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.H{"error": "Database error"})
	}

	// Check if user is blocked
	if user.IsBlocked {
		return c.Status(fiber.StatusForbidden).JSON(fiber.H{"error": "Account is blocked"})
	}

	// Verify password
	if err := utils.CheckPassword(user.Password, req.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.H{"error": "Invalid credentials"})
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Role, h.cfg.JWT.Secret, h.cfg.JWT.ExpirationHours)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.H{"error": "Failed to generate token"})
	}

	// Clear password from response
	user.Password = ""

	return c.Status(fiber.StatusOK).JSON(AuthResponse{
		Token: token,
		Role:  string(user.Role),
		User:  &user,
	})
}

// GetProfileFiber godoc
// @Summary Get user profile
// @Description Get current user's profile information
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]string
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfileFiber(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFiber(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.H{"error": "User not found"})
	}

	var user models.User
	err := database.DB.QueryRow(`
		SELECT id, phone_number, name, role, language, avatar, is_blocked, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.PhoneNumber, &user.Name, &user.Role,
		&user.Language, &user.Avatar, &user.IsBlocked, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.H{"error": "Failed to get profile"})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// UpdateProfileFiber godoc
// @Summary Update user profile
// @Description Update user's name and language preference
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile update details"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Router /auth/profile [put]
func (h *AuthHandler) UpdateProfileFiber(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFiber(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.H{"error": "User not found"})
	}

	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.H{"error": err.Error()})
	}

	var user models.User
	err := database.DB.QueryRow(`
		UPDATE users SET name = $1, language = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING id, phone_number, name, role, language, avatar, is_blocked, created_at, updated_at
	`, req.Name, req.Language, userID).Scan(
		&user.ID, &user.PhoneNumber, &user.Name, &user.Role,
		&user.Language, &user.Avatar, &user.IsBlocked, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.H{"error": "Failed to update profile"})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// ChangePasswordFiber godoc
// @Summary Change user password
// @Description Change user's password
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Password change details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePasswordFiber(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFiber(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.H{"error": "User not found"})
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.H{"error": err.Error()})
	}

	// Validate new passwords match
	if req.NewPassword != req.ConfirmNewPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.H{"error": "New passwords do not match"})
	}

	// Get current password
	var currentPassword string
	err := database.DB.QueryRow("SELECT password FROM users WHERE id = $1", userID).Scan(&currentPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.H{"error": "Database error"})
	}

	// Verify old password
	if err := utils.CheckPassword(currentPassword, req.OldPassword); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.H{"error": "Invalid old password"})
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.H{"error": "Failed to process password"})
	}

	// Update password
	_, err = database.DB.Exec("UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", hashedPassword, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.H{"error": "Failed to update password"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.H{"message": "Password changed successfully"})
}

// UploadAvatarFiber godoc
// @Summary Upload user avatar
// @Description Upload an avatar image for the user
// @Tags Auth
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar image"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/avatar [post]
func (h *AuthHandler) UploadAvatarFiber(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFiber(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.H{"error": "User not found"})
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.H{"error": "No file uploaded"})
	}

	// Check file size
	if file.Size > h.cfg.Upload.MaxFileSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.H{"error": "File too large"})
	}

	// Get old avatar to delete
	var oldAvatar sql.NullString
	database.DB.QueryRow("SELECT avatar FROM users WHERE id = $1", userID).Scan(&oldAvatar)

	// Save new file
	relativePath, err := utils.SaveUploadedFileFiber(file, h.cfg.Upload.Directory, "avatars")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.H{"error": err.Error()})
	}

	// Update user avatar
	_, err = database.DB.Exec("UPDATE users SET avatar = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", relativePath, userID)
	if err != nil {
		utils.DeleteFile(h.cfg.Upload.Directory, relativePath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.H{"error": "Failed to update avatar"})
	}

	// Delete old avatar
	if oldAvatar.Valid {
		utils.DeleteFile(h.cfg.Upload.Directory, oldAvatar.String)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.H{
		"message": "Avatar uploaded successfully",
		"avatar":  relativePath,
	})
}

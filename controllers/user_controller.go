package controllers

import (
	"net/http"

	"gotestbackend/database"
	"gotestbackend/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// LoginPayload is used to bind login request body
type LoginPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UpdateUserPayload is used to bind update request body
type UpdateUserPayload struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Password      string `json:"password"`
	AccountNumber string `json:"account_number"`
}

// Login handles user login
func Login(c *gin.Context) {
	var payload LoginPayload
	var user models.User

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Where("username = ?", payload.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user_id": user.ID})
}

// GetUser retrieves the logged-in user's details
func GetUserProfile(c *gin.Context) {
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser updates the logged-in user's details
func UpdateUser(c *gin.Context) {
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in"})
		return
	}

	var payload UpdateUserPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if payload.FirstName != "" {
		user.FirstName = payload.FirstName
	}
	if payload.LastName != "" {
		user.LastName = payload.LastName
	}
	if payload.AccountNumber != "" {
		user.AccountNumber = payload.AccountNumber
	}
	if payload.Password != "" {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		user.Password = string(hashedPassword)
	}

	database.DB.Save(&user)
	c.JSON(http.StatusOK, user)
}

func Register(c *gin.Context) {
	var newUser models.User

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	newUser.Password = string(hashedPassword)

	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simulate credit addition
	newUser.Credit = 1000.0

	// Save user to database (assuming db is initialized in main.go)
	database.DB.Create(&newUser)

	c.JSON(http.StatusCreated, newUser)
}

// api/accounting.go
// TransferCredit transfers credit from one user to another
func TransferCredit(c *gin.Context) {
	// Parse request body
	var transferRequest struct {
		SenderAccount   string  `json:"sender_account"`
		ReceiverAccount string  `json:"receiver_account"`
		Amount          float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&transferRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Implement transfer logic
	// Check if sender and receiver IDs are valid
	sender, err := GetUserByAccount(transferRequest.SenderAccount)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sender not found"})
		return
	}

	receiver, err := GetUserByAccount(transferRequest.ReceiverAccount)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receiver not found"})
		return
	}

	// Update sender and receiver credits in database
	// Validate if sender has enough credit
	if sender.Credit < transferRequest.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient credit"})
		return
	}

	// Perform credit transfer
	// db.Model(&sender).Update("credit", sender.Credit - amount)
	sender.Credit -= transferRequest.Amount
	// db.Model(&receiver).Update("credit", receiver.Credit + amount)
	receiver.Credit += transferRequest.Amount

	// Update sender and receiver in database
	err = database.DB.Save(&sender).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sender"})
		return
	}

	err = database.DB.Save(&receiver).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update receiver"})
		return
	}

	// Record transaction
	transaction := models.Transaction{
		SenderID:   sender.ID,
		ReceiverID: receiver.ID,
		Amount:     transferRequest.Amount,
	}
	err = database.DB.Create(&transaction).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}

func GetAllUser(c *gin.Context) {
	var user models.User
	if err := database.DB.Find(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "All record not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}
func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}
func GetUserByAccount(account_number string) (models.User, error) {
	var user models.User
	if err := database.DB.First(&user, account_number).Error; err != nil {
		return user, nil
	}
	return user, nil
}

func UpdateUserByID(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var updatedUser models.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user fields
	database.DB.Model(&user).Updates(updatedUser)

	c.JSON(http.StatusOK, user)
}

func DeleteUserByID(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Delete user
	database.DB.Delete(&user)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

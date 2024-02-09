package auth_controllers

import (
	"math/rand"
	"time"

	"github.com/Pratham-Mishra04/fiber-template/config"
	"github.com/Pratham-Mishra04/fiber-template/helpers"
	"github.com/Pratham-Mishra04/fiber-template/initializers"
	"github.com/Pratham-Mishra04/fiber-template/models"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func generateRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func SendResetURL(c *fiber.Ctx) error {
	var reqBody struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return &fiber.Error{Code: 400, Message: "Validation Failed"}
	}

	var user models.User
	if err := initializers.DB.Where("email=?", reqBody.Email).First(&user).Error; err != nil {
		return &fiber.Error{Code: 400, Message: "No User with this email found."}
	}

	code := generateRandomString(32)
	hash, err := bcrypt.GenerateFromPassword([]byte(code), 10)
	if err != nil {
		go helpers.LogServerError("Error while hashing an Reset Password Token.", err, c.Path())
		return helpers.AppError{Code: 500, Message: config.SERVER_ERROR, Err: err}
	}

	expirationTime := time.Now().Add(config.VERIFICATION_OTP_EXPIRATION_TIME)

	user.PasswordResetToken = string(hash)
	user.PasswordResetTokenExpires = expirationTime
	result := initializers.DB.Save(&user)
	if result.Error != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: result.Error.Error(), Err: result.Error}
	}

	resetURL := initializers.CONFIG.FRONTEND_URL + "/account_recovery?uid=" + user.ID.String() + "&token=" + code

	err = helpers.SendMail("Account Recovery | fiber-template", "Hi "+user.Username+"(your username)"+", "+"Reset your Password on this URL: "+resetURL, user.Name, user.Email, "<div><strong>This is Valid for next 10 minutes only!</strong></div>")
	if err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "OTP sent to registered mail",
	})
}

func ResetPassword(c *fiber.Ctx) error {
	var reqBody struct {
		UserID   string `json:"userID"`
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return &fiber.Error{Code: 400, Message: "Validation Failed"}
	}

	var user models.User
	if err := initializers.DB.Where("id=?", reqBody.UserID).First(&user).Error; err != nil {
		return &fiber.Error{Code: 400, Message: "Invalid Credentials."}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordResetToken), []byte(reqBody.Token)); err != nil {
		return &fiber.Error{Code: 400, Message: "Invalid Reset Token"}
	}

	if time.Now().After(user.PasswordResetTokenExpires) {
		return &fiber.Error{Code: 400, Message: "URL has Expired, generate a new one"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(reqBody.Password), 10)
	if err != nil {
		go helpers.LogServerError("Error while hashing Password.", err, c.Path())
		return helpers.AppError{Code: 500, Message: config.SERVER_ERROR, Err: err}
	}

	user.Password = string(hash)
	user.PasswordChangedAt = time.Now()
	user.PasswordResetTokenExpires = time.Now()

	result := initializers.DB.Save(&user)
	if result.Error != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: result.Error.Error(), Err: result.Error}
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Password Changed",
	})
}

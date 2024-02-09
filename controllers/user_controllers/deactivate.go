package user_controllers

import (
	"errors"
	"time"

	"github.com/Pratham-Mishra04/fiber-template/cache"
	"github.com/Pratham-Mishra04/fiber-template/config"
	"github.com/Pratham-Mishra04/fiber-template/controllers/auth_controllers"
	"github.com/Pratham-Mishra04/fiber-template/helpers"
	"github.com/Pratham-Mishra04/fiber-template/initializers"
	"github.com/Pratham-Mishra04/fiber-template/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Deactivate(c *fiber.Ctx) error {
	userID := c.GetRespHeader("loggedInUserID")

	var reqBody struct {
		VerificationCode string `json:"otp"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return &fiber.Error{Code: 400, Message: "OTP not provided."}
	}

	var user models.User
	if err := initializers.DB.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &fiber.Error{Code: 400, Message: "No user of this ID found."}
		}
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	data, err := cache.GetOtpFromCache(user.ID.String())
	if err != nil {
		return &fiber.Error{Code: 400, Message: "OTP Expired"}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(data), []byte(reqBody.VerificationCode)); err != nil {
		return &fiber.Error{Code: 400, Message: "Incorrect OTP"}
	}

	user.Active = false
	user.DeactivatedAt = time.Now()

	if err := initializers.DB.Save(&user).Error; err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	cache.RemoveUser(user.ID.String())

	return c.Status(204).JSON(fiber.Map{
		"status":  "success",
		"message": "Account Deactivated",
	})
}

func SendDeactivateVerificationCode(c *fiber.Ctx) error {
	loggedInUserID := c.GetRespHeader("loggedInUserID")
	parsedLoggedInUserID, _ := uuid.Parse(loggedInUserID)

	code := auth_controllers.GenerateOTP(6)
	hash, err := bcrypt.GenerateFromPassword([]byte(code), 10)
	if err != nil {
		go helpers.LogServerError("Error while hashing an OTP.", err, c.Path())
		return helpers.AppError{Code: 500, Message: config.SERVER_ERROR, Err: err}
	}

	var user models.User
	if err := initializers.DB.Where("id=?", parsedLoggedInUserID).First(&user).Error; err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}
	err = helpers.SendMail(config.VERIFICATION_DELETE_SUBJECT, config.VERIFICATION_EMAIL_BODY+code, user.Name, user.Email, "<div><strong>This is Valid for next 10 minutes only!</strong></div>")
	if err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	err = cache.SetOtpToCache(user.ID.String(), []byte(hash))
	if err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "OTP sent to registered mail",
	})
}

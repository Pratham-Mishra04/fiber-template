package auth_controllers

import (
	"crypto/rand"
	"io"
	"time"

	"github.com/Pratham-Mishra04/fiber-template/config"
	"github.com/Pratham-Mishra04/fiber-template/helpers"
	"github.com/Pratham-Mishra04/fiber-template/initializers"
	"github.com/Pratham-Mishra04/fiber-template/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func GenerateOTP(max int) string {
	var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

func SendVerificationCode(c *fiber.Ctx) error {
	//TODO add a limiter to keep a check on spam requests
	loggedInUserID := c.GetRespHeader("loggedInUserID")
	parsedLoggedInUserID, _ := uuid.Parse(loggedInUserID)

	code := GenerateOTP(6)
	hash, err := bcrypt.GenerateFromPassword([]byte(code), 10)
	if err != nil {
		go helpers.LogServerError("Error while hashing an OTP.", err, c.Path())
		return helpers.AppError{Code: 500, Message: config.SERVER_ERROR, Err: err}
	}

	expirationTime := time.Now().Add(config.VERIFICATION_OTP_EXPIRATION_TIME)

	var user models.User
	if err := initializers.DB.Where("id=?", parsedLoggedInUserID).First(&user).Error; err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	if user.Verified {
		return &fiber.Error{Code: 400, Message: "User is already verified"}
	}

	var verification models.UserVerification
	if err := initializers.DB.Where("user_id=?", user.ID).First(&verification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			newVerification := models.UserVerification{
				UserID:         parsedLoggedInUserID,
				Code:           string(hash),
				ExpirationTime: expirationTime,
			}
			result := initializers.DB.Create(&newVerification)
			if result.Error != nil {
				return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: result.Error.Error(), Err: result.Error}
			}
		} else {
			return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
		}
	} else {
		verification.Code = string(hash)
		verification.ExpirationTime = expirationTime
		result := initializers.DB.Save(&verification)
		if result.Error != nil {
			return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: result.Error.Error(), Err: result.Error}
		}
	}
	err = helpers.SendMail(config.VERIFICATION_EMAIL_SUBJECT, config.VERIFICATION_EMAIL_BODY+code, user.Name, user.Email, "<div><strong>This is Valid for next 10 minutes only!</strong></div>")
	if err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "OTP sent to registered mail",
	})
}

func VerifyCode(c *fiber.Ctx) error {
	loggedInUserID := c.GetRespHeader("loggedInUserID")
	parsedLoggedInUserID, _ := uuid.Parse(loggedInUserID)

	var reqBody struct {
		VerificationCode string `json:"otp"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return &fiber.Error{Code: 400, Message: "Validation Failed"}
	}

	var user models.User
	if err := initializers.DB.Where("id=?", parsedLoggedInUserID).First(&user).Error; err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	if user.Verified {
		return &fiber.Error{Code: 400, Message: "User is already verified"}
	}

	var verification models.UserVerification
	if err := initializers.DB.Where("user_id=?", user.ID).First(&verification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &fiber.Error{Code: 400, Message: "First Request for the Verification Code"}
		} else {
			return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(verification.Code), []byte(reqBody.VerificationCode)); err != nil {
		return &fiber.Error{Code: 400, Message: "Incorrect OTP"}
	}

	if time.Now().After(verification.ExpirationTime) {
		return &fiber.Error{Code: 400, Message: "OTP has Expired, generate a new one"}
	}

	user.Verified = true
	result := initializers.DB.Save(&user)
	if result.Error != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: result.Error.Error(), Err: result.Error}
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "User Verified",
	})
}

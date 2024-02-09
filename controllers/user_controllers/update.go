package user_controllers

import (
	"errors"
	"time"

	"github.com/Pratham-Mishra04/fiber-template/config"
	"github.com/Pratham-Mishra04/fiber-template/controllers/auth_controllers"
	"github.com/Pratham-Mishra04/fiber-template/helpers"
	"github.com/Pratham-Mishra04/fiber-template/initializers"
	"github.com/Pratham-Mishra04/fiber-template/models"
	"github.com/Pratham-Mishra04/fiber-template/routines"
	"github.com/Pratham-Mishra04/fiber-template/schemas"
	"github.com/Pratham-Mishra04/fiber-template/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func UpdateMe(c *fiber.Ctx) error {
	userID := c.GetRespHeader("loggedInUserID")
	var user models.User
	if err := initializers.DB.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &fiber.Error{Code: 400, Message: "No user of this ID found."}
		}
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	var reqBody schemas.UserUpdateSchema
	if err := c.BodyParser(&reqBody); err != nil {
		return &fiber.Error{Code: 400, Message: "Invalid Request Body."}
	}

	// if err := helpers.Validate[schemas.UserUpdateSchema](reqBody); err != nil {
	// 	return &fiber.Error{Code: 400, Message: err.Error()}
	// }

	oldProfilePic := user.ProfilePic
	oldCoverPic := user.CoverPic

	// picName, err := utils.SaveFile(c, "profilePic", "user/profilePics", true, 500, 500)
	picName, err := utils.UploadImage(c, "profilePic", helpers.UserProfileClient, 500, 500)
	if err != nil {
		return err
	}
	reqBody.ProfilePic = &picName

	// coverName, err := utils.SaveFile(c, "coverPic", "user/coverPics", true, 900, 400)
	coverName, err := utils.UploadImage(c, "coverPic", helpers.UserCoverClient, 900, 400)
	if err != nil {
		return err
	}
	reqBody.CoverPic = &coverName

	// updateUserValue := reflect.ValueOf(&reqBody).Elem()
	// userValue := reflect.ValueOf(&user).Elem()

	// for i := 0; i < updateUserValue.NumField(); i++ {
	// 	field := updateUserValue.Type().Field(i)
	// 	fieldName := field.Name

	// 	if fieldValue := updateUserValue.Field(i); fieldValue.IsValid() && !fieldValue.IsZero() {
	// 		userField := userValue.FieldByName(fieldName)
	// 		if userField.IsValid() && userField.CanSet() {
	// 			userField.Set(fieldValue)
	// 		}
	// 	}
	// }

	if reqBody.Name != nil {
		user.Name = *reqBody.Name
	}
	if reqBody.Bio != nil {
		user.Bio = *reqBody.Bio
	}
	if reqBody.ProfilePic != nil && *reqBody.ProfilePic != "" {
		user.ProfilePic = *reqBody.ProfilePic
	}
	if reqBody.CoverPic != nil && *reqBody.CoverPic != "" {
		user.CoverPic = *reqBody.CoverPic
	}
	if reqBody.Tags != nil {
		user.Tags = *reqBody.Tags
	}

	if err := initializers.DB.Save(&user).Error; err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	if *reqBody.ProfilePic != "" {
		go routines.DeleteFromBucket(helpers.UserProfileClient, oldProfilePic)
	}

	if *reqBody.CoverPic != "" {
		go routines.DeleteFromBucket(helpers.UserCoverClient, oldCoverPic)
	}

	if c.Query("action", "") == "onboarding" && !user.OnboardingCompleted {
		go func() {
			user.OnboardingCompleted = true
			if err := initializers.DB.Save(&user).Error; err != nil {
				helpers.LogDatabaseError("Error while updating User-UpdateMe", err, "go_routine")
			}
		}()
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "User updated successfully",
		"user":    user,
	})
}

func UpdatePassword(c *fiber.Ctx) error {
	var reqBody struct {
		Password        string `json:"password"`
		NewPassword     string `json:"newPassword"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return &fiber.Error{Code: 400, Message: "Validation Failed"}
	}

	if reqBody.NewPassword != reqBody.ConfirmPassword {
		return &fiber.Error{Code: 400, Message: "Passwords do not match."}
	}

	loggedInUserID := c.GetRespHeader("loggedInUserID")

	var user models.User
	initializers.DB.First(&user, "id = ?", loggedInUserID)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reqBody.Password)); err != nil {
		return &fiber.Error{Code: 400, Message: "Incorrect Password."}
	}

	//TODO send email for verification

	hash, err := bcrypt.GenerateFromPassword([]byte(reqBody.NewPassword), 10)

	if err != nil {
		return helpers.AppError{Code: 500, Message: config.SERVER_ERROR, Err: err}
	}

	user.Password = string(hash)
	user.PasswordChangedAt = time.Now()

	if err := initializers.DB.Save(&user).Error; err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	return auth_controllers.CreateSendToken(c, user, 200, "Password updated successfully")
}

func UpdateEmail(c *fiber.Ctx) error {
	userID := c.GetRespHeader("loggedInUserID")

	var reqBody struct {
		Email string `json:"email" validate:"required,email"`
	}
	if err := c.BodyParser(&reqBody); err != nil {
		return &fiber.Error{Code: 400, Message: "Validation Failed"}
	}

	var emailCheckUser models.User
	if err := initializers.DB.First(&emailCheckUser, "email = ?", reqBody.Email).Error; err == nil {
		return &fiber.Error{Code: 400, Message: "Email Address Already In Use."}
	}

	var user models.User
	if err := initializers.DB.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &fiber.Error{Code: 400, Message: "No user of this ID found."}
		}
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	user.Email = reqBody.Email
	user.Verified = false

	if err := initializers.DB.Save(&user).Error; err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "User updated successfully",
	})
}

func UpdatePhoneNo(c *fiber.Ctx) error {
	userID := c.GetRespHeader("loggedInUserID")

	var reqBody struct {
		PhoneNo string `json:"phoneNo"  validate:"e164"`
	}
	if err := c.BodyParser(&reqBody); err != nil {
		return &fiber.Error{Code: 400, Message: "Validation Failed"}
	}

	var phoneNoCheckUser models.User
	if err := initializers.DB.First(&phoneNoCheckUser, "phone_no = ?", reqBody.PhoneNo).Error; err == nil {
		return &fiber.Error{Code: 400, Message: "Phone Number Already In Use."}
	}

	var user models.User
	if err := initializers.DB.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &fiber.Error{Code: 400, Message: "No user of this ID found."}
		}
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	user.PhoneNo = reqBody.PhoneNo

	if err := initializers.DB.Save(&user).Error; err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "User updated successfully",
	})
}

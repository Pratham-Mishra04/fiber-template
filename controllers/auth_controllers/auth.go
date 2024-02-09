package auth_controllers

import (
	"errors"
	"fmt"
	"time"

	"github.com/Pratham-Mishra04/fiber-template/config"
	"github.com/Pratham-Mishra04/fiber-template/helpers"
	"github.com/Pratham-Mishra04/fiber-template/initializers"
	"github.com/Pratham-Mishra04/fiber-template/models"
	"github.com/Pratham-Mishra04/fiber-template/schemas"
	"github.com/Pratham-Mishra04/fiber-template/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CreateSendToken(c *fiber.Ctx, user models.User, statusCode int, message string) error {
	access_token_claim := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"crt": time.Now().Unix(),
		"exp": time.Now().Add(config.ACCESS_TOKEN_TTL).Unix(),
	})

	refresh_token_claim := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"crt": time.Now().Unix(),
		"exp": time.Now().Add(config.REFRESH_TOKEN_TTL).Unix(),
	})

	access_token, err := access_token_claim.SignedString([]byte(initializers.CONFIG.JWT_SECRET))
	if err != nil {
		go helpers.LogServerError("Error while decrypting JWT Token.", err, c.Path())
		return helpers.AppError{Code: 500, Message: config.SERVER_ERROR, Err: err}
	}

	refresh_token, err := refresh_token_claim.SignedString([]byte(initializers.CONFIG.JWT_SECRET))
	if err != nil {
		go helpers.LogServerError("Error while decrypting JWT Token.", err, c.Path())
		return helpers.AppError{Code: 500, Message: config.SERVER_ERROR, Err: err}
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refresh_token,
		Expires:  time.Now().Add(config.REFRESH_TOKEN_TTL),
		HTTPOnly: true,
		Secure:   true,
	})

	return c.Status(statusCode).JSON(fiber.Map{
		"status":  "success",
		"message": message,
		"token":   access_token,
		"user":    user,
		"email":   user.Email,
		"phoneNo": user.PhoneNo,
	})
}

func SignUp(c *fiber.Ctx) error {
	var reqBody schemas.UserCreateSchema

	c.BodyParser(&reqBody)

	hash, err := bcrypt.GenerateFromPassword([]byte(reqBody.Password), 12)
	if err != nil {
		go helpers.LogServerError("Error while hashing Password.", err, c.Path())
		return helpers.AppError{Code: 500, Message: config.SERVER_ERROR, Err: err}
	}

	newUser := models.User{
		Name:              reqBody.Name,
		Email:             reqBody.Email,
		Password:          string(hash),
		Username:          reqBody.Username,
		PasswordChangedAt: time.Now(),
	}

	result := initializers.DB.Create(&newUser)
	if result.Error != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: result.Error.Error(), Err: result.Error}
	}

	c.Set("loggedInUserID", newUser.ID.String())

	picName, err := utils.UploadImage(c, "profilePic", helpers.UserProfileClient, 500, 500)
	if err != nil {
		initializers.Logger.Warnw("Error in Saving Profile Pic on Sign Up", "Err", err)
	} else {
		if picName != "" {
			newUser.ProfilePic = picName

			result = initializers.DB.Save(&newUser)
			if result.Error != nil {
				return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: result.Error.Error(), Err: result.Error}
			}
		}
	}

	return CreateSendToken(c, newUser, 201, "Account Created")
}

func OAuthSignUp(c *fiber.Ctx) error {
	loggedInUserID := c.GetRespHeader("loggedInUserID")
	var reqBody struct {
		Username string `json:"username" validate:"alphanum,required"`
	}

	c.BodyParser(&reqBody)

	var user models.User
	if err := initializers.DB.Session(&gorm.Session{SkipHooks: true}).First(&user, "id = ?", loggedInUserID).Error; err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	var existingUser models.User
	initializers.DB.Session(&gorm.Session{SkipHooks: true}).First(&existingUser, "username = ?", reqBody.Username)
	if existingUser.ID != uuid.Nil {
		return &fiber.Error{Code: 400, Message: "User with this Username already exists"}
	}

	var oauth models.OAuth
	if err := initializers.DB.First(&oauth, "user_id = ?", loggedInUserID).Error; err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	user.Username = reqBody.Username
	user.Verified = true
	oauth.OnBoardingCompleted = true

	result := initializers.DB.Save(&user)
	if result.Error != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: result.Error.Error(), Err: result.Error}
	}

	result = initializers.DB.Save(&oauth)
	if result.Error != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: result.Error.Error(), Err: result.Error}
	}

	c.Set("loggedInUserID", user.ID.String())

	// picName, err := utils.SaveFile(c, "profilePic", "user/profilePics", false, 500, 500)
	picName, err := utils.UploadImage(c, "profilePic", helpers.UserProfileClient, 500, 500)
	if err != nil {
		initializers.Logger.Warnw("Error in Saving Profile Pic on Sign Up", "Err", err)
	} else {
		if picName != "" {
			user.ProfilePic = picName

			result = initializers.DB.Save(&user)
			if result.Error != nil {
				return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: result.Error.Error(), Err: result.Error}
			}
		}
	}

	return CreateSendToken(c, user, 201, "Account Created")
}

func LogIn(c *fiber.Ctx) error {
	var reqBody struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return &fiber.Error{Code: 400, Message: "Validation Failed"}
	}

	var user models.User
	if err := initializers.DB.Session(&gorm.Session{SkipHooks: true}).First(&user, "username = ? AND organization_status = false", reqBody.Username).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &fiber.Error{Code: 400, Message: "No account with these credentials found."}
		} else {
			return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reqBody.Password)); err != nil {
		return &fiber.Error{Code: 400, Message: "No account with these credentials found."}
	}

	if !user.Active {
		if time.Now().After(user.DeactivatedAt.Add(30 * 24 * time.Hour)) {
			return &fiber.Error{Code: 400, Message: "Cannot Log into a deactivated account."}
		}
		user.Active = true
	}

	user.LastLoggedIn = time.Now()

	if err := initializers.DB.Save(&user).Error; err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	return CreateSendToken(c, user, 200, "Logged In")
}

func OAuthLogIn(c *fiber.Ctx) error {
	loggedInUserID := c.GetRespHeader("loggedInUserID")

	var user models.User
	if err := initializers.DB.Session(&gorm.Session{SkipHooks: true}).First(&user, "id = ? AND organization_status = false", loggedInUserID).Error; err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	if user.ID == uuid.Nil {
		return &fiber.Error{Code: 400, Message: "No user with these credentials found."}
	}

	if !user.Active {
		if time.Now().After(user.DeactivatedAt.Add(30 * 24 * time.Hour)) {
			return &fiber.Error{Code: 400, Message: "Cannot Log into a deactivated account."}
		}
		user.Active = true
	}

	user.LastLoggedIn = time.Now()

	if err := initializers.DB.Save(&user).Error; err != nil {
		return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
	}

	return CreateSendToken(c, user, 200, "Logged In")
}

func Refresh(c *fiber.Ctx) error {
	var reqBody struct {
		Token string `json:"token"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return &fiber.Error{Code: 400, Message: "Validation Failed"}
	}

	access_token_string := reqBody.Token

	access_token, err := jwt.Parse(access_token_string, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(initializers.CONFIG.JWT_SECRET), nil
	})

	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		initializers.Logger.Infow("Token Expiration: ", "Error", err)
		return &fiber.Error{Code: 400, Message: config.TOKEN_EXPIRED_ERROR}
	}

	if access_token_claims, ok := access_token.Claims.(jwt.MapClaims); ok {
		access_token_userID, ok := access_token_claims["sub"].(string)
		if !ok {
			return &fiber.Error{Code: 401, Message: "Invalid user ID in token claims."}
		}

		var user models.User
		err := initializers.DB.First(&user, "id = ?", access_token_userID).Error
		if err != nil {
			return helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
		}

		if user.ID == uuid.Nil {
			return &fiber.Error{Code: 401, Message: "User of this token no longer exists"}
		}

		refresh_token_string := c.Cookies("refresh_token")
		if refresh_token_string == "" {
			initializers.Logger.Infow("Token Expiration: ", "Error", err)
			return &fiber.Error{Code: 401, Message: config.TOKEN_EXPIRED_ERROR}
		}

		refresh_token, err := jwt.Parse(refresh_token_string, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(initializers.CONFIG.JWT_SECRET), nil
		})

		if err != nil {
			initializers.Logger.Infow("Token Expiration: ", "Error", err)
			return &fiber.Error{Code: 400, Message: config.TOKEN_EXPIRED_ERROR}
		}

		if refresh_token_claims, ok := refresh_token.Claims.(jwt.MapClaims); ok && refresh_token.Valid {
			refresh_token_userID, ok := refresh_token_claims["sub"].(string)
			if !ok {
				return &fiber.Error{Code: 401, Message: "Invalid user ID in token claims."}
			}

			if refresh_token_userID != access_token_userID {
				initializers.Logger.Warnw("Mismatched Tokens: ", "Access Token User ID", access_token_userID, "Refresh Token User ID", refresh_token_userID)
				return &fiber.Error{Code: 401, Message: "Mismatched Tokens."}
			}

			if time.Now().After(time.Unix(int64(refresh_token_claims["exp"].(float64)), 0)) {
				initializers.Logger.Infow("Token Expiration: ", "Error", err)
				return &fiber.Error{Code: 401, Message: config.TOKEN_EXPIRED_ERROR}
			}

			new_access_token_claim := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub": user.ID,
				"crt": time.Now().Unix(),
				"exp": time.Now().Add(config.ACCESS_TOKEN_TTL).Unix(),
			})

			new_access_token, err := new_access_token_claim.SignedString([]byte(initializers.CONFIG.JWT_SECRET))
			if err != nil {
				go helpers.LogServerError("Error while decrypting JWT Token.", err, c.Path())
				return helpers.AppError{Code: 500, Message: config.SERVER_ERROR, Err: err}
			}

			return c.Status(200).JSON(fiber.Map{
				"status": "success",
				"token":  new_access_token,
			})
		}

		return nil
	} else {
		return &fiber.Error{Code: 401, Message: "Invalid Token"}
	}
}

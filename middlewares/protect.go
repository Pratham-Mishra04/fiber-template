package middlewares

import (
	"fmt"
	"strings"
	"time"

	"github.com/Pratham-Mishra04/fiber-template/cache"
	"github.com/Pratham-Mishra04/fiber-template/config"
	"github.com/Pratham-Mishra04/fiber-template/helpers"
	"github.com/Pratham-Mishra04/fiber-template/initializers"
	"github.com/Pratham-Mishra04/fiber-template/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func verifyToken(tokenString string, user *models.User, checkRedirect bool) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(initializers.CONFIG.JWT_SECRET), nil
	})

	if err != nil {
		return nil, &fiber.Error{Code: 403, Message: config.TOKEN_EXPIRED_ERROR}
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			return nil, &fiber.Error{Code: 403, Message: "Your token has expired."}
		}
		if checkRedirect {
			rdt, ok := claims["rdt"].(bool)
			if !ok {
				if initializers.CONFIG.ENV == initializers.DevelopmentEnv {
					return nil, &fiber.Error{Code: 403, Message: "Not a redirect Token."}
				} else {
					return nil, &fiber.Error{Code: 403, Message: "Connection Timeout, Login again"}
				}
			}

			if !rdt {
				if initializers.CONFIG.ENV == initializers.DevelopmentEnv {
					return nil, &fiber.Error{Code: 403, Message: "Not a redirect Token."}
				} else {
					return nil, &fiber.Error{Code: 403, Message: "Connection Timeout, Login again"}
				}
			}
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			return nil, &fiber.Error{Code: 401, Message: "Invalid user ID in token claims."}
		}

		userInCache, err := cache.GetUser(userID)
		if err == nil {
			user = userInCache
		} else {
			if err := initializers.DB.First(&user, "id = ?", userID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, &fiber.Error{Code: 401, Message: "User of this token no longer exists"}
				}
				return nil, helpers.AppError{Code: 500, Message: config.DATABASE_ERROR, LogMessage: err.Error(), Err: err}
			}

			go cache.SetUser(user.ID.String(), user)
		}

		// TODO
		// if user.PasswordChangedAt.After(time.Unix(int64(claims["crt"].(float64)), 0)) {
		// 	return &fiber.Error{Code: 401, Message: "Password was recently changed, log in again."}
		// }

		return user, nil
	} else {
		return nil, &fiber.Error{Code: 403, Message: "Invalid Token"}
	}
}

func Protect(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	tokenArr := strings.Split(authHeader, " ")

	if len(tokenArr) != 2 {
		return &fiber.Error{Code: 401, Message: "You are Not Logged In."}
	}

	tokenString := tokenArr[1]

	var user *models.User
	user, err := verifyToken(tokenString, user, false)
	if err != nil {
		return err
	}

	c.Set("loggedInUserID", user.ID.String())

	return c.Next()
}

func PartialProtect(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	tokenArr := strings.Split(authHeader, " ")

	if len(tokenArr) != 2 {
		return c.Next()
	}

	tokenString := tokenArr[1]

	var user *models.User
	user, err := verifyToken(tokenString, user, false)
	if err != nil {
		return err
	}

	c.Set("loggedInUserID", user.ID.String())

	return c.Next()
}

func ProtectRedirect(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	tokenArr := strings.Split(authHeader, " ")

	if len(tokenArr) != 2 {
		return &fiber.Error{Code: 401, Message: "You are Not Logged In."}
	}

	tokenString := tokenArr[1]

	var user *models.User
	user, err := verifyToken(tokenString, user, true)
	if err != nil {
		return err
	}

	c.Set("loggedInUserID", user.ID.String())

	return c.Next()
}

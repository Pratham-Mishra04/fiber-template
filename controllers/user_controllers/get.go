package user_controllers

import (
	"github.com/Pratham-Mishra04/fiber-template/initializers"
	"github.com/Pratham-Mishra04/fiber-template/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetMe(c *fiber.Ctx) error {
	userID := c.GetRespHeader("loggedInUserID")

	var user models.User
	initializers.DB.
		Preload("Profile").
		Preload("Profile.Achievements").
		First(&user, "id = ?", userID)

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "",
		"user":    user,
	})
}

func GetUser(c *fiber.Ctx) error {
	username := c.Params("username")

	var user models.User
	initializers.DB.Preload("Profile").First(&user, "username = ?", username)

	if user.ID == uuid.Nil {
		return &fiber.Error{Code: 400, Message: "No user of this username found."}
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "User Found",
		"user":    user,
	})
}

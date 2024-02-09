package routers

import (
	"github.com/Pratham-Mishra04/fiber-template/controllers/auth_controllers"
	"github.com/Pratham-Mishra04/fiber-template/middlewares"
	"github.com/gofiber/fiber/v2"
)

func VerificationRouter(app *fiber.App) {
	workspaceRoutes := app.Group("/verification", middlewares.Protect)
	workspaceRoutes.Get("/otp", auth_controllers.SendVerificationCode)
	workspaceRoutes.Post("/otp", auth_controllers.VerifyCode)
}

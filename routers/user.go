package routers

import (
	"github.com/Pratham-Mishra04/fiber-template/controllers/auth_controllers"
	"github.com/Pratham-Mishra04/fiber-template/controllers/user_controllers"
	"github.com/Pratham-Mishra04/fiber-template/middlewares"
	"github.com/Pratham-Mishra04/fiber-template/validators"
	"github.com/gofiber/fiber/v2"
)

func UserRouter(app *fiber.App) {
	app.Post("/signup", validators.UserCreateValidator, auth_controllers.SignUp)
	app.Post("/login", auth_controllers.LogIn)
	app.Post("/refresh", auth_controllers.Refresh)

	app.Post("/recovery", auth_controllers.SendResetURL)
	app.Post("/recovery/verify", auth_controllers.ResetPassword)

	userRoutes := app.Group("/users", middlewares.Protect)
	userRoutes.Get("/me", user_controllers.GetMe)

	userRoutes.Patch("/update_password", user_controllers.UpdatePassword)
	userRoutes.Patch("/update_email", user_controllers.UpdateEmail)
	userRoutes.Patch("/update_phone_number", user_controllers.UpdatePhoneNo)

	userRoutes.Get("/deactivate", user_controllers.SendDeactivateVerificationCode)
	userRoutes.Post("/deactivate", user_controllers.Deactivate)

	userRoutes.Patch("/me", user_controllers.UpdateMe)
	userRoutes.Delete("/me", user_controllers.Deactivate)
}

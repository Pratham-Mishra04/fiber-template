package routers

import (
	"github.com/Pratham-Mishra04/fiber-template/controllers/auth_controllers"
	"github.com/Pratham-Mishra04/fiber-template/middlewares"
	"github.com/gofiber/fiber/v2"
)

func OauthRouter(app *fiber.App) {
	oauthRoutes := app.Group("/auth")
	oauthRoutes.Post("/signup", middlewares.ProtectRedirect, auth_controllers.OAuthSignUp)
	oauthRoutes.Get("/login", middlewares.ProtectRedirect, auth_controllers.OAuthLogIn)

	oauthRoutes.Get("/google", auth_controllers.GoogleRedirect)
	oauthRoutes.Get("/google/callback", auth_controllers.GoogleCallback)
}

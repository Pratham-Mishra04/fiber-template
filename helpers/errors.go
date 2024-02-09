package helpers

import (
	"github.com/Pratham-Mishra04/fiber-template/config"
	"github.com/gofiber/fiber/v2"
)

type AppError struct {
	Code       int
	Message    string
	LogMessage string
	Err        error
}

func (err AppError) Error() string {
	return err.LogMessage
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	Code := 500
	Message := config.SERVER_ERROR
	Error := err

	if e, ok := err.(*fiber.Error); ok {
		Code = e.Code
		Message = e.Message
	}

	if e, ok := err.(*AppError); ok {
		Code = e.Code
		Message = e.Message
		Error = e.Err
	}

	if Message == config.DATABASE_ERROR {
		go LogDatabaseError("Database Error", Error, c.Path())
	} else if Code == 500 {
		go LogServerError("Server Error", Error, c.Path())
	}

	return c.Status(Code).JSON(fiber.Map{
		"status":  "failed",
		"message": Message,
	})
}

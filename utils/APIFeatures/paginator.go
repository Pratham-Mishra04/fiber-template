package utils

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Paginator(c *fiber.Ctx) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		pageStr := c.Query("page", "1")
		limitStr := c.Query("limit", "10")

		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Println("Failed to Paginate due to integer conversion.")
			return db
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			log.Println("Failed to Paginate due to integer conversion.")
			return db
		}

		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}

package utils

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Search(c *fiber.Ctx, index int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		searchStr := c.Query("search", "")
		if searchStr == "" {
			return db
		}

		searchStr = strings.ToLower(searchStr)

		switch index {
		case 0: //* users
			db = db.Where("LOWER(name) LIKE ? OR LOWER(username) LIKE ? OR ? = ANY (tags)", "%"+searchStr+"%", "%"+searchStr+"%", searchStr)
			return db
		default:
			return db
		}
	}
}

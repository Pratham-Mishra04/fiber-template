package utils

import (
	"github.com/Pratham-Mishra04/fiber-template/helpers"
	"github.com/gofiber/fiber/v2"
)

func SaveMultipleFiles(c *fiber.Ctx, fieldName string, path string, resize bool, d1 int, d2 int) ([]string, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return []string{}, err
	}

	files := form.File[fieldName]
	if files == nil {
		return []string{}, nil
	}

	var savedImages []string
	for _, file := range files {

		filePath := "public/" + path + "/" + c.GetRespHeader("loggedInUserID") + "-" + file.Filename

		if err := c.SaveFile(file, filePath); err != nil {
			helpers.LogServerError("Error while saving a file", err, c.Path())
			return nil, err
		}

		if resize {
			picName, err := ResizeSavedImage(filePath, d1, d2)
			if err != nil {
				return []string{}, err
			}
			savedImages = append(savedImages, picName)
		} else {
			picName := c.GetRespHeader("loggedInUserID") + "-" + file.Filename
			savedImages = append(savedImages, picName)
		}
	}

	return savedImages, nil
}

package utils

import (
	"bytes"
	"image"
	"image/jpeg"
	"mime/multipart"
	"os"
	"path"
	"strings"

	"github.com/nfnt/resize"
)

func ResizeSavedImage(picPath string, d1 int, d2 int) (string, error) {
	file, err := os.Open(picPath)
	if err != nil {
		return "", err
	}

	img, err := jpeg.Decode(file)
	if err != nil {
		return "", err
	}
	file.Close()

	m := resize.Resize(uint(d1), uint(d2), img, resize.Lanczos3)

	extension := path.Ext(picPath)
	fileNameWithoutExt := picPath[:len(picPath)-len(extension)]

	resizedPicPath := fileNameWithoutExt + "-resized.jpg"

	out, err := os.Create(resizedPicPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	err = jpeg.Encode(out, m, nil)
	if err != nil {
		return "", err
	}

	err = os.Remove(picPath)
	if err != nil {
		return "", err
	}

	resizedPicArr := strings.Split(resizedPicPath, "/")

	resizedPic := resizedPicArr[len(resizedPicArr)-1]

	return resizedPic, nil
}

func ResizeFormImage(file *multipart.FileHeader, width int, height int) (*bytes.Buffer, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		return nil, err
	}

	resizedImg := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	// Encode the resized image to a buffer
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resizedImg, nil); err != nil {
		return nil, err
	}

	return &buf, nil
}

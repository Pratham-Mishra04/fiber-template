package cache

import (
	"github.com/Pratham-Mishra04/fiber-template/models"
)

func GetUser(slug string) (*models.User, error) {
	var user models.User
	err := GetFromCacheGeneric("user-"+slug, &user)
	return &user, err
}

func SetUser(slug string, user *models.User) error {
	return SetToCacheGeneric("user-"+slug, user)
}

func RemoveUser(slug string) error {
	return RemoveFromCacheGeneric("user-" + slug)
}

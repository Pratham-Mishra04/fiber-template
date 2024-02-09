package cache

import (
	"encoding/json"
	"fmt"
)

func GetFromCacheGeneric(key string, model interface{}) error {
	data, err := GetFromCache(key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(data), model); err != nil {
		return fmt.Errorf("error while unmarshaling %s: %w", key, err)
	}

	return nil
}

func SetToCacheGeneric(key string, model interface{}) error {
	data, err := json.Marshal(model)
	if err != nil {
		return fmt.Errorf("error while marshaling %s: %w", key, err)
	}

	if err := SetToCache(key, data); err != nil {
		return err
	}

	return nil
}

func RemoveFromCacheGeneric(key string) error {
	if err := RemoveFromCache(key); err != nil {
		return err
	}

	return nil
}

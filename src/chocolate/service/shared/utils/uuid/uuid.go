package uuid

import (
	gouuid "github.com/satori/go.uuid"
)

func New() (string, error) {
	_uuid, err := gouuid.NewV4()
	if err != nil {
		return "", err
	}
	return _uuid.String(), nil
}

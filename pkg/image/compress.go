package image

import (
	"github.com/google/uuid"
	"github.com/h2non/bimg"
	"strings"
)

func CompressImage(buffer []byte, quality int) ([]byte, string, error) {
	filename := strings.Replace(uuid.New().String(), "-", "", -1) + ".webp"

	converted, err := bimg.NewImage(buffer).Convert(bimg.WEBP)
	if err != nil {
		return nil, filename, err
	}

	processed, err := bimg.NewImage(converted).Process(bimg.Options{Quality: quality})
	if err != nil {
		return nil, filename, err
	}

	return processed, filename, nil
}

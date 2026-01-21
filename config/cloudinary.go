// config/cloudinary.go
package config

import (
	"github.com/cloudinary/cloudinary-go/v2"
)

func ConnectCloudinary(cfg *Config) *cloudinary.Cloudinary {
	cloudinary, err := cloudinary.NewFromParams(cfg.CLOUDINARY_CLOUD_NAME, cfg.CLOUDINARY_API_KEY, cfg.CLOUDINARY_API_SECRET)
	if err != nil {
		panic("CLOUDINARY_NOT_SET")
	}

	return cloudinary
}

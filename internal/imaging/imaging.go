package imaging

import (
	"bytes"
	"image"

	"github.com/deepteams/webp"
	"github.com/disintegration/imaging"

	"github.com/danilov-go/gophprofile/internal/models"
)

type Resizer struct{}

func NewResizer() *Resizer {
	return &Resizer{}
}

func (r *Resizer) Resize(imageData []byte) ([]models.ThumbnailData, error) {
	img, err := imaging.Decode(bytes.NewReader(imageData), imaging.AutoOrientation(true))
	if err != nil {
		return nil, models.ErrInvalidFormat
	}
	thumbnailData := make([]models.ThumbnailData, 0, len(models.Targets))
	for size, target := range models.Targets {
		thumb := imaging.Fill(img, target.Width, target.Height, imaging.Center, imaging.Lanczos)
		var buf bytes.Buffer
		if err := imaging.Encode(&buf, thumb, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
			return nil, err
		}
		thumbnailData = append(thumbnailData, models.ThumbnailData{
			Size: size,
			Data: buf.Bytes(),
		})
	}
	return thumbnailData, nil
}

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (r *Converter) Convert(imageData []byte, format string) ([]byte, string, error) {
	if format == "" {
		return nil, "", models.ErrInvalidFormat
	}
	img, err := imaging.Decode(bytes.NewReader(imageData), imaging.AutoOrientation(true))
	if err != nil {
		return nil, "", models.ErrInvalidFormat
	}
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		if err := imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/jpeg", nil

	case "png":
		if err := imaging.Encode(&buf, img, imaging.PNG); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/png", nil

	case "webp":
		if err := webp.Encode(&buf, img, &webp.EncoderOptions{Quality: 80}); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/webp", nil
	default:
		return nil, "", models.ErrInvalidFormat
	}
}

func (c *Converter) GetDimensions(imageData []byte) (models.Dimensions, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(imageData))
	if err != nil {
		return models.Dimensions{}, models.ErrInvalidFormat
	}
	return models.Dimensions{
		Width:  cfg.Width,
		Height: cfg.Height,
	}, nil
}

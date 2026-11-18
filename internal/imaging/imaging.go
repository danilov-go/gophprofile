package imaging

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/png"

	"github.com/danilov-go/gophprofile/internal/models"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

type Resizer struct{}

func New() *Resizer {
	return &Resizer{}
}

func (r *Resizer) Resize(imageData []byte) ([]models.ThumbnailData, error) {
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, err
	}
	var thumbnailData []models.ThumbnailData
	for size, target := range models.Targets {
		imgRect := image.Rect(0, 0, target.Width, target.Height)
		imgRGBA := image.NewRGBA(imgRect)
		draw.CatmullRom.Scale(imgRGBA, imgRect, img, img.Bounds(), draw.Over, nil)
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, imgRGBA, &jpeg.Options{Quality: 85}); err != nil {
			return nil, err
		}
		thumbnailData = append(thumbnailData, models.ThumbnailData{
			Size: size,
			Data: buf.Bytes(),
		})
	}
	return thumbnailData, nil
}

package imaging_test

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	_ "github.com/deepteams/webp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danilov-go/gophprofile/internal/imaging"
	"github.com/danilov-go/gophprofile/internal/models"
)

func createPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h))))
	return buf.Bytes()
}

func TestResizer_Resize(t *testing.T) {
	resizer := imaging.NewResizer()
	tests := []struct {
		name   string
		data   []byte
		expErr error
	}{
		{
			name: "положительный тест",
			data: createPNG(t, 550, 620),
		},
		{
			name:   "невалидные данные изображения",
			data:   []byte("invalid"),
			expErr: models.ErrInvalidFormat,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thumbs, err := resizer.Resize(tt.data)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
				assert.Nil(t, thumbs)
				return
			}
			require.NoError(t, err)
			require.Len(t, thumbs, len(models.Targets))
			for _, thumb := range thumbs {
				target, ok := models.Targets[thumb.Size]
				require.True(t, ok)
				cfg, format, err := image.DecodeConfig(bytes.NewReader(thumb.Data))
				require.NoError(t, err)
				assert.Equal(t, "jpeg", format)
				assert.Equal(t, target.Width, cfg.Width)
				assert.Equal(t, target.Height, cfg.Height)
			}
		})
	}
}

func TestConverter_Convert(t *testing.T) {
	converter := imaging.NewConverter()
	imgBytes := createPNG(t, 550, 620)
	tests := []struct {
		name    string
		data    []byte
		format  string
		expMime string
		expErr  error
	}{
		{
			name:    "конвертация в jpeg",
			data:    imgBytes,
			format:  "jpeg",
			expMime: "image/jpeg",
		},
		{
			name:    "конвертация в png",
			data:    imgBytes,
			format:  "png",
			expMime: "image/png",
		},
		{
			name:    "конвертация в webp",
			data:    imgBytes,
			format:  "webp",
			expMime: "image/webp",
		},
		{
			name:   "пустой формат",
			data:   imgBytes,
			format: "",
			expErr: models.ErrInvalidFormat,
		},
		{
			name:   "неподдерживаемый формат",
			data:   imgBytes,
			format: "gif",
			expErr: models.ErrInvalidFormat,
		},
		{
			name:   "поврежденные данные изображения",
			data:   []byte("invalid"),
			format: "jpeg",
			expErr: models.ErrInvalidFormat,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, mime, err := converter.Convert(tt.data, tt.format)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
				assert.Nil(t, out)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expMime, mime)
			cfg, decodedFmt, err := image.DecodeConfig(bytes.NewReader(out))
			require.NoError(t, err)
			assert.Equal(t, tt.format, decodedFmt)
			assert.Equal(t, 550, cfg.Width)
			assert.Equal(t, 620, cfg.Height)
		})
	}
}

func TestConverter_GetDimensions(t *testing.T) {
	converter := imaging.NewConverter()
	tests := []struct {
		name   string
		data   []byte
		exp    models.Dimensions
		expErr error
	}{
		{
			name: "положительный тест",
			data: createPNG(t, 550, 620),
			exp:  models.Dimensions{Width: 550, Height: 620},
		},
		{
			name:   "невалидные данные изображения",
			data:   []byte("invalid"),
			expErr: models.ErrInvalidFormat,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dims, err := converter.GetDimensions(tt.data)
			if tt.expErr != nil {
				require.ErrorIs(t, err, tt.expErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.exp, dims)
		})
	}
}

package models

// SizeAvatar представляет доступные размеры аватарок.
type SizeAvatar string

// Поддерживаемых размеры аватарок.
const (
	SizeSmall    SizeAvatar = "100x100"
	SizeMedium   SizeAvatar = "300x300"
	SizeOriginal SizeAvatar = "original"
)

// Dimensions определяет ширину и высоту аватари в пикселях.
type Dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Thumbnail определяет размер и URL-адрес миниатюры аватарки.
type Thumbnail struct {
	Size string `json:"size"`
	URL  string `json:"url"`
}

type ThumbnailData struct {
	Size SizeAvatar
	Data []byte
}

var Targets = map[SizeAvatar]Dimensions{
	SizeSmall:  {Width: 100, Height: 100},
	SizeMedium: {Width: 300, Height: 300},
}

var mimeToExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// GetExt возвращает расширение аватарки по ее MIME-типу.
func GetExt(mimeType string) (string, bool) {
	ext, ok := mimeToExt[mimeType]
	return ext, ok
}

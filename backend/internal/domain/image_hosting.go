package domain

type ImageHostingSetting struct {
	Enabled bool   `json:"enabled"`
	APIKey  string `json:"apiKey"`
}

type ImageHostingFile struct {
	FileID    int    `json:"file_id"`
	Filename  string `json:"filename"`
	Storename string `json:"storename"`
	URL       string `json:"url"`
	Delete    string `json:"delete"`
	Hash      string `json:"hash"`
	Path      string `json:"path"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Size      int    `json:"size"`
	CreatedAt int    `json:"created_at"`
	Page      string `json:"page"`
}

type ImageHostingListResponse struct {
	Code    int                 `json:"code"`
	Data    []ImageHostingFile  `json:"data"`
	Message string              `json:"message"`
	Success bool                `json:"success"`
}

type ImageHostingUploadResponse struct {
	Code    int               `json:"code"`
	Data    ImageHostingFile  `json:"data"`
	Message string            `json:"message"`
	Success bool              `json:"success"`
}

type ImageHostingDeleteResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type ImageHostingRepository interface {
	GetSetting() (*ImageHostingSetting, error)
	SaveSetting(setting *ImageHostingSetting) error
}

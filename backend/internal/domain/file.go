package domain

type FileInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type UploadedFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

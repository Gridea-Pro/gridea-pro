package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const seeAPIBase = "https://s.ee/api/v1"

type ImageHostingService struct {
	repo domain.ImageHostingRepository
}

func NewImageHostingService(repo domain.ImageHostingRepository) *ImageHostingService {
	return &ImageHostingService{repo: repo}
}

func (s *ImageHostingService) GetSetting() (*domain.ImageHostingSetting, error) {
	return s.repo.GetSetting()
}

func (s *ImageHostingService) SaveSetting(setting *domain.ImageHostingSetting) error {
	return s.repo.SaveSetting(setting)
}

func (s *ImageHostingService) Upload(filePath string) (*domain.ImageHostingFile, error) {
	setting, err := s.repo.GetSetting()
	if err != nil {
		return nil, err
	}
	if setting.APIKey == "" {
		return nil, fmt.Errorf("API 密钥未配置")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", seeAPIBase+"/file/upload", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", setting.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("上传请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result domain.ImageHostingUploadResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if !result.Success && result.Code != 0 && result.Code != 200 {
		return nil, fmt.Errorf("上传失败: %s", result.Message)
	}

	return &result.Data, nil
}

func (s *ImageHostingService) List(page int) (*domain.ImageHostingListResponse, error) {
	setting, err := s.repo.GetSetting()
	if err != nil {
		return nil, err
	}
	if setting.APIKey == "" {
		return nil, fmt.Errorf("API 密钥未配置")
	}

	url := seeAPIBase + "/files"
	if page > 0 {
		url += "?page=" + strconv.Itoa(page)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", setting.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取文件列表失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result domain.ImageHostingListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if !result.Success && result.Code != 0 && result.Code != 200 {
		return nil, fmt.Errorf("获取失败: %s", result.Message)
	}

	return &result, nil
}

func (s *ImageHostingService) Delete(hash string) error {
	setting, err := s.repo.GetSetting()
	if err != nil {
		return err
	}
	if setting.APIKey == "" {
		return fmt.Errorf("API 密钥未配置")
	}

	req, err := http.NewRequest("GET", seeAPIBase+"/file/delete/"+hash, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", setting.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("删除请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result domain.ImageHostingDeleteResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("删除失败: %s", result.Message)
	}

	return nil
}

func (s *ImageHostingService) UploadFromFrontend(files []domain.UploadedFile) ([]string, error) {
	var urls []string
	for _, f := range files {
		result, err := s.Upload(f.Path)
		if err != nil {
			return nil, fmt.Errorf("上传 %s 失败: %w", f.Name, err)
		}
		urls = append(urls, result.URL)
	}
	return urls, nil
}

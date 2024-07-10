package comfyui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	PreviewNodeUrl = "%s/manager/preview_method?value=%s"
	Queueurl       = "%s/queue"
	HistoryUrl     = "%s/history?max_items=200"
	PromptUrl      = "%s/prompt"
	ViewUrl        = "%s/view?%s"
	UploadLink     = "%s/upload/image"
	// ws
	ComfyWSLink = "%s/ws?clientId=%s"
)

func PreviewMode(ctx context.Context, comfyEndpoint, mode string) error {
	_, err := get(fmt.Sprintf(PreviewNodeUrl, comfyEndpoint, mode))
	if err != nil {
		fmt.Println("PreviewMode error:", err)
		return err
	}
	return nil
}

func Queue(ctx context.Context, endpoint string) (queuePendingSize, queueRunningSize int64, err error) {
	r, err := get(fmt.Sprintf(Queueurl, endpoint))
	if err != nil {
		return 0, 0, err
	}
	var comfyQueueRes = &QueueRes{}
	err = json.Unmarshal([]byte(r), comfyQueueRes)
	if err != nil {
		return 0, 0, err
	}
	return int64(len(comfyQueueRes.QueuePending)), int64(len(comfyQueueRes.QueueRunning)), nil
}

func History(ctx context.Context, endpoint string) (outputMap map[string]*HistoryData, err error) {
	resp, err := getBytes(fmt.Sprintf(HistoryUrl, endpoint))
	if err != nil {
		return
	}
	fmt.Println("history resp:", string(resp))
	outputMap = make(map[string]*HistoryData)
	if err = json.Unmarshal(resp, &outputMap); err != nil {
		return nil, err
	}
	return
}

func QueuePrompt(ctx context.Context, endpoint, params string) (promptID string, err error) {
	queueUrl := fmt.Sprintf(PromptUrl, endpoint)
	resp, err := postBytes(queueUrl, params)
	if err != nil {
		return "", err
	}
	fmt.Println(string(resp))
	queuePromptRes := &ComfyQueuePromptRes{}
	err = json.Unmarshal(resp, queuePromptRes)
	if err != nil {
		return
	}
	if queuePromptRes.PromptID == "" {
		return "", errors.New("prompt_id is nil")
	}
	return queuePromptRes.PromptID, nil
}

func GetComfyPreviewLink(ctx context.Context, endpoint string, img *ImageInfo) string {
	data := url.Values{}
	data.Set("filename", img.Filename)
	data.Set("subfolder", img.SubFolder)
	data.Set("type", img.Type)
	baseURL := fmt.Sprintf(ViewUrl, endpoint, data.Encode())
	return baseURL
}

func GetComfyImage(ctx context.Context, endpoint string, img *ImageInfo) (bytes []byte, err error) {
	baseUrl := GetComfyPreviewLink(ctx, endpoint, img)
	fmt.Println("view url:", baseUrl)
	// 发送 HTTP 请求
	bytes, err = getBytes(baseUrl)
	if err != nil {
		return
	}
	return
}

// fileLocation like: /xxx/xxxx/xxxx.jpg
func UploadFileToServer(endpoint, fileLocation string) (returns *ComfyImageUploadRes, err error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, err := os.Open(fileLocation)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(fileLocation)
	}()
	part1, err := writer.CreateFormFile("image", filepath.Base(fileLocation))
	_, err = io.Copy(part1, file)
	if err != nil {
		return nil, err
	}
	_ = writer.WriteField("subfolder", time.Now().Format("2006/01/02"))

	if err != nil {
		return nil, err
	}
	_ = writer.Close()
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf(UploadLink, endpoint), payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	var body []byte
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	returns = &ComfyImageUploadRes{}
	err = json.Unmarshal(body, returns)
	if err != nil {
		return nil, err
	}
	return returns, nil
}

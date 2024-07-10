package comfyui

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func postBytes(url, body string) (bodyBytes []byte, err error) {
	payload := strings.NewReader(body)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	bodyBytes, err = io.ReadAll(res.Body)
	if err != nil {
		return
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status + string(bodyBytes))
	}
	return
}

func get(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return string(respBytes), err
	}
	if resp.StatusCode != 200 {
		return "", errors.New("resp status code:" + fmt.Sprintf("%d, body:%s", resp.StatusCode, string(respBytes)))
	}
	return string(respBytes), nil
}

func getBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("resp status code:" + fmt.Sprintf("%d, body:%s", resp.StatusCode, string(respBytes)))
	}
	return respBytes, nil
}

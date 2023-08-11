package lib

import (
	"encoding/json"
	"io"
	"net/http"
)

func TMI(item interface{}) map[string]interface{} {
	return item.(map[string]interface{})
}
func GetHttpString(url string) (string, error) {
	respBody, err := GetHttpByte(url)
	if err != nil {
		return "", err
	}
	return string(respBody), nil
}

func GetHttpJson(url string, model any) (interface{}, error) {
	body, err := GetHttpByte(url)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &model)
	if err != nil {
		return nil, err
	}
	return model, nil
}
func GetHttpByte(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

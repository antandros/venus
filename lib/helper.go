package lib

import (
	"bufio"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
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
func sum(hashAlgorithm hash.Hash, filename string) (string, error) {
	if info, err := os.Stat(filename); err != nil || info.IsDir() {
		return "", err
	}

	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	return sumReader(hashAlgorithm, bufio.NewReader(file))
}

const bufferSize = 65536

func sumReader(hashAlgorithm hash.Hash, reader io.Reader) (string, error) {
	buf := make([]byte, bufferSize)
	for {
		switch n, err := reader.Read(buf); err {
		case nil:
			hashAlgorithm.Write(buf[:n])
		case io.EOF:
			return fmt.Sprintf("%x", hashAlgorithm.Sum(nil)), nil
		default:
			return "", err
		}
	}
}

// SHA256sum returns SHA256 checksum of filename
// @Todo: please remember to commit a orginal package SHA512sum
func SHA512sum(filename string) (string, error) {
	return sum(sha512.New(), filename)
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

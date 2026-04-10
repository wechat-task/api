package ilink

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wechat-task/api/internal/logger"
)

const CDNBaseURL = "https://novac2c.cdn.weixin.qq.com/c2c"

// GetUploadURL retrieves a pre-signed CDN upload URL.
func (c *Client) GetUploadURL() (*GetUploadURLResponse, error) {
	data, err := c.doPOST("/ilink/bot/getuploadurl", nil)
	if err != nil {
		return nil, fmt.Errorf("get upload url: %w", err)
	}

	var resp GetUploadURLResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal upload url response: %w", err)
	}

	if resp.Ret != 0 {
		return nil, fmt.Errorf("get upload url failed: ret=%d", resp.Ret)
	}

	return &resp, nil
}

// UploadMedia encrypts file data with AES-128-ECB, uploads to CDN,
// and returns the base64-encoded AES key and CDN reference parameters.
func (c *Client) UploadMedia(fileData []byte) (aesKeyBase64 string, uploadResp *GetUploadURLResponse, err error) {
	// Generate random 16-byte AES key
	aesKey := make([]byte, 16)
	if _, err := rand.Read(aesKey); err != nil {
		return "", nil, fmt.Errorf("generate aes key: %w", err)
	}

	// Encrypt with AES-128-ECB + PKCS7 padding
	encrypted, err := encryptAES128ECB(fileData, aesKey)
	if err != nil {
		return "", nil, fmt.Errorf("encrypt file: %w", err)
	}

	// Get pre-signed upload URL
	uploadResp, err = c.GetUploadURL()
	if err != nil {
		return "", nil, err
	}
	if uploadResp.UploadURL == "" {
		return "", nil, fmt.Errorf("empty upload url in response")
	}

	// PUT encrypted data to CDN
	req, err := http.NewRequest(http.MethodPut, uploadResp.UploadURL, bytes.NewReader(encrypted))
	if err != nil {
		return "", nil, fmt.Errorf("create upload request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("upload to cdn: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("cdn upload failed: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	logger.Infof("iLink media uploaded to CDN (%d bytes)", len(encrypted))

	aesKeyBase64 = base64.StdEncoding.EncodeToString(aesKey)
	return aesKeyBase64, uploadResp, nil
}

// encryptAES128ECB encrypts data with AES-128-ECB and PKCS7 padding.
func encryptAES128ECB(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// PKCS7 padding
	blockSize := block.BlockSize()
	padding := blockSize - len(data)%blockSize
	padded := make([]byte, len(data)+padding)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}

	// Encrypt block by block
	ciphertext := make([]byte, len(padded))
	for i := 0; i < len(padded); i += blockSize {
		block.Encrypt(ciphertext[i:i+blockSize], padded[i:i+blockSize])
	}

	return ciphertext, nil
}

// DecryptAES128ECB decrypts data with AES-128-ECB and removes PKCS7 padding.
func DecryptAES128ECB(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	if len(data)%blockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of block size")
	}

	// Decrypt block by block
	plaintext := make([]byte, len(data))
	for i := 0; i < len(data); i += blockSize {
		block.Decrypt(plaintext[i:i+blockSize], data[i:i+blockSize])
	}

	// Remove PKCS7 padding
	padding := int(plaintext[len(plaintext)-1])
	if padding > blockSize || padding == 0 {
		return nil, fmt.Errorf("invalid pkcs7 padding")
	}

	return plaintext[:len(plaintext)-padding], nil
}

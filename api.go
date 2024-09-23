package pullaway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
)

type PushoverClient struct {
	APIURL string
}

func (pc *PushoverClient) GetApiURL() (url.URL, error) {
	purl := "https://api.pushover.net/1"
	if pc != nil && pc.APIURL != "" {
		purl = pc.APIURL
	}

	u, err := url.Parse(purl)
	if err != nil {
		return url.URL{}, err
	}
	return *u, nil
}

func (pc *PushoverClient) Login(username, password, twofa string) (*LoginResponse, error) {
	apiURL, err := pc.GetApiURL()
	if err != nil {
		return nil, err
	}
	return Login(apiURL, username, password, twofa)
}

func (pc *PushoverClient) Register(secret, name string) (*RegistrationResponse, error) {
	apiURL, err := pc.GetApiURL()
	if err != nil {
		return nil, err
	}
	return Register(apiURL, secret, name)
}

type AuthorizedClient struct {
	UserSecret string
	DeviceID   string

	*PushoverClient
}

func NewAuthorizedClient(userSecret, deviceID string) *AuthorizedClient {
	return &AuthorizedClient{
		UserSecret:     userSecret,
		DeviceID:       deviceID,
		PushoverClient: &PushoverClient{},
	}
}

func (ac *AuthorizedClient) DownloadMessages() (*DownloadResponse, error) {
	apiURL, err := ac.GetApiURL()
	if err != nil {
		return nil, err
	}
	return DownloadMessages(apiURL, ac.UserSecret, ac.DeviceID)
}

func (ac *AuthorizedClient) DeleteMessages(id int64) (*DeleteResponse, error) {
	apiURL, err := ac.GetApiURL()
	if err != nil {
		return nil, err
	}
	return DeleteMessages(apiURL, ac.UserSecret, ac.DeviceID, id)
}

func (ac *AuthorizedClient) DownloadAndDeleteMessages() (*DownloadResponse, *DeleteResponse, error) {
	return ac.PushoverClient.DownloadAndDeleteMessages(ac.UserSecret, ac.DeviceID)
}

func (ac *AuthorizedClient) GetAuthorizedListener(l LeveledLogger) *AuthorizedListener {
	return NewAuthorizedListener(ac, l)
}

// Helper method to make HTTP requests
func doRequest(method, urlStr string, body io.Reader, headers map[string]string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching request: %s - %s", resp.Status, respBody)
	}

	return respBody, nil
}

func Login(api url.URL, username, password, twofa string) (*LoginResponse, error) {
	// Build request body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("email", username)
	writer.WriteField("password", password)
	if twofa != "" {
		writer.WriteField("twofa", twofa)
	}
	writer.Close()

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	api.Path = path.Join(api.Path, "/users/login.json")

	respBody, err := doRequest("POST", api.String(), body, headers)
	if err != nil {
		return nil, err
	}

	jsonResponse := &LoginResponse{}
	err = json.Unmarshal(respBody, jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}

	if !jsonResponse.IsValid() {
		return jsonResponse, fmt.Errorf("error logging in: %s", jsonResponse.Error())
	}

	return jsonResponse, nil
}

func Register(api url.URL, secret, name string) (*RegistrationResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("secret", secret)
	writer.WriteField("name", name)
	writer.WriteField("os", "O")
	writer.Close()

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	api.Path = path.Join(api.Path, "devices.json")

	respBody, err := doRequest("POST", api.String(), body, headers)
	if err != nil {
		return nil, err
	}

	jsonResponse := &RegistrationResponse{}
	err = json.Unmarshal(respBody, jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}

	if !jsonResponse.IsValid() {
		return jsonResponse, fmt.Errorf("error registering: %s", jsonResponse.Error())
	}

	return jsonResponse, nil
}

func DownloadMessages(api url.URL, secret, deviceID string) (*DownloadResponse, error) {
	api.Path = path.Join(api.Path, "messages.json")

	q := api.Query()
	q.Set("secret", secret)
	q.Set("device_id", deviceID)
	api.RawQuery = q.Encode()

	headers := map[string]string{}

	respBody, err := doRequest("GET", api.String(), nil, headers)
	if err != nil {
		return nil, err
	}

	jsonResponse := &DownloadResponse{}
	err = json.Unmarshal(respBody, jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}

	if !jsonResponse.IsValid() {
		return jsonResponse, fmt.Errorf("error downloading: %s", jsonResponse.Error())
	}

	return jsonResponse, nil
}

func DeleteMessages(api url.URL, secret, deviceID string, id int64) (*DeleteResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("secret", secret)
	writer.WriteField("message", fmt.Sprintf("%d", id))
	writer.Close()

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	api.Path = path.Join(api.Path, "devices", deviceID, "update_highest_message.json")

	respBody, err := doRequest("POST", api.String(), body, headers)
	if err != nil {
		return nil, err
	}

	jsonResponse := &DeleteResponse{}
	err = json.Unmarshal(respBody, jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}

	if !jsonResponse.IsValid() {
		return jsonResponse, fmt.Errorf("error deleting messages: %s", jsonResponse.Error())
	}

	return jsonResponse, nil
}

func (pc *PushoverClient) DownloadAndDeleteMessages(secret, deviceID string) (*DownloadResponse, *DeleteResponse, error) {
	apiURL, err := pc.GetApiURL()
	if err != nil {
		return nil, nil, err
	}

	dr, err := DownloadMessages(apiURL, secret, deviceID)
	if err != nil {
		return dr, nil, err
	}

	dm, err := DeleteMessages(apiURL, secret, deviceID, dr.MaxID())
	if err != dr {
		return dr, dm, err
	}

	return dr, dm, nil
}

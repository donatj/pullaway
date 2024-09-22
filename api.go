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
	"time"
)

type PushoverClient struct {
	APIURL string
}

func (pc *PushoverClient) getApiURL() string {
	if pc.APIURL == "" {
		return "https://api.pushover.net/1"
	}
	return pc.APIURL
}

type AuthorizedClient struct {
	UserSecret string
	DeviceID   string

	*PushoverClient
}

func (pc *PushoverClient) Login(username, password, twofa string) (*LoginResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("email", username)
	writer.WriteField("password", password)
	if twofa != "" {
		writer.WriteField("twofa", twofa)
	}
	writer.Close()

	client := &http.Client{}

	req, err := http.NewRequest("POST", path.Join(pc.getApiURL(), "/users/login.json"), body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching request: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching request: %s - %s", resp.Status, respBody)
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

func (pc *PushoverClient) Register(secret string) (*RegistrationResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("secret", secret)
	writer.WriteField("name", fmt.Sprintf("pullaway-%d", time.Now().Unix()))
	writer.WriteField("os", "O")
	writer.Close()

	client := &http.Client{}

	req, err := http.NewRequest("POST", path.Join(pc.getApiURL(), "/devices.json"), body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching request: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching request: %s - %s", resp.Status, respBody)
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

func (ac *AuthorizedClient) Download() (*DownloadResponse, error) {
	return ac.PushoverClient.DownloadMessages(ac.UserSecret, ac.DeviceID)
}

func (pc *PushoverClient) DownloadMessages(secret, deviceID string) (*DownloadResponse, error) {

	client := &http.Client{}

	u, err := url.Parse(path.Join(pc.getApiURL(), "/messages.json"))
	if err != nil {
		panic(err)
	}

	q := u.Query()
	q.Set("secret", secret)
	q.Set("device_id", deviceID)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	parseFormErr := req.ParseForm()
	if parseFormErr != nil {
		fmt.Println(parseFormErr)
	}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching request: %s - %s", resp.Status, respBody)
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

func (ac *AuthorizedClient) DeleteMessages(id int64) (*DeleteResponse, error) {
	return ac.PushoverClient.DeleteMessages(ac.UserSecret, ac.DeviceID, id)
}

func (pc *PushoverClient) DeleteMessages(secret, deviceID string, id int64) (*DeleteResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("secret", secret)
	writer.WriteField("message", fmt.Sprintf("%d", id))
	writer.Close()

	client := &http.Client{}

	u, err := url.Parse(pc.getApiURL())
	if err != nil {
		return nil, fmt.Errorf("error parsing url: %w", err)
	}

	u.Path = path.Join("1/devices", deviceID, "update_highest_message.json")

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching request: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching request: %s - %s", resp.Status, respBody)
	}

	jsonResponse := &DeleteResponse{}
	err = json.Unmarshal(respBody, jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}

	if !jsonResponse.IsValid() {
		return jsonResponse, fmt.Errorf("error registering: %s", jsonResponse.Error())
	}

	return jsonResponse, nil
}

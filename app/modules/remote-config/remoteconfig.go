package remoteconfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

// GetEtag return etag string and error
func (h *Handler) GetEtag() (string, error) {
	//Set up new Client HTTP
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, h.RemoteConfigURL, nil)
	if err != nil {
		return "", err
	}

	// Set Authorization Header
	req.Header.Set("Authorization", "Bearer "+h.Token.AccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	// if resp.Status is 200
	if resp.StatusCode == http.StatusOK {
		return resp.Header["Etag"][0], nil
	}
	return "", nil
}

// GetToken return error
func (h *Handler) GetToken() error {
	// b, err := ioutil.ReadFile(h.CredentialsFile)
	// if err != nil {
	// 	return err
	// }
	var c = struct {
		Email      string `json:"client_email"`
		PrivateKey string `json:"private_key"`
	}{}
	json.Unmarshal([]byte(h.CredentialsFile), &c)
	config := &jwt.Config{
		Email:      c.Email,
		PrivateKey: []byte(c.PrivateKey),
		Scopes: []string{
			"https://www.googleapis.com/auth/firebase.remoteconfig",
		},
		TokenURL: google.JWTTokenURL,
	}
	token, err := config.TokenSource(oauth2.NoContext).Token()
	if err != nil {
		return err
	}
	h.Token = token
	return nil
}

// Init to init remote config
func (h *Handler) Init() error {
	h.CredentialsFile = os.Getenv("GOOGLE_CREDENTIALS")
	h.ConfigFile = os.Getenv("configFile")
	h.ProjectID = os.Getenv("PROJECT_ID")
	baseURL := "https://firebaseremoteconfig.googleapis.com"
	remoteConfigEndpoint := "v1/projects/" + h.ProjectID + "/remoteConfig"
	h.RemoteConfigURL = baseURL + "/" + remoteConfigEndpoint

	err := h.GetToken()
	if err != nil {
		return err
	}
	fmt.Println("Remote Config Init Successful")
	return nil
}
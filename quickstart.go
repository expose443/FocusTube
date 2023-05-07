package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

const missingClientSecretsMessage = `
please configure OAuth 2.0
`

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("unable to get path to cached credential file. %v", err)
	}
	token, err := tokenFromFile(cacheFile)
	if err != nil {
		token = getTokenFromWeb(config)
		saveToken(cacheFile, token)
	}
	return config.Client(ctx, token)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("go to the following link in your browser then type the "+"authorization code: \n%v\n", authURL)
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("unable to read authorization code %v", err)
	}
	token, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("unable to retrieve token from web %v", err)
	}
	return token
}

func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("saving credintial file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		log.Fatalf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func handleError(err error, message string) {
	if message == "" {
		message = "error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	return t, err
}

func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credintials")
	err = os.MkdirAll(tokenCacheDir, 0o700)
	return filepath.Join(tokenCacheDir, url.QueryEscape("youtube-go-api.json")), err
}

func channelsListByUsername(service *youtube.Service, part []string, forUsername string) {
	call := service.Channels.List(part)
	call = call.ForUsername(forUsername)

	response, err := call.Do()
	handleError(err, "")
	fmt.Println(fmt.Sprintf("This channels's ID is %s. Its title is '%s', "+"and it has %d views.",
		response.Items[0].Id,
		response.Items[0].Snippet.Title,
		response.Items[0].Statistics.ViewCount))
}

func main() {
	ctx := context.Background()

	b, err := os.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, youtube.YoutubeReadonlyScope)
	if err != nil {
		log.Fatalf("unable to parse client secret file to config: %v", err)
	}

	client := getClient(ctx, config)

	service, err := youtube.New(client)
	handleError(err, "Error creating YouTube client")
	channelsListByUsername(service, []string{"snippet", "contentDetails", "statistics"}, "GoogleDevelopers")
}

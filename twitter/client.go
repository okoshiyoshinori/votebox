package twitter

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/okoshiyoshinori/votebox/logger"

	"github.com/okoshiyoshinori/votebox/model"

	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/okoshiyoshinori/votebox/config"
)

func GetConnect() *oauth.Client {
	return &oauth.Client{
		TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authorize",
		TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
		Credentials: oauth.Credentials{
			Token:  config.APConfig.TwitterConsumerKey,
			Secret: config.APConfig.TwitterConsumerSecret,
		},
	}
}

func GetAccessToken(rt *oauth.Credentials, oauthVerifer string) (*oauth.Credentials, error) {
	oc := GetConnect()
	at, _, err := oc.RequestToken(nil, rt, oauthVerifer)
	return at, err
}

func GetMe(at *oauth.Credentials, user *model.Account) error {
	oc := GetConnect()

	v := url.Values{}
	v.Set("include_email", "true")

	resp, err := oc.Get(nil, at, "https://api.twitter.com/1.1/account/verify_credentials.json", v)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return errors.New("Twitter is unavailable")
	}

	if resp.StatusCode >= 400 {
		return errors.New("Twitter request is invalid")
	}

	err = json.NewDecoder(resp.Body).Decode(user)
	if err != nil {
		return err
	}

	return nil
}

func InitClient() *anaconda.TwitterApi {
	anaconda.SetConsumerKey(config.APConfig.TwitterConsumerKey)
	anaconda.SetConsumerSecret(config.APConfig.TwitterConsumerSecret)
	var client *anaconda.TwitterApi
	//if isAnonymous {
	client = anaconda.NewTwitterApi(config.APConfig.TwitterAccessToken, config.APConfig.TwitterAccessTokenSecret)
	//} else {
	//	client = anaconda.NewTwitterApi(user.Token, user.Secret)
	//	}
	return client
}

func PostTweet(client *anaconda.TwitterApi, box *model.Box) {
	text := "投票箱が作成されました。\n" + "#votebox #投票箱 #アンケート" + " " + config.APConfig.TweetBaseUrl + "" + box.Hid.ID
	_, err := client.PostTweet(text, nil)
	if err != nil {
		logger.Error.Println(err)
	}
}

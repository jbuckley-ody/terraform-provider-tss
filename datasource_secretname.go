package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/thycotic/tss-sdk-go/server"
)

func dataSourceSecretNameRead(d *schema.ResourceData, meta interface{}) error {
	path := d.Get("path").(string)
	config := meta.(server.Configuration)
	token := getToken(config.ServerURL, config.Credentials.Username, config.Credentials.Password)
	secret := getSecret(config.ServerURL, token, path)
	d.Set("value", secret)
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}

func dataSourceSecretName() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSecretNameRead,

		Schema: map[string]*schema.Schema{
			"value": {
				Computed:    true,
				Description: "the value of the field of the secret",
				Type:        schema.TypeInt,
			},
			"path": {
				Description: "the path to the secret",
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func getToken(serverUrl string, username string, password string) string {
	uri := serverUrl + "oauth2/token"
	method := "POST"

	payload := strings.NewReader("grant_type=password&username=" + url.QueryEscape(username) + "&password=" + url.QueryEscape(password))

	client := &http.Client{}
	req, err := http.NewRequest(method, uri, payload)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	type Token struct {
		Access_token  string
		Token_type    string
		Expires_in    int
		Refresh_token string
	}

	var token Token
	json.Unmarshal(body, &token)

	return token.Access_token
}

func getSecret(serverUrl string, token string, path string) int {
	uri := serverUrl + "api/v1/secrets/0?secretPath=" + url.QueryEscape(path)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, uri, nil)

	if err != nil {
		fmt.Println(err)
		return -1
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	type Secret struct {
		Id int
	}

	var secret Secret
	json.Unmarshal(body, &secret)

	return secret.Id
}

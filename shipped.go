package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"secondbit.org/adn"
	"strings"
)

const configFileName = ".shipped_config"
const clientID = "NxN3hzTmYgjQ4j8ZB5WvH232EqqnVY78"

func hasConfigFile(name string) bool {
	_, err := os.Stat(name)
	if err != nil {
		if e, ok := err.(*os.PathError); ok && os.IsNotExist(e) {
			return false
		}
	}
	return true
}

type Config struct {
	Token string `json:"token"`
}

func (c *Config) save(path string) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, b, 0600)
	if err != nil {
		return err
	}
	return nil
}

func loadConfig(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := new(Config)
	err = json.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func getShippedMessage() (string, error) {
	fmt.Println("What did you ship?")
	reader := bufio.NewReader(os.Stdin)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	msg = strings.TrimSpace(msg)
	return msg, nil
}

func postShippedMessage(client *adn.ADN, msg string) error {
	post := adn.Post{
		Text: msg,
	}
	_, err := client.CreatePost(post)
	return err
}

func main() {
	usr, err := user.Current()
	if err != nil {
		panic(err.Error())
	}
	confPath := usr.HomeDir + "/" + configFileName
	client := adn.NewClient(clientID, "", "http://localhost:8000/", []string{adn.SCOPE_WRITE_POST})
	var conf *Config
	if hasConfigFile(confPath) {
		conf, err = loadConfig(confPath)
		if err != nil {
			panic(err.Error())
		}
	}
	if conf == nil || conf.Token == "" {
		auth_url, err := client.GetClientSideAuthURL()
		if err != nil {
			panic(err)
		}
		fmt.Println("Please open " + auth_url + " in your browser and approve access.")
		token, err := client.ListenForClientSideAuth()
		if err != nil {
			panic(err)
		}
		conf = &Config{Token: token}
		err = conf.save(confPath)
		if err != nil {
			fmt.Println("Error occurred while saving configuration: " + err.Error())
		}
	}
	if conf == nil || conf.Token == "" {
		panic("Access token not set.")
	}
	client.Token = conf.Token
	msg, err := getShippedMessage()
	if err != nil {
		panic(err)
	}
	err = postShippedMessage(client, msg)
	if err != nil {
		panic(err)
	}
	fmt.Println("Congratulations!")
}

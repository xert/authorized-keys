package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"path/filepath"

	"code.google.com/p/go.crypto/ssh"
	yaml "gopkg.in/yaml.v1"
)

type authorizedKeys struct {
	Users  map[string]map[string][]string
	Access map[string]map[string][]string
}

type configData struct {
	Log  string
	Data string
}

// Allow only lowercase letters, dot, underscore or hyphen
func safestring(s string) bool {
	r, err := regexp.Compile("[-.a-z_]+")
	if err != nil {
		log.Fatalln(err)
	}

	return r.MatchString(s)
}

// Error during program init
// print error + exit with code 2
func initError(s string) {
	os.Stderr.WriteString(s + "\n")
	os.Exit(2)
}

func main() {
	if len(os.Args) != 2 {
		initError("Need exactly one argument - username")
	}

	// Load config file
	const configfile = "/usr/local/etc/authorized-keys.conf"
	configsource, err := ioutil.ReadFile(configfile)
	config := configData{}
	if err != nil {
		os.Stderr.WriteString("Error opening config file " + configfile + ", using defaults\n")
		programDir := filepath.Dir(os.Args[0])
		config.Data = programDir + "/authorized-keys.yaml"
		config.Log = programDir + "/authorized-keys.log"
	} else {
		err = yaml.Unmarshal(configsource, &config)
		if err != nil {
			initError(err.Error())
		}
	}

	// Open log file
	l, err := os.OpenFile(config.Log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		initError("error opening log file: '" + config.Log + "'" + err.Error())
	}
	defer l.Close()
	log.SetOutput(io.MultiWriter(l, os.Stderr))

	// Username
	username := os.Args[1]
	if !safestring(username) {
		log.Fatalln("User '" + username + " is not safe string")
	}

	// Hostname (server name)
	hostname, _ := os.Hostname()
	if !safestring(hostname) {
		log.Fatalln("Host '" + hostname + "' is not safe string")
	}

	// Data file
	source, err := ioutil.ReadFile(config.Data)
	if err != nil {
		initError("Can't read data file " + config.Data)
	}
	authorizedKeys := authorizedKeys{}
	err = yaml.Unmarshal(source, &authorizedKeys)
	if err != nil {
		initError("Cannot parse yaml file " + err.Error())
	}

	// Debug
	log.Println("Requested authorized_keys for server '" + hostname + "' and user '" + username + "'")

	// Find aliases for server/user
	users, ok := authorizedKeys.Access[hostname][username]
	if !ok {
		log.Fatalln("No users found")
	}

	// Find all keys for aliases
	var keys []string
	for _, user := range users {
		k, ok := authorizedKeys.Users[user]["keys"]
		if !ok {
			log.Fatalln("User '" + user + "' doesn't have any keys defined")
		}
		// Validate and append each key
		for index, key := range k {
			_, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(key))

			if err == nil {
				keys = append(keys, key)
				log.Printf("Using key %q for user %q", comment, user)
			} else {
				log.Printf("Skipping key %d for user %q", index, user)
			}
		}
	}

	// Output for sshd
	os.Stdout.WriteString(strings.Join(keys, "\n"))
}

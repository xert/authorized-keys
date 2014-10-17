// Authorized keys SSHD helper
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"path/filepath"

	"crypto/md5"

	"code.google.com/p/go.crypto/ssh"
	docopt "github.com/docopt/docopt-go"
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
func initError(s string, code int) {
	os.Stderr.WriteString(s + "\n")
	os.Exit(code)
}

var builddate string

func main() {
	usage := `Authorized Keys.

Usage:
  authorized-keys <user> [--force-server=<server>] [--test]
  authorized-keys -h | --help
  authorized-keys -v | --version

Options:
  -h --help                Show this screen.
  -v --version             Show version.
  --force-server=<server>  Force server name.
  --test                  Test mode - no logging, just print users and key fingerprints.
`
	arguments, err := docopt.Parse(usage, nil, true, "Authorized Keys build "+builddate, false)

	if err != nil {
		initError(usage, 2)
	}

	var test = false
	if arguments["--test"] == true {
		test = true
	}

	if arguments["<user>"] == nil {
		initError("No user defined", 3)
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
			initError(err.Error(), 4)
		}
	}

	// Open log file
	if !test {
		l, err := os.OpenFile(config.Log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
		if err != nil {
			initError("error opening log file: '"+config.Log+"'"+err.Error(), 5)
		}
		defer l.Close()
		log.SetOutput(io.MultiWriter(l, os.Stderr))
	} else {
		log.SetOutput(os.Stderr)
	}

	// Username
	username := fmt.Sprintf("%s", arguments["<user>"])
	if !safestring(username) {
		log.Fatalln("User '" + username + "' is not safe string")
	}

	// Hostname (server name)
	var hostname string
	if arguments["--force-server"] == nil {
		hostname, _ = os.Hostname()
	} else {
		hostname = fmt.Sprintf("%s", arguments["--force-server"])
	}

	if !safestring(hostname) {
		log.Fatalln("Host '" + hostname + "' is not safe string")
	}

	// Data file
	source, err := ioutil.ReadFile(config.Data)
	if err != nil {
		initError("Can't read data file "+config.Data, 5)
	}
	authorizedKeys := authorizedKeys{}
	err = yaml.Unmarshal(source, &authorizedKeys)
	if err != nil {
		initError("Cannot parse yaml file "+err.Error(), 6)
	}

	// Debug
	log.Println("Requested authorized_keys for server '" + hostname + "' and user '" + username + "'")

	// Find server/user
	_, ok := authorizedKeys.Access[hostname][username]
	if !ok {
		log.Fatalln("Server '" + hostname + "' not found")
	}

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
			publickey, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(key))

			if err == nil {
				keys = append(keys, key)
				if test {
					fingerprint := fmt.Sprintf("%x", md5.Sum(publickey.Marshal()))
					log.Printf("Using key %q for user %q. Fingerprint: %s", comment, user, formatFingerPrint(fingerprint))
				} else {
					log.Printf("Using key %q for user %q", comment, user)
				}

			} else {
				log.Printf("Skipping key %d for user %q", index, user)
			}
		}
	}

	// Output for sshd
	if !test {
		os.Stdout.WriteString(strings.Join(keys, "\n"))
	}
}

func formatFingerPrint(f string) string {
	var fprint []string
	for i := 0; i < len(f); i = i + 2 {
		fprint = append(fprint, f[i:i+2])
	}
	return strings.Join(fprint, ":")
}

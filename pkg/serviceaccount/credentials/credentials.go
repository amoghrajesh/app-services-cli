package credentials

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
)

// Templates
var (
	templateProperties = heredoc.Doc(`
	## Generated by rhoas cli
	user=%v
	password=%v
	`)

	templateEnv = heredoc.Doc(`
	## Generated by rhoas cli
	USER=%v
	PASSWORD=%v
	`)

	templateJSON = heredoc.Doc(`
	{ 
		"user":"%v", 
		"password":"%v" 
	}`)

	templateSecret = heredoc.Doc(`
	kind: Secret
	apiVersion: v1
	metadata:
	  name: "rhoas-service-account-secret"
	stringData:
	  clientID: "%v"
	  clientSecret: "%v"
	type: Opaque
	`)
)

// Credentials is a type which represents the SASL/Plain credentials
// for a service account
type Credentials struct {
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

// AbsolutePath returns the absolute path for the credentials file
// returning a default location based on the output format if customLocation
// is empty
func AbsolutePath(outputFormat string, customLocation string) (filePath string) {
	filePath = customLocation
	if filePath == "" {
		switch outputFormat {
		case "env":
			filePath = ".env"
		case "properties":
			filePath = "kafka.properties"
		case "json":
			filePath = "credentials.json"
		case "kube":
			filePath = "credentials.yaml"
		}
	}

	pwd, err := os.Getwd()
	if err != nil {
		pwd = "./"
	}

	filePath = filepath.Join(pwd, filePath)

	return filePath
}

// Write saves the credentials to a file
// in the specified output format
func Write(output string, fileName string, credentials *Credentials) error {
	fileTemplate := getFileFormat(output)
	fileBody := fmt.Sprintf(fileTemplate, credentials.ClientID, credentials.ClientSecret)

	fileData := []byte(fileBody)

	return ioutil.WriteFile(fileName, fileData, 0600)
}

func getFileFormat(output string) (format string) {
	switch output {
	case "env":
		format = templateEnv
	case "properties":
		format = templateProperties
	case "json":
		format = templateJSON
	case "kube":
		format = templateSecret
	}

	return format
}

package config

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func (s *Configuration) NewContext(name string) {
	name = strings.ToLower(name) // forces lowercase contexts
	answers := GrafanaConfig{
		ConnectionSettings: &ConnectionSettings{
			MatchingRules: make([]RegexMatchesList, 0),
		},
	}
	promptAnswers := struct {
		AuthType   string
		Folders    string
		DSUser     string
		DSPassword string
	}{}
	// Setup question that drive behavior
	behaviorQuestions := []*survey.Question{
		{
			Name: "AuthType",
			Prompt: &survey.Select{
				Message: "Will you be using a Token, BasicAuth, or both?",
				Options: []string{"token", "basicauth", "both"},
				Default: "basicauth",
			},
		},
		{
			Name:   "DSUser",
			Prompt: &survey.Input{Message: "Please enter your datasource default username"},
		},
		{
			Name:   "DSPassword",
			Prompt: &survey.Password{Message: "Please enter your datasource default password"},
		},
		{
			Name:   "Folders",
			Prompt: &survey.Input{Message: "List the folders you wish to manage (example: folder1,folder2)? (Blank for General)?"},
		},
	}
	err := survey.Ask(behaviorQuestions, &promptAnswers)
	if err != nil {
		log.Fatal("Failed to get valid answers to generate a new context")
	}

	// Set Watched Folders
	foldersList := strings.Split(promptAnswers.Folders, ",")
	if len(foldersList) > 0 && foldersList[0] != "" {
		answers.MonitoredFolders = foldersList
	} else {
		answers.MonitoredFolders = []string{"General"}
	}

	// Setup grafana required field based on responses
	questions := []*survey.Question{
		{
			Name:   "URL",
			Prompt: &survey.Input{Message: "What is the Grafana URL include http(s)?"},
		},
		{
			Name:     "OutputPath",
			Prompt:   &survey.Input{Message: "Destination Folder?"},
			Validate: survey.Required,
		},
	}

	if promptAnswers.AuthType == "both" || promptAnswers.AuthType == "token" {
		questions = append(questions, &survey.Question{
			Name:     "APIToken",
			Prompt:   &survey.Input{Message: "Please enter your API Token"},
			Validate: survey.Required,
		})
	}

	if promptAnswers.AuthType == "both" || promptAnswers.AuthType == "basicauth" {
		questions = append(questions, &survey.Question{
			Name:     "UserName",
			Prompt:   &survey.Input{Message: "Please enter your grafana admin Username"},
			Validate: survey.Required,
		})
		questions = append(questions, &survey.Question{
			Name:     "Password",
			Prompt:   &survey.Password{Message: "Please enter your grafana admin Password"},
			Validate: survey.Required,
		})

	}

	err = survey.Ask(questions, &answers)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Set Default Datasource
	if promptAnswers.DSUser != "" && promptAnswers.DSPassword != "" {
		ds := GrafanaConnection{
			"user":              promptAnswers.DSUser,
			"basicAuthPassword": promptAnswers.DSPassword,
		}

		location := filepath.Join(answers.OutputPath, SecureSecretsResource)
		err = os.MkdirAll(location, 0750)
		if err != nil {
			log.Fatalf("unable to create default secret location.  location: %s, %v", location, err)
		}
		data, err := json.MarshalIndent(&ds, "", "    ")
		if err != nil {
			log.Fatalf("unable to turn map into json representation.  location: %s, %v", location, err)
		}
		secretFileLocation := filepath.Join(location, "default.json")
		err = os.WriteFile(secretFileLocation, data, 0600)
		if err != nil {
			log.Fatalf("unable to write secret default file.  location: %s, %v", secretFileLocation, err)
		}
		answers.ConnectionSettings.MatchingRules = []RegexMatchesList{
			{
				Rules: []MatchingRule{
					{
						Field: "name",
						Regex: ".*",
					},
				},
				SecureData: "default.json",
			},
		}

	}

	contextMap := s.GetGDGConfig().GetContexts()
	contextMap[name] = &answers
	s.GetGDGConfig().ContextName = name

	err = s.SaveToDisk(false)
	if err != nil {
		log.Fatal("could not save configuration.")
	}
	slog.Info("New configuration has been created", "newContext", name)
}

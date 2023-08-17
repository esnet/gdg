package config

import (
	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"strings"
)

func (s *Configuration) NewContext(name string) {

	name = strings.ToLower(name) // forces lowercase contexts
	answers := GrafanaConfig{
		DataSourceSettings: &ConnectionSettings{
			MatchingRules: make([]RegexMatchesList, 0),
		},
	}
	promptAnswers := struct {
		AuthType   string
		Folders    string
		DSUser     string
		DSPassword string
	}{}
	//Setup question that drive behavior
	var behaviorQuestions = []*survey.Question{
		{
			Name: "AuthType",
			Prompt: &survey.Select{
				Message: "Will you be using a Token, BasicAuth, or both?",
				Options: []string{"token", "basicauth", "both"},
				Default: "basicauth",
			},
		},
		{
			Name:   "Folders",
			Prompt: &survey.Input{Message: "List the folders you wish to manage (example: folder1,folder2)? (Blank for General)?"},
		},
		{
			Name:   "DSUser",
			Prompt: &survey.Input{Message: "Please enter your datasource default username"},
		},
		{
			Name:   "DSPassword",
			Prompt: &survey.Password{Message: "Please enter your datasource default password"},
		},
	}
	err := survey.Ask(behaviorQuestions, &promptAnswers)
	if err != nil {
		log.Fatal("Failed to get valid answers to generate a new context")
	}

	//Set Watched Folders
	foldersList := strings.Split(promptAnswers.Folders, ",")
	if len(foldersList) > 0 && foldersList[0] != "" {
		answers.MonitoredFolders = foldersList
	} else {
		answers.MonitoredFolders = []string{"General"}
	}
	//Set Default Datasource
	if promptAnswers.DSUser != "" && promptAnswers.DSPassword != "" {
		ds := GrafanaConnection{
			User:     promptAnswers.DSUser,
			Password: promptAnswers.DSPassword,
		}
		answers.DataSourceSettings.MatchingRules = []RegexMatchesList{
			{
				Rules: []MatchingRule{
					{
						Field: "name",
						Regex: ".*",
					},
				},
				Auth: &ds,
			},
		}

	}

	//Setup grafana required field based on responses
	var questions = []*survey.Question{
		{
			Name:   "URL",
			Prompt: &survey.Input{Message: "What is the Grafana URL include http(s)?"},
		},
		{
			Name:   "OutputPath",
			Prompt: &survey.Input{Message: "Destination Folder?"},
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
			Prompt:   &survey.Input{Message: "Please enter your admin UserName"},
			Validate: survey.Required,
		})
		questions = append(questions, &survey.Question{
			Name:     "Password",
			Prompt:   &survey.Password{Message: "Please enter your admin Password"},
			Validate: survey.Required,
		})

	}

	err = survey.Ask(questions, &answers)
	if err != nil {
		log.Fatal(err.Error())
	}

	contextMap := s.GetAppConfig().GetContexts()
	contextMap[name] = &answers
	s.GetAppConfig().ContextName = name

	err = s.SaveToDisk(false)
	if err != nil {
		log.Fatal("could not save configuration.")
	}
	log.Infof("New configuration %s has been created", name)

}

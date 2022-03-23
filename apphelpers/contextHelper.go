package apphelpers

import (
	"fmt"
	"os"
	"strings"

	// "github.com/labstack/gommon/log"
	"github.com/AlecAivazis/survey/v2"
	"github.com/esnet/gdg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v2"
)

//GetContext returns the name of the selected context
func GetContext() string {
	name := config.Config().ViperConfig().GetString("context_name")
	return strings.ToLower(name)
}

func NewContext(name string) {
	name = strings.ToLower(name) // forces lowercase contexts
	answers := config.GrafanaConfig{
		DataSourceSettings: make(map[string]*config.GrafanaDataSource),
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
		ds := config.GrafanaDataSource{
			User:     promptAnswers.DSUser,
			Password: promptAnswers.DSPassword,
		}
		answers.DataSourceSettings["default"] = &ds
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
		answers.AdminEnabled = true
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

	v := config.Config().ViperConfig()
	contextMap := config.Config().Contexts()

	contextMap[name] = &answers
	v.Set("contexts", contextMap)
	err = v.WriteConfig()
	if err != nil {
		log.Fatal("could not save configuration.")
	}
	SetContext(name)
	log.Infof("New configuration %s has been created", name)

}

//ShowContext displays the selected context
func ShowContext(ctx string) {
	grafana := GetCtxGrafanaConfig(ctx)
	d, err := yaml.Marshal(grafana)
	if err != nil {
		log.Info("Failed to serialize context")
		os.Exit(1)
	}
	fmt.Printf("---%s:\n%s\n\n", ctx, string(d))

}

//ClearContexts clear all contexts except a simple running example
// (required for app not to error out)
func ClearContexts() {
	v := config.Config().ViperConfig()
	newContext := make(map[string]*config.GrafanaConfig)
	newContext["example"] = &config.GrafanaConfig{
		APIToken: "dummy",
	}
	v.Set("context_name", "example")
	v.Set("contexts", newContext)
	err := v.WriteConfig()
	if err != nil {
		log.Fatal("could not save configuration.")
	}
}

func CopyContext(src, dest string) {
	v, contexts := getContextReferences()
	srcCtx := v.GetStringMap(fmt.Sprintf("contexts.%s", src))
	//Validate context
	if len(contexts) == 0 {
		log.Fatal("Cannot set context.  No valid configuration found in importer.yml")
	}
	contexts[dest] = srcCtx
	v.Set("contexts", contexts)
	SetContext(dest)
	log.Infof("Copied %s context to %s please check your config to confirm", src, dest)
}

//SetContext will try to find the specified context, if it exists in the file, will re-write the importer.yml
//with the selected context
func SetContext(context string) {
	v, _ := getContextReferences()
	m := config.Config().Contexts()
	if len(m) == 0 {
		log.Fatal("Cannot set context.  No valid configuration found in importer.yml")
	}
	v.Set("context_name", context)
	err := v.WriteConfig()
	if err != nil {
		log.Fatal("could not save configuration.")
	}

}

//getContextReferences Helper method to get viper and context map
func getContextReferences() (*viper.Viper, map[string]interface{}) {
	v := config.Config().ViperConfig()
	contexts := config.Config().ViperConfig().GetStringMap("contexts")

	return v, contexts

}

//DeleteContext Delete a specific
func DeleteContext(context string) {
	activeCtx := GetContext()
	if activeCtx == strings.ToLower(context) {
		log.Fatalf("Cannot delete context since it's currently active, please change context before deleting %s", context)
	}
	v, contextMap := getContextReferences()
	delete(contextMap, context)
	v.Set("contexts", contextMap)
	err := v.WriteConfig()
	if err != nil {
		log.Fatal("could not save configuration.")
	}
}

//GetContexts returns all available contexts
func GetContexts() []string {
	contextMap := config.Config().ViperConfig().GetStringMap("contexts")
	return funk.Keys(contextMap).([]string)
}

//GetCtxGrafanaConfig returns the selected context or terminates app if not found
func GetCtxGrafanaConfig(name string) *config.GrafanaConfig {
	val, ok := config.Config().Contexts()[name]
	if ok {
		return val
	} else {
		log.Error("Context is not found.  Please check your config")
		os.Exit(1)
	}

	return nil
}

//GetCtxDefaultGrafanaConfig returns the default aka. selected grafana config
func GetCtxDefaultGrafanaConfig() *config.GrafanaConfig {
	return GetCtxGrafanaConfig(GetContext())
}

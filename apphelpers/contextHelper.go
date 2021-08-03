package apphelpers

import (
	"fmt"
	"os"

	// "github.com/labstack/gommon/log"
	"github.com/netsage-project/grafana-dashboard-manager/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/thoas/go-funk"
)

//GetContext returns the name of the selected context
func GetContext() string {
	name := config.Config().ViperConfig().GetString("context_name")
	return name
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
	v, contextMap := getContextReferences()
	m := contextMap[context]
	if len(m.(map[string]interface{})) == 0 {
		log.Fatal("Cannot set context.  No valid configuration found in importer.yml")
	}
	v.Set("context_name", context)
	v.WriteConfig()
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
	if activeCtx == context {
		log.Fatalf("Cannot delete context since it's currently active, please change context before deleting %s", context)
	}
	v, contextMap := getContextReferences()
	delete(contextMap, context)
	v.Set("contexts", contextMap)
	v.WriteConfig()
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

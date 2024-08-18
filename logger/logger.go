package logger

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"log"
)

func InitLogging() *logrus.Logger {

	logrusInstance := logrus.New()
	logrusInstance.Formatter = &logrus.JSONFormatter{
		// disable, as we set our own
		DisableTimestamp: true,
	}

	//bugsnag.Configure(bugsnag.Configuration{
	//	// Your Bugsnag project API key
	//	APIKey: viper.GetString("bugsnag_api_key"),
	//	// The development stage of your application build, like "alpha" or
	//	// "production"
	//	ReleaseStage: "production",
	//	// The import paths for the Go packages containing your source files
	//	ProjectPackages: []string{"main", "github.com/org/myapp"},
	//
	//	Logger: logrusInstance,
	//})

	return logrusInstance
}

func LogInfo(x interface{}) {
	s, _ := json.MarshalIndent(x, "", "\t")
	log.Println(string(s))
}

func LogError(err error) {
	//bugsnag.Notify(err)
	log.Println(err)
}
func LogFatal(err error) {
	//bugsnag.Notify(err)
	log.Fatalln(err)
}

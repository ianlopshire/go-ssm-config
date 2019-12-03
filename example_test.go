package ssmconfig_test

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	ssmconfig "github.com/ianlopshire/go-ssm-config"
)

func ExampleProcess() {
	// Assuming the following Parameter Store parameters:
	//
	// | Name                         | Value                | Type         | Key ID        |
	// | ---------------------------- | -------------------- | ------------ | ------------- |
	// | /example_service/prod/debug  | false                | String       | -             |
	// | /example_service/prod/port   | 8080                 | String       | -             |
	// | /example_service/prod/user   | Ian                  | String       | -             |
	// | /example_service/prod/rate   | 0.5                  | String       | -             |
	// | /example_service/prod/secret | zOcZkAGB6aEjN7SAoVBT | SecureString | alias/aws/ssm |

	type Config struct {
		Debug  bool    `smm:"debug" default:"true"`
		Port   int     `smm:"port"`
		User   string  `smm:"user"`
		Rate   float32 `smm:"rate"`
		Secret string  `smm:"secret" required:"true"`
	}

	var c Config
	err := ssmconfig.Process("/example_service/prod/", &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	format := "Debug: %v\nPort: %d\nUser: %s\nRate: %f\nSecret: %s\n"
	_, err = fmt.Printf(format, c.Debug, c.Port, c.User, c.Rate, c.Secret)
	if err != nil {
		log.Fatal(err.Error())
	}
}

// ExampleProcess_multipleEnvironments demonstrates how ssmcofig can be used to configure
// multiple environments.
func ExampleProcess_multipleEnvironments() {
	// Assuming the following Parameter Store parameters:
	//
	// | Name                         | Value                | Type         | Key ID        |
	// | ---------------------------- | -------------------- | ------------ | ------------- |
	// | /example_service/prod/debug  | false                | String       | -             |
	// | /example_service/prod/port   | 8080                 | String       | -             |
	// | /example_service/prod/user   | Ian                  | String       | -             |
	// | /example_service/prod/rate   | 0.5                  | String       | -             |
	// | /example_service/prod/secret | zOcZkAGB6aEjN7SAoVBT | SecureString | alias/aws/ssm |
	// | ---------------------------- | -------------------- | ------------ | ------------- |
	// | /example_service/test/debug  | true                 | String       | -             |
	// | /example_service/test/port   | 8080                 | String       | -             |
	// | /example_service/test/user   | Ian                  | String       | -             |
	// | /example_service/test/rate   | 0.9                  | String       | -             |
	// | /example_service/test/secret | TBVoAS7NjEa6BGAkZcOz | SecureString | alias/aws/ssm |

	type Config struct {
		Debug  bool    `smm:"debug" default:"true"`
		Port   int     `smm:"port"`
		User   string  `smm:"user"`
		Rate   float32 `smm:"rate"`
		Secret string  `smm:"secret" required:"true"`
	}

	// An environment variable is used to set the config path. In this example it would be
	// set to `/example_service/test/` for test and `/example_service/prod/` for
	// production.
	configPath := os.Getenv("CONFIG_PATH")

	var c Config
	err := ssmconfig.Process(configPath, &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	format := "Debug: %v\nPort: %d\nUser: %s\nRate: %f\nSecret: %s\n"
	_, err = fmt.Printf(format, c.Debug, c.Port, c.User, c.Rate, c.Secret)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func ExampleProvider_Process() {
	// Assuming the following Parameter Store parameters:
	//
	// | Name                         | Value                | Type         | Key ID        |
	// | ---------------------------- | -------------------- | ------------ | ------------- |
	// | /example_service/prod/debug  | false                | String       | -             |
	// | /example_service/prod/port   | 8080                 | String       | -             |
	// | /example_service/prod/user   | Ian                  | String       | -             |
	// | /example_service/prod/rate   | 0.5                  | String       | -             |
	// | /example_service/prod/secret | zOcZkAGB6aEjN7SAoVBT | SecureString | alias/aws/ssm |

	type Config struct {
		Debug  bool    `smm:"debug" default:"true"`
		Port   int     `smm:"port"`
		User   string  `smm:"user"`
		Rate   float32 `smm:"rate"`
		Secret string  `smm:"secret" required:"true"`
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	provider := &ssmconfig.Provider{
		SSM: ssm.New(sess),
	}

	var c Config
	err = provider.Process("/example_service/prod/", &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	format := "Debug: %v\nPort: %d\nUser: %s\nRate: %f\nSecret: %s\n"
	_, err = fmt.Printf(format, c.Debug, c.Port, c.User, c.Rate, c.Secret)
	if err != nil {
		log.Fatal(err.Error())
	}
}

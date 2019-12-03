# ssmconfig [![GoDoc](https://godoc.org/github.com/ianlopshire/go-ssm-config?status.svg)](http://godoc.org/github.com/ianlopshire/go-ssm-config) [![Report card](https://goreportcard.com/badge/github.com/ianlopshire/go-ssm-config)](https://goreportcard.com/report/github.com/ianlopshire/go-ssm-config) [![Go Cover](http://gocover.io/_badge/github.com/ianlopshire/go-ssm-config)](http://gocover.io/github.com/ianlopshire/go-ssm-config)

`import "github.com/ianlopshire/go-ssm-config"`

SSMConfig is a utility for loading configuration parameters from AWS SSM (Parameter Store) directly into a struct. This
package is largely inspired by [kelseyhightower/envconfig](https://github.com/kelseyhightower/envconfig).

## Motivation

This package was created to reduce the boilerplate code required when using Parameter Store to provide configuration to 
AWS Lambda functions. It should be suitable for additional applications.

## Usage

Set some parameters in [AWS Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html):

| Name                         | Value                | Type         | Key ID        |
| ---------------------------- | -------------------- | ------------ | ------------- |
| /exmaple_service/prod/debug  | false                | String       | -             |
| /exmaple_service/prod/port   | 8080                 | String       | -             |
| /exmaple_service/prod/user   | Ian                  | String       | -             |
| /exmaple_service/prod/rate   | 0.5                  | String       | -             |
| /exmaple_service/prod/secret | zOcZkAGB6aEjN7SAoVBT | SecureString | alias/aws/ssm |
        
Write some code:

```go
package main

import (
    "fmt"
    "log"
    "time"

    ssmconfig "github.com/ianlopshire/go-ssm-config"
)

type Config struct {
    Debug  bool    `smm:"debug" default:"true"`
    Port   int     `smm:"port"`
    User   string  `smm:"user"`
    Rate   float32 `smm:"rate"`
    Secret string  `smm:"secret" required:"true"`
}

func main() {
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
```

Result:

```
Debug: false
Port: 8080
User: Ian
Rate: 0.500000
Secret: zOcZkAGB6aEjN7SAoVBT
```

[Additional examples](https://godoc.org/github.com/ianlopshire/go-ssm-config#pkg-examples) can be found in godoc.

### Struct Tag Support

ssmconfig supports the use of struct tags to specify parameter name, default value, and required parameters.

```go
type Config struct {
    Param         string `ssm:"param"`
    RequiredParam string `ssm:"required_param" required:"true"`
    DefaultParam  string `ssm:"default_param" default:"foobar"`
}
```

The `ssm` tag is used to lookup the parameter in Parameter Store. It is joined to the base path passed into `Process()`.
If the `ssm` tag is missing ssmconfig will ignore the struct field.

The `default` tag is used to set the default value of a parameter. The default value will only be set if Parameter Store
returns the parameter as invalid.

The `required` tag is used to mark a parameter as required. If Parameter Store returns a required parameter as invalid,
ssmconfig will return an error.

The behavior of using the `default` and `required` tags on the same struct field is currently undefined.

### Supported Struct Field Types

ssmconfig supports these struct field types:

* string
* int, int8, int16, int32, int64
* bool
* float32, float64

More supported types may be added in the future.

## Licence

MIT
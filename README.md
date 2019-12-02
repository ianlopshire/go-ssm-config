# ssmconfig [![GoDoc](https://godoc.org/github.com/ianlopshire/go-ssm-config?status.svg)](http://godoc.org/github.com/ianlopshire/go-ssm-config) [![Report card](https://goreportcard.com/badge/github.com/ianlopshire/go-ssm-config)](https://goreportcard.com/report/github.com/ianlopshire/go-ssm-config) [![Go Cover](http://gocover.io/_badge/github.com/ianlopshire/go-ssm-config)](http://gocover.io/github.com/ianlopshire/go-ssm-config)

`import "github.com/ianlopshire/go-ssm-config"`

SSM Config is a utility for loading configuration values from AWS SSM (Parameter Store).

## Usage

SSM Config can be used to load configuration values directly into a struct. 

The path for each struct field is controlled by the `ssm` tag. If the `ssm` tag is omitted or empty it will be ignored. The field is set to the vale of the `default` tag if the path is invalid. If the `required` flag is set to `true` and the path is invalid, an error will be returned. 

```go
var config struct {
	Value1 string `ssm:"/path/to/value_1"`
	Value2 int    `ssm:"/path/to/value_2" default:"88"`
	Value3 string `ssm:"/path/to/value_3" required:"true"`
}

err := ssmconfig.Process("/base_path/", &config)
if err != nil {
	log.Fatal(err)
}
```

// Package ssmconfig is a utility for loading configuration values from AWS SSM (Parameter
// Store) directly into a struct.
package ssmconfig

import (
	"path"
	"reflect"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/pkg/errors"
)

// Process processes the config with a new default provider.
//
// See Provider.Process() for full documentation.
func Process(configPath string, c interface{}) error {
	sess, err := session.NewSession()
	if err != nil {
		err = errors.Wrap(err, "ssmconfig: could not create aws session")
		return err
	}

	p := Provider{SSM: ssm.New(sess)}
	return p.Process(configPath, c)
}

type Provider struct {
	SSM ssmiface.SSMAPI
}

// Process loads config values from smm (parameter store) into c. Encrypted parameters
// will automatically be decrypted. c must be a pointer to a struct.
//
// The `ssm` tag is used to lookup the parameter in Parameter Store. It is joined to the
// provided base path. If the `ssm` tag is missing the struct field will be ignored.
//
// The `default` tag is used to set the default value of a parameter. The default value
// will only be set if Parameter Store returns the parameter as invalid.
//
// The `required` tag is used to mark a parameter as required. If Parameter Store returns
// a required parameter as invalid an error will be returned.
//
// The behavior of using the `default` and `required` tags on the same struct field is
// currently undefined.
func (p *Provider) Process(configPath string, c interface{}) error {
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("ssmconfig: c must be a pointer to a struct")
	}
	v = reflect.Indirect(reflect.ValueOf(c))
	if v.Kind() != reflect.Struct {
		return errors.New("ssmconfig: c must be a pointer to a struct")
	}
	// get params required from struct
	spec := make(map[string]fieldSpec)
	buildStructSpec(configPath, v.Type(), spec)

	// get params from ssm parameter store
	params, invalidPrams, err := p.getParameters(spec)
	if err != nil {
		return errors.Wrap(err, "ssmconfig: could not get parameters")
	}

	// set values in struct
	return setValues(v, params, invalidPrams, spec)
}

func (p *Provider) getParameters(spec map[string]fieldSpec) (params map[string]string, invalidParams map[string]struct{}, err error) {
	// find all of the params that need to be requested
	var names []*string
	for key, val := range spec {
		if val.name == "" {
			continue
		}
		curr := spec[key]
		names = append(names, &curr.name)
	}

	input := &ssm.GetParametersInput{
		Names:          names,
		WithDecryption: aws.Bool(true),
	}

	output, err := p.SSM.GetParameters(input)
	if err != nil {
		return nil, nil, err
	}
	if output == nil {
		return nil, nil, nil
	}

	// convert the response to a map for easier use later
	params = map[string]string{}
	for i := range output.Parameters {
		params[*output.Parameters[i].Name] = *output.Parameters[i].Value
	}

	invalidParams = map[string]struct{}{}
	for i := range output.InvalidParameters {
		invalidParams[*output.InvalidParameters[i]] = struct{}{}
	}
	return params, invalidParams, nil
}

func setValues(v reflect.Value, params map[string]string, invalidParams map[string]struct{}, spec map[string]fieldSpec) error {
	if v.Kind() == reflect.Ptr {
		// Return on nil pointer
		if v.IsNil() {
			return nil
		}
		v = reflect.Indirect(v)
	}

	for i := 0; i < v.NumField(); i++ {
		// Add support for struct pointers
		ft := v.Type().Field(i).Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		if ft.Kind() == reflect.Struct {
			if err := setValues(v.Field(i), params, invalidParams, spec); err != nil {
				return err
			}
			continue
		}
		name := v.Type().Field(i).Tag.Get("ssm")
		if name == "" {
			continue
		}
		field := spec[name]
		if _, ok := invalidParams[field.name]; ok && field.required {
			return errors.Errorf("ssmconfig: %s is required", field.name)
		}

		value, ok := params[field.name]
		if !ok {
			value = field.defaultValue
		}
		if value == "" {
			continue
		}
		err := setBasicValue(v.Field(i), value)
		if err != nil {
			return errors.Wrapf(err, "ssmconfig: error setting field %s", v.Type().Field(i).Name)
		}
	}
	return nil
}

func setBasicValue(v reflect.Value, s string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(s)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.Atoi(s)
		if err != nil {
			return errors.Errorf("could not decode %q into type %v", s, v.Type().String())
		}
		v.SetInt(int64(i))

	case reflect.Float32:
		f, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return errors.Errorf("could not decode %q into type %v: %v", s, v.Type().String(), err)
		}
		v.SetFloat(f)

	case reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errors.Errorf("could not decode %q into type %v: %v", s, v.Type().String(), err)
		}
		v.SetFloat(f)

	case reflect.Bool:
		if s != "true" && s != "false" {
			return errors.Errorf("could not decode %q into type %v", s, v.Type().String())
		}
		v.SetBool(s == "true")

	default:
		return errors.Errorf("could not decode %q into type %v", s, v.Type().String())
	}

	return nil
}

type fieldSpec struct {
	name         string
	defaultValue string
	required     bool
}

func buildStructSpec(configPath string, t reflect.Type, spec map[string]fieldSpec) {
	for i := 0; i < t.NumField(); i++ {
		// Add support for struct pointers
		ft := t.Field(i).Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		if ft.Kind() == reflect.Struct {
			buildStructSpec(configPath, ft, spec)
			continue
		}
		name := t.Field(i).Tag.Get("ssm")
		if name == "" {
			continue
		}
		spec[name] = fieldSpec{
			name:         path.Join(configPath, name),
			defaultValue: t.Field(i).Tag.Get("default"),
			required:     t.Field(i).Tag.Get("required") == "true",
		}
	}
}

package ssmconfig_test

import (
	"encoding/hex"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	ssmconfig "github.com/ianlopshire/go-ssm-config"
)

type mockSSMClient struct {
	ssmiface.SSMAPI
	calledWithInput *ssm.GetParametersInput
	output          *ssm.GetParametersOutput
	err             error
}

type Hex struct {
	V string
}

func (h *Hex) UnmarshalText(val []byte) error {
	dst := make([]byte, hex.DecodedLen(len(val)))
	n, err := hex.Decode(dst, val)
	if err != nil {
		return err
	}
	h.V = string(dst[:n])
	return nil
}

func (c *mockSSMClient) GetParameters(input *ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	c.calledWithInput = input
	if c.err != nil {
		return nil, c.err
	}
	return c.output, nil
}

func TestProvider_Process(t *testing.T) {
	t.Run("base case", func(t *testing.T) {
		var s struct {
			S1      string  `ssm:"/strings/s1"`
			S2      string  `ssm:"/strings/s2" default:"string2"`
			I1      int     `ssm:"/int/i1"`
			I2      int     `ssm:"/int/i2" default:"42"`
			B1      bool    `ssm:"/bool/b1"`
			B2      bool    `ssm:"/bool/b2" default:"false"`
			F321    float32 `ssm:"/float32/f321"`
			F322    float32 `ssm:"/float32/f322" default:"42.42"`
			F641    float64 `ssm:"/float64/f641"`
			F642    float64 `ssm:"/float64/f642" default:"42.42"`
			H1      *Hex    `ssm:"/hex/hex1"`
			H2      *Hex    `ssm:"/hex/hex2" default:"737472696e6731"`
			Invalid string
		}

		mc := &mockSSMClient{
			output: &ssm.GetParametersOutput{
				Parameters: []*ssm.Parameter{
					{
						Name:  aws.String("/base/strings/s1"),
						Value: aws.String("string1"),
					},
					{
						Name:  aws.String("/base/int/i1"),
						Value: aws.String("42"),
					},
					{
						Name:  aws.String("/base/bool/b1"),
						Value: aws.String("true"),
					},
					{
						Name:  aws.String("/base/float32/f321"),
						Value: aws.String("42.42"),
					},
					{
						Name:  aws.String("/base/float64/f641"),
						Value: aws.String("42.42"),
					},
					{
						Name:  aws.String("/base/hex/hex1"),
						Value: aws.String("737472696e6731"),
					},
				},
			},
		}

		p := &ssmconfig.Provider{
			SSM: mc,
		}

		err := p.Process("/base/", &s)

		if err != nil {
			t.Errorf("Process() unexpected error: %q", err.Error())
		}

		names := make([]string, len(mc.calledWithInput.Names))
		for i := range mc.calledWithInput.Names {
			names[i] = *mc.calledWithInput.Names[i]
		}
		expectedNames := []string{
			"/base/strings/s1",
			"/base/strings/s2",
			"/base/int/i1",
			"/base/int/i2",
			"/base/bool/b1",
			"/base/bool/b2",
			"/base/float32/f321",
			"/base/float32/f322",
			"/base/float64/f641",
			"/base/float64/f642",
			"/base/hex/hex1",
			"/base/hex/hex2",
		}

		if !reflect.DeepEqual(names, expectedNames) {
			t.Errorf("Process() unexpected input names: have %v, want %v", names, expectedNames)
		}

		if s.S1 != "string1" {
			t.Errorf("Process() S1 unexpected value: want %q, have %q", "string1", s.S1)
		}
		if s.S2 != "string2" {
			t.Errorf("Process() S2 unexpected value: want %q, have %q", "string2", s.S2)
		}
		if s.I1 != 42 {
			t.Errorf("Process() I1 unexpected value: want %d, have %d", 101, s.I1)
		}
		if s.I2 != 42 {
			t.Errorf("Process() I2 unexpected value: want %d, have %d", 5, s.I2)
		}
		if s.B1 != true {
			t.Errorf("Process() B1 unexpected value: want %v, have %v", true, s.B1)
		}
		if s.B2 != false {
			t.Errorf("Process() B2 unexpected value: want %v, have %v", false, s.B1)
		}
		if s.F321 != 42.42 {
			t.Errorf("Process() F321 unexpected value: want %f, have %f", 42.42, s.F321)
		}
		if s.F322 != 42.42 {
			t.Errorf("Process() F322 unexpected value: want %f, have %f", 42.42, s.F322)
		}
		if s.F641 != 42.42 {
			t.Errorf("Process() F641 unexpected value: want %f, have %f", 42.42, s.F641)
		}
		if s.F642 != 42.42 {
			t.Errorf("Process() F642 unexpected value: want %f, have %f", 42.42, s.F642)
		}
		if s.H1.V != "string1" {
			t.Errorf("Process() H1 unexpected value: want %s, have %s", "string1", s.H1.V)
		}
		if s.H2.V != "string1" {
			t.Errorf("Process() H2 unexpected value: want %s, have %s", "string1", s.H2.V)
		}
		if s.Invalid != "" {
			t.Errorf("Process() Missing unexpected value: want %q, have %q", "", s.Invalid)
		}
	})

	for _, tt := range []struct {
		name       string
		configPath string
		c          interface{}
		want       interface{}
		client     ssmiface.SSMAPI
		shouldErr  bool
	}{
		{
			name:       "invalid int",
			configPath: "/base/",
			c: &struct {
				I1 int `ssm:"/int/i1" default:"notAnInt"`
			}{},
			client:    &mockSSMClient{},
			shouldErr: true,
		},
		{
			name:       "invalid float32",
			configPath: "/base/",
			c: &struct {
				F32 float32 `ssm:"/float32/f32" default:"notAFloat"`
			}{},
			client:    &mockSSMClient{},
			shouldErr: true,
		},
		{
			name:       "invalid float64",
			configPath: "/base/",
			c: &struct {
				F32 float64 `ssm:"/float64/f64" default:"notAFloat"`
			}{},
			client:    &mockSSMClient{},
			shouldErr: true,
		},
		{
			name:       "invalid bool",
			configPath: "/base/",
			c: &struct {
				B1 bool `ssm:"/bool/b1" default:"notABool"`
			}{},
			client:    &mockSSMClient{},
			shouldErr: true,
		},
		{
			name:       "missing required parameter",
			configPath: "/base/",
			c: &struct {
				S1 string `ssm:"/strings/s1" required:"true"`
			}{},
			client: &mockSSMClient{
				output: &ssm.GetParametersOutput{
					InvalidParameters: []*string{aws.String("/base/strings/s1")},
				},
			},
			shouldErr: true,
		},
		{
			name:       "unsupported field type",
			configPath: "/base/",
			c: &struct {
				M1 map[string]string `ssm:"/map/m1"`
			}{},
			client: &mockSSMClient{
				output: &ssm.GetParametersOutput{
					Parameters: []*ssm.Parameter{{
						Name:  aws.String("/base/map/m1"),
						Value: aws.String("notSupported"),
					}},
				},
			},
			shouldErr: true,
		},
		{
			name:       "blank value from ssm",
			configPath: "/base/",
			c: &struct {
				S1 string `ssm:"/strings/s1"`
			}{},
			want: &struct {
				S1 string `ssm:"/strings/s1"`
			}{},
			client: &mockSSMClient{
				output: &ssm.GetParametersOutput{
					Parameters: []*ssm.Parameter{{
						Name:  aws.String("/strings/s1"),
						Value: aws.String(""),
					}},
				},
			},
			shouldErr: false,
		},
		{
			name:       "input config not a pointer",
			configPath: "/base/",
			c: struct {
				S1 string `ssm:"/strings/s1"`
			}{},
			client:    &mockSSMClient{},
			shouldErr: true,
		},
		{
			name:       "input config not a struct",
			configPath: "/base/",
			c: &[]struct {
				S1 string `ssm:"/strings/s1"`
			}{},
			client:    &mockSSMClient{},
			shouldErr: true,
		},
		{
			name:       "ssm client error",
			configPath: "/base/",
			c: &struct {
				S1 string `ssm:"/strings/s1" required:"true"`
			}{},
			client: &mockSSMClient{
				err: errors.New("ssm client error"),
			},
			shouldErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := &ssmconfig.Provider{
				SSM: tt.client,
			}

			err := p.Process(tt.configPath, tt.c)

			if tt.shouldErr && err == nil {
				t.Errorf("Process() Expected error but have nil")
			}

			if !tt.shouldErr && err != nil {
				t.Errorf("Process() unexpected error: %v", err)
			}

			if !tt.shouldErr && !reflect.DeepEqual(tt.c, tt.want) {
				t.Errorf("Process() want %+v, have %+v", tt.want, tt.c)
			}

		})
	}
}

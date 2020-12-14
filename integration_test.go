package ssmconfig_test

import (
	"flag"
	"os"
	"testing"

	ssmconfig "github.com/ianlopshire/go-ssm-config"
)

var IsIntegrationRun bool

func init() {
	flag.BoolVar(&IsIntegrationRun, "integration", false, "run integration tests")
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestProcess_integration(t *testing.T) {
	if !IsIntegrationRun {
		t.Skip("Use the -integration flag to run integration tests")
		return
	}

	t.Run("base case", func(t *testing.T) {
		var s struct {
			S1   string  `ssm:"/strings/s1"`
			S2   string  `ssm:"/base/strings/s2"`
			I1   int     `ssm:"/int/i1"`
			B1   bool    `ssm:"/bool/b1"`
			F321 float32 `ssm:"/float32/f321"`
			F641 float64 `ssm:"/float64/f641"`
		}

		err := ssmconfig.Process("/go-ssm-config", &s)
		if err != nil {
			t.Errorf("Process() unexpected error: %q", err.Error())
		}

		if s.S1 != "string1" {
			t.Errorf("Process() S1 unexpected value: want %q, have %q", "string1", s.S1)
		}

		if s.S2 != "string2" {
			t.Errorf("Process() S2 unexpected value: want %q, have %q", "string2", s.S2)
		}

		if s.I1 != 42 {
			t.Errorf("Process() I1 unexpected value: want %d, have %d", 42, s.I1)
		}

		if s.B1 != true {
			t.Errorf("Process() B1 unexpected value: want %v, have %v", true, s.B1)
		}

		if s.F321 != 42.42 {
			t.Errorf("Process() F321 unexpected value: want %f, have %f", 42.42, s.F321)
		}

		if s.F641 != 42.42 {
			t.Errorf("Process() F641 unexpected value: want %f, have %f", 42.42, s.F641)
		}

	})

	t.Run("use default when value is not present", func(t *testing.T) {
		var s struct {
			S3 string `ssm:"/strings/s3" default:"default"`
		}

		err := ssmconfig.Process("/go-ssm-config", &s)
		if err != nil {
			t.Errorf("Process() unexpected error: %q", err.Error())
		}

		if s.S3 != "default" {
			t.Errorf("Process() S3 unexpected value: want %q, have %q", "default", s.S3)
		}
	})

	t.Run("do not use default when value is present", func(t *testing.T) {
		var s struct {
			S1     string `ssm:"/strings/s1" default:"default"`
			IZero  int    `ssm:"/int/i_zero" default:"42"`
			BFalse bool   `ssm:"/bool/b2" default:"true"`
		}

		err := ssmconfig.Process("/go-ssm-config", &s)
		if err != nil {
			t.Errorf("Process() unexpected error: %q", err.Error())
		}

		if s.S1 != "string1" {
			t.Errorf("Process() S1 unexpected value: want %q, have %q", "string1", s.S1)
		}

		if s.IZero != 0 {
			t.Errorf("Process() IZero unexpected value: want %d, have %d", 0, s.IZero)
		}

		if s.BFalse != false {
			t.Errorf("Process() BFalse unexpected value: want %v, have %v", false, s.BFalse)
		}
	})

	t.Run("error on missing required value", func(t *testing.T) {
		var s struct {
			S3 string `ssm:"/strings/s3" required:"true"`
		}

		err := ssmconfig.Process("/go-ssm-config", &s)
		if err == nil {
			t.Error("Process() expexted error but got nil")
		}
	})

	t.Run("dont error on zero value required value", func(t *testing.T) {
		var s struct {
			IZero int `ssm:"/int/i_zero" require:"true"`
		}

		err := ssmconfig.Process("/go-ssm-config", &s)
		if err != nil {
			t.Errorf("Process() unexpected error: %q", err.Error())
		}
	})
}

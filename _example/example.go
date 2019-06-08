package main

import (
	"fmt"

	"github.com/arteev/go-checks"
)

type Config struct {
	Enabled  bool
	Listen   string `check:"required"`
	LogLevel string `check:"required,expect:info;debug;error;"`
	Timeout  int    `check:"deprecated"`

	ValueForFunc string `check:"call:ValueCheck"`
}

func (c Config) Check() error {
	if !c.Enabled {
		return checks.ErrSkip
	}
	//special check
	return nil
}

func (c Config) ValueCheck(name string, s string) error {
	// special check for specific fields
	if name != "ValueForFunc" || s == "valid" {
		return nil
	}
	return fmt.Errorf("Not valid value: %q", s)
}

func main() {
	//all errors
	v := &Config{
		Enabled:  true,
		LogLevel: "warn",
		Timeout:  10,
	}
	errs := checks.CheckAll(v)
	if len(errs) != 0 {
		fmt.Println("Errors:", errs)
	}

	//first error
	err := checks.Check(v)
	if err != nil {
		fmt.Println("Error:", err)
	}

	//skip checks
	v.Enabled = false
	err = checks.Check(v)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Skip checks")
	}

	//with deprecated
	v.Enabled = true
	checker := checks.New(checks.ModeAll, checks.WarningType)
	errs = checker.Check(v)
	if len(errs) != 0 {
		fmt.Println("Errors:", errs)
	}

	//normal
	v = &Config{
		Enabled:      true,
		Listen:       ":8080",
		LogLevel:     "debug",
		ValueForFunc: "valid",
	}
	errs = checks.CheckAll(v)
	if len(errs) == 0 {
		fmt.Println("OK")
	} else {
		fmt.Println("Errors:", errs)
	}
}

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
}

func (c Config) Check() error {
	if !c.Enabled {
		return checks.ErrSkip
	}
	//specific check
	return nil
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
		Enabled:  true,
		Listen:   ":8080",
		LogLevel: "debug",
	}
	errs = checks.CheckAll(v)
	if len(errs) == 0 {
		fmt.Println("OK")
	}
}

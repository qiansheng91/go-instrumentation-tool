package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

type configuration struct {
	Plugins []plugin `yaml:"plugins"`
}

type plugin struct {
	Path            string          `yaml:"path"`
	Name            string          `yaml:"name"`
	Plugin_package  string          `yaml:"package"`
	Target_packages []targetPackage `yaml:"target_packages"`
}

type targetPackage struct {
	PackagePath string     `yaml:"packagePath"`
	Name        string     `yaml:"name"`
	PointCuts   []pointCut `yaml:"point_cuts"`
}

type pointCut struct {
	TargetSignature string `yaml:"target_signature"`
	BeforeAdvice    string `yaml:"before_advice"`
	AfterAdvice     string `yaml:"after_advice"`
}

func Parser(filePath string) (InstrumentationInfo, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
		return nil, err
	}

	config := &configuration{}
	if err = yaml.Unmarshal(data, config); err != nil {
		log.Fatalf("failed to unmarshal: %v", err)
		return nil, err
	}
	return newInstrumentationInfo(config), nil
}

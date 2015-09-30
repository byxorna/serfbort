package main

type Config struct {
	Targets map[string]TargetConfig `json:"targets"`
}

// TargetConfig is a struct containing information about how to deploy given application
// (i.e. what script to run to deploy and verify it)
type TargetConfig struct {
	DeployScript string `json:"deploy"`    // the text/template script to execute for deploy
	VerifyScript string `json:"verify"`    // the text/template script to execute for verify
	Directory    string `json:"directory"` // what directory to cd into before running deploy and verify
}

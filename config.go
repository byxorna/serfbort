package main

type Config struct {
	Targets map[string]Target `json:"deploy"`
}

type Target struct {
	ActionTemplate string `json:"action"`
	VerifyTemplate string `json:"verify"`
}

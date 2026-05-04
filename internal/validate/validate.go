package validate

import "os"

type Finding struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path,omitempty"`
}
type Result struct {
	Status   string    `json:"status"`
	Findings []Finding `json:"findings"`
}

func Validate(path string) (Result, error) {
	res := Result{Status: "ok", Findings: []Finding{}}
	if _, err := os.Stat(path + "/cosmos.yaml"); err != nil {
		res.Status = "failed"
		res.Findings = append(res.Findings, Finding{Code: "COSMOS_MISSING", Severity: "error", Message: "cosmos.yaml fehlt", Path: "cosmos.yaml"})
	}
	return res, nil
}

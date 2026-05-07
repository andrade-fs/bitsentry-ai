package system

import "os/exec"

type DependencyStatus struct {
	Name      string
	Found     bool
	Path      string
	Mandatory bool
}

func CheckDependencies() []DependencyStatus {
	deps := []struct {
		name      string
		mandatory bool
	}{
		{name: "go", mandatory: true},
		{name: "git", mandatory: true},
		{name: "brew", mandatory: false},
		{name: "apt", mandatory: false},
		{name: "yum", mandatory: false},
		{name: "pacman", mandatory: false},
	}

	results := make([]DependencyStatus, 0, len(deps))
	for _, dep := range deps {
		path, err := exec.LookPath(dep.name)
		results = append(results, DependencyStatus{
			Name:      dep.name,
			Found:     err == nil,
			Path:      path,
			Mandatory: dep.mandatory,
		})
	}

	return results
}

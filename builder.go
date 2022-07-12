package tykky

import "path/filepath"

type CondaInstallTemplate struct {
	Conda              CondaInstallation
	PipRequirementFile string
	CondaEnvFile       string
	UseMamba           bool
}

type venvInstallTemplate struct {
	pipRequirementFile string
	venvName           string
}

func (v venvInstallTemplate) ConstructRunScript() []string {
	var script []string
	script = append(script, ShPipVenvCreate(v.venvName))
	if v.pipRequirementFile != "" {
		script = append(script, ShPipInstall(v.pipRequirementFile, ""))
	}
	return script

}

type InstallTemplate interface {
	ConstructRunScript() []string
}

func (c CondaInstallTemplate) envIsExplicit() bool {
	fileExtension := filepath.Ext(c.CondaEnvFile)
	if fileExtension == "yml" || fileExtension == "yaml" {
		return false
	} else {
		return true
	}
}

func (c CondaInstallTemplate) ConstructRunScript(UserPreActionFile string, UserPostActionFile string) []string {
	var script []string
	script = append(script, c.Conda.ShCondaPreActivate())
	script = append(script, "export envroot="+c.Conda.getEnvRoot())
	script = append(script, "export condaroot="+c.Conda.InstallRoot)
	if UserPreActionFile != "" {
		script = append(script, UserPreActionFile)
	}

	if c.UseMamba {
		script = append(script, c.Conda.ShInstallMamba())
		script = append(script, c.Conda.ShEnvMamba(c.CondaEnvFile, c.envIsExplicit()))
	} else {
		script = append(script, c.Conda.ShEnvConda(c.CondaEnvFile, c.envIsExplicit()))
	}
	script = append(script, c.Conda.ShCondaActivate())
	if c.PipRequirementFile != "" {
		script = append(script, ShPipInstall(c.PipRequirementFile, ""))
	}
	if UserPostActionFile != "" {
		script = append(script, ShSource(UserPostActionFile))
	}
	script = append(script, c.Conda.ShCondaList())
	return script
}

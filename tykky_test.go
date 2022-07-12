package tykky

import (
	"strings"
	"testing"
)

func TestCondaConstructRunScript(t *testing.T) {
	cInst := CondaInstallation{
		CmdPath:     "/TEST/miniconda/bin/conda",
		InstallRoot: "/TEST/miniconda",
		EnvName:     "env1",
	}
	T := CondaInstallTemplate{
		Conda:              cInst,
		PipRequirementFile: "",
		CondaEnvFile:       "env.txt",
		UseMamba:           false,
	}

	refStr := `eval "$(/TEST/miniconda/bin/conda shell.bash hook)"
conda env create --name env1 --file env.txt 
conda activate env1
if [[ -d /TEST/miniconda/envs/env1/bin ]]; then
	echo "echo \"$(conda list)\"" > /TEST/miniconda/envs/env1/bin/list-packages
	chmod +x /TEST/miniconda/envs/env1/bin/list-packages 
fi`
	refStr = strings.Trim(refStr, " \n\t")
	res := strings.Join(T.ConstructRunScript("", ""), "\n")
	res = strings.Trim(res, " \n\t")
	if res != refStr {
		t.Fatalf("ConstructRunScript() GOT\n%s\nWANTED\n%s", res, refStr)
	}
}

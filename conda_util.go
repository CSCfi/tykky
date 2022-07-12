package tykky

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

type CondaInstallerConfig struct {
	DownloadDir     string
	InstallationDir string
	BaseUrl         string
	Version         string
	Arch            string
}

var LogDir = ""

const CondaDefaultBaseUrl = "https://repo.anaconda.com/miniconda/Miniconda3"

func (c CondaInstallerConfig) constructCondaUrl() string {
	return fmt.Sprintf("%s-%s-%s.sh", c.BaseUrl, c.Version, c.Arch)
}

type CondaInstallation struct {
	CmdPath     string
	InstallRoot string
	EnvName     string
}

func PrintCondaLiscense(pathToCondaInstaller string) {
	b := new(bytes.Buffer)
	commandOutput := outputWriter{nil, b, -1, fmt.Sprintf("%s/%s", LogDir, "conda.license")}
	RunPseudoInteractiveCommand(fmt.Sprintf("%s", pathToCondaInstaller), commandOutput, "\nno")
	a := strings.Split(b.String(), "\n")
	fmt.Printf("%s", strings.Join(a[:len(a)-3], "\n"))
}

func CondaLiscenseIsAccepted(pathToCondaInstaller string, askUser bool) bool {

	if !askUser {
		return true
	}
	if askForConfirmation("Show miniconda liscense?") {
		PrintCondaLiscense(pathToCondaInstaller)
	}
	if askForConfirmation("I have read and accepted the miniconda liscense") {
		return true
	} else {
		return false
	}
}
func DownloadCondaInstaller(condaInstallerUrl string, downloadPath string) error {
	downloadErr := DownloadFile(condaInstallerUrl, downloadPath, false)
	if downloadErr != nil {
		fmt.Printf(downloadErr.Error())
		return downloadErr
	} else {
		os.Chmod(downloadPath, 0700)
		return nil
	}
}
func InstallConda(installerPath string, targetPrefix string) {
	commandOutput := outputWriter{nil, os.Stdout, 20, fmt.Sprintf("%s/%s", LogDir, "conda_install.log")}
	RunCommand(fmt.Sprintf("%s -b -p %s", installerPath, targetPrefix), commandOutput)
}
func (c CondaInstallerConfig) SetupConda() error {
	tempDownloadpath := fmt.Sprintf("%s/%s", c.DownloadDir, "conda_installer.sh")
	DownloadCondaInstaller(c.constructCondaUrl(), tempDownloadpath)
	if CondaLiscenseIsAccepted(tempDownloadpath, true) {
		InstallConda(tempDownloadpath, c.InstallationDir)
		return nil
	} else {
		return errors.New("License was not accepted")
	}
}

func (c CondaInstallation) ShCondaPreActivate() string {
	return fmt.Sprintf("eval \"$(%s shell.bash hook)\"", c.CmdPath)
}

func (c CondaInstallation) getEnvRoot() string {
	return path.Clean(c.InstallRoot + "/envs/" + c.EnvName)
}

// Assumes that environment has been activated
func (c CondaInstallation) ShInstallEnv(command string, envFile string, envIsExplicit bool) string {
	fileFlag := ""
	subCommand := ""
	if envIsExplicit {
		fileFlag = "--file"
		subCommand = "env"
	} else {
		fileFlag = "-f"
	}
	return fmt.Sprintf("%s %s create --name %s %s %s ", command, subCommand, c.EnvName, fileFlag, envFile)
}
func (c CondaInstallation) ShEnvConda(envFile string, envIsExplicit bool) string {
	return c.ShInstallEnv("conda", envFile, envIsExplicit)
}
func (c CondaInstallation) ShEnvMamba(envFile string, enviIsExplicit bool) string {
	return c.ShInstallEnv("mamba", envFile, enviIsExplicit)
}
func (c CondaInstallation) ShCondaActivate() string {
	return fmt.Sprintf("conda activate %s", c.EnvName)
}

func (c CondaInstallation) ShInstallMamba() string {
	return "conda install -y mamba -n base -c conda-forge"
}

func ShSource(path string) string {
	return fmt.Sprintf("source %s", path)
}

func ShPipInstall(reqFile string, installArgs string) string {
	return fmt.Sprintf("pip install %s -r%s", installArgs, reqFile)
}
func ShPipVenvCreate(envName string) string {
	return fmt.Sprintf("python -m venv %s", envName)
}

func (c CondaInstallation) ShCondaList() string {
	s := `if [[ -d %s ]]; then
	echo "echo \"$(conda list)\"" > %s
	chmod +x %s 
fi
	`
	envr := c.getEnvRoot()
	listPkgCmd := envr + "/bin/list-packages"
	return fmt.Sprintf(s, envr+"/bin", listPkgCmd, listPkgCmd)
}

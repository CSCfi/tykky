package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"tykky"
)

func main() {
	DownloadDir := tykky.CreateBuildDir("/tmp")
	instd := DownloadDir + "/_instdir"
	os.Mkdir(DownloadDir+"/_deploy", os.ModePerm)
	os.Mkdir(DownloadDir+"/_instdir", os.ModePerm)
	tykky.PrintInfo("Fetching container")
	tykky.GetApptainerContainer("library://centos:7.9", DownloadDir+"/_deploy/container.sif", false)

	containerInstd := "/TYKKY_CONTAINER"
	c := tykky.CondaInstallerConfig{
		DownloadDir:     DownloadDir,
		InstallationDir: containerInstd + "/miniconda",
		BaseUrl:         tykky.CondaDefaultBaseUrl,
		Version:         "latest",
		Arch:            "Linux-x86_64",
	}
	file, _ := json.MarshalIndent(c, "", " ")
	_ = ioutil.WriteFile("conda_config.json", file, 0644)
	cont := tykky.ApptainerContainerInstance{ContainerPath: DownloadDir + "/_deploy/container.sif",
		BindMounts: []string{instd + ":" + containerInstd, DownloadDir, os.Getenv("PWD")},
	}
	progPath, _ := os.Executable()
	tykky.PrintInfo("Installing conda")
	cont.RunInContainer(fmt.Sprintf("%s/condaInstall -log-dir %s -config conda_config.json", filepath.Dir(progPath), DownloadDir), false)
	cInst := tykky.CondaInstallation{
		CmdPath:     c.InstallationDir + "/bin/conda",
		InstallRoot: containerInstd + "/miniconda",
		EnvName:     "env1",
	}

	T := tykky.CondaInstallTemplate{
		Conda:              cInst,
		PipRequirementFile: "",
		CondaEnvFile:       "env.yaml",
		UseMamba:           true,
	}
	installScript := T.ConstructRunScript("", "")
	installScript = append([]string{"export installroot=" + containerInstd}, installScript...)
	tykky.WriteShellScript(DownloadDir+"/script.sh", installScript)
	tykky.PrintInfo("Installing conda environment")
	cont.RunInContainer(DownloadDir+"/script.sh", true)
	tykky.PrintInfo("Creating squashfs image")
	tykky.RunPlainOutput(fmt.Sprintf("mksquashfs %s %s/img.sqfs -processors 4 -noappend", instd, DownloadDir+"/_deploy"))
	//cont.RunInContainer("./min.sh", true)

	/*
		installErr := condaSrc{condaDefaultBaseUrl, "latest", "Linux-x86_64"}.SetupConda(config.instDir)
		if installErr != nil {
			fmt.Printf("%s", installErr.Error())
			os.Exit(1)
		}

		condaCmdPath := fmt.Sprintf("%s/%s", config.instDir, "bin/conda")
		conda := condaInstallation{condaCmdPath, config.instDir, "env1"}
		fmt.Printf("%s", conda.cmdPath)
	*/
	//output := outputWriter{nil, 10, config.wrkDir + "/prog.out"}
	//RunCommand("/home/hnortamo/Projects/ContConda/Tykky_go/tykky/prog", output)
	//syscall.Exec("/usr/bin/python3", append([]string{"python3"}, os.Args[1:]...), os.Environ())
	//w := createWrapper()
	//fmt.Printf("location: %s name: %s\n", w.location, w.name)

	//installationContainer := apptainerContainerInstance{
	//	config.builDir + "/_deploy/container.sif",
	//	[]string{config.builDir + "/_deploy/:/CSC_CONTAINER"}}
	//cont := apptainerContainerInstance{config.builDir + "/_deploy/container.sif", nil}
}

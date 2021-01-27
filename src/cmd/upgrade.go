package cmd

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/khanhas/spicetify-cli/src/utils"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func Upgrade(currentVersion string) {
	utils.PrintBold("Fetch latest release info:")
	tagName, err := FetchLatestTag()
	if err != nil {
		utils.PrintError("Cannot fetch latest release info")
		utils.PrintError(err.Error())
		return
	}
	utils.PrintGreen("OK")

	utils.PrintInfo("Current version: " + currentVersion)
	utils.PrintInfo("Latest release: " + tagName)
	if currentVersion == tagName {
		utils.PrintSuccess("Already up-to-date.")
		return
	}

	var assetURL string = "https://github.com/khanhas/spicetify-cli/releases/download/v" + tagName + "/spicetify-" + tagName
	var location string
	switch runtime.GOOS {
	case "windows":
		assetURL += "-windows-x64.zip"
		location = os.Getenv("TEMP") + "spicetify-" + tagName + ".zip"
	case "linux":
		assetURL += "-linux-amd64.tar.gz"
		location = "/tmp/spicetify-" + tagName + ".tar.gz"
	case "darwin":
		assetURL += "-darwin-amd64.tar.gz"
		location = os.Getenv("TMPDIR") + "spicetify-" + tagName + ".tar.gz"
	}

	utils.PrintBold("Downloading:")

	out, err := os.Create(location)
	if err != nil {
		utils.Fatal(err)
	}
	defer out.Close()

	resp2, err := http.Get(assetURL)
	if err != nil {
		utils.Fatal(err)
	}

	_, err = io.Copy(out, resp2.Body)
	if err != nil {
		utils.Fatal(err)
	}
	utils.PrintGreen("OK")

	exe, err := os.Executable()
	if err != nil {
		utils.Fatal(err)
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		utils.Fatal(err)
	}

	exeOld := exe + ".old"
	utils.CheckExistAndDelete(exeOld)

	if err = os.Rename(exe, exeOld); err != nil {
		permissionError(err)
	}

	utils.PrintBold("Extracting:")
	switch runtime.GOOS {
	case "windows":
		err = utils.Unzip(location, utils.GetExecutableDir())

	case "linux", "darwin":
		err = exec.Command("tar", "-xzf", location, "-C", utils.GetExecutableDir()).Run()
	}
	if err != nil {
		permissionError(err)
	}

	utils.CheckExistAndDelete(exeOld)
	utils.PrintGreen("OK")
	utils.PrintSuccess("spicetify is up-to-date.")
}

func permissionError(err error) {
	utils.PrintInfo("If fatal error is \"Permission denied\", please check read/write permission of spicetify executable directory.")
	utils.PrintInfo("However, if you used a package manager to install spicetify, please upgrade by using the same package manager.")
	utils.Fatal(err)
}

func FetchLatestTag() (string, error) {
	res, err := http.Get("https://api.github.com/repos/khanhas/spicetify-cli/releases/latest")
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var release githubRelease
	if err = json.Unmarshal(body, &release); err != nil {
		return "", err
	}

	return release.TagName[1:], nil
}

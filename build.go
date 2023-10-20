package main

import (
	"fmt"
	"goutils"
	"log"
	_ "math"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func main() {

	var build_config_path string
	var env_type string
	var err error
	//var ok bool
	var database_server_name string
	var database_name string
	var database_username string
	var database_password string
	var buildConfigs map[string]interface{}
	var appConfigs map[string]interface{}
	var gitFolder string
	var gitBranch string
	var gitCommitHash string
	var app_name string
	var version string
	var builddate string

	now := time.Now()
	builddate = now.Format("2006-01-02 15:04") //YYYY-MM-dd HH mm

	builddir := "build"
	buildappdir := "build/app"

	// build config file path
	build_config_path = getBuildConfigPath()

	if build_config_path == "" {
		panic("build_config_path not found")
	}

	fmt.Println("build_config_path=", build_config_path)

	//get the directory name of the git folder
	gitFolder = getGitFolder()
	gitBranch = getGitBranch()
	gitCommitHash = getGitCommit(gitBranch)
	fmt.Println("git=", gitFolder, gitBranch, gitCommitHash)

	if gitBranch == "master" || gitBranch == "main" {
		env_type = "prod"
	} else {
		env_type = "test"
	}

	fmt.Println("env_type=", env_type)

	appconfigpath := getAppConfigPath(env_type)
	appConfigs, err = goutils.ParseJSONFromFile(appconfigpath)
	fmt.Println("appConfigs=", appConfigs)

	//parse appconfigs
	app_name = goutils.ToString(appConfigs["app_name"])
	version = goutils.ToString(appConfigs["version"])
	fmt.Println("app_name=", app_name, "version=", version)

	buildConfigs = getBuildConfig(gitFolder, build_config_path)
	fmt.Println("buildconfigs=", buildConfigs)

	if buildConfigs["app_name"] == nil {
		panic("app config not found")
	}

	database_server_name = goutils.ToString(buildConfigs["database_server_name"])
	database_name = goutils.ToString(buildConfigs["database_server_name"])
	database_username = goutils.ToString(buildConfigs["database_username"])
	database_password = goutils.ToString(buildConfigs["database_password"])

	if database_server_name == "" {
		panic("database_server_name is empty")
	}

	if database_name == "" {
		panic("database_name is empty")
	}

	if database_username == "" {
		panic("database_username is empty")
	}

	var dockerFileContents string
	dockerFileContents, err = goutils.ReadFile("Dockerfile")
	if dockerFileContents != "" {

		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{app_name}", app_name)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{version}", version)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{builddate}", builddate)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{commithash}", gitCommitHash)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{database_name}", database_name)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{database_password}", database_password)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{database_server_name}", database_server_name)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{database_username}", database_username)

		//write new file
		err = goutils.WriteFile(builddir+"/Dockerfile", dockerFileContents, true)
		writeError(err)
	}

	if env_type == "test" || env_type == "" {
		err = goutils.CopyFile1("appconfig.test.json", buildappdir+"/appconfig.json")
		writeError(err)
		if err == nil {
			fmt.Println("copied appconfig.test.json " + buildappdir + "/appconfig.json")
		}
	}

	if env_type == "prod" {
		err = goutils.CopyFile1("appconfig.prod.json", buildappdir+"/appconfig.json")
		writeError(err)
		if err == nil {
			fmt.Println("copied appconfig.prod.json " + buildappdir + "/appconfig.json")
		}
	}

	//update the appconfig branch name and the committhash
	var content = ""
	content, err = goutils.ReadFile(buildappdir + "/appconfig.json")
	content = strings.ReplaceAll(content, "{commithash}", gitCommitHash)
	content = strings.ReplaceAll(content, "{builddate}", builddate)

	err = goutils.WriteFile(buildappdir+"/appconfig.json", content, true)
	writeError(err)

	//copy pm2
	err = goutils.CopyFile1("pm2.yml", buildappdir+"/pm2.yml")
	writeError(err)

	fmt.Println("build complete")
	os.Exit(0)
}

func getGitFolder() string {
	buffer, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		log.Fatal("failed getGitFolder ", err)
	}
	var rv = strings.ReplaceAll(string(buffer), "\r", "")
	rv = strings.ReplaceAll(string(rv), "\n", "")
	rv = strings.ReplaceAll(string(rv), " ", "")
	rv = goutils.GetFileName(rv)
	return rv
}

func getGitBranch() string {

	buffer, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		log.Fatal("failed getGitBranch ", err)
	}

	var rv = strings.ReplaceAll(string(buffer), "\r", "")
	rv = strings.ReplaceAll(string(rv), "\n", "")
	rv = strings.ReplaceAll(string(rv), " ", "")
	return rv
}

func getGitCommit(branch string) string {

	buffer, err := exec.Command("git", "rev-parse", branch).Output()

	if err != nil {
		log.Fatal("failed getGitCommit ", err)
	}

	var rv = strings.ReplaceAll(string(buffer), "\r", "")
	rv = strings.ReplaceAll(string(rv), "\n", "")
	rv = strings.ReplaceAll(string(rv), " ", "")
	return rv
}

func getBuildConfig(app_name, buildfilename string) map[string]interface{} {
	fmt.Println("begin getConfig")
	var rv map[string]interface{}
	//rv = make(map[string]interface{})

	data, err := goutils.ParseJSONObject(buildfilename)
	writeError(err)
	//fmt.Println("data", data)
	if err == nil {
		dataarr := data.([]interface{})
		//fmt.Println("dataarr", dataarr)
		if len(dataarr) > 0 {
			//loop through the array
			for i := 0; i < len(dataarr); i++ {
				item := dataarr[i].(map[string]interface{}) //interface
				if item["gitfolder"] == app_name {
					rv = item
					break
				}
			}
		}
	}
	return rv
}

func getAppConfigPath(env string) string {
	if env == "prod" {
		return "appconfig.prod.json"
	}
	return "appconfig.test.json"
}

func getBuildConfigPath() string {
	buildfilename := "buildconfig.json"
	retPath := ""
	exists := false

	checkpath, err := os.UserHomeDir()
	if err != nil {
		checkpath += path.Join(checkpath, buildfilename)
		exists, err = goutils.FileOrDirExists(checkpath)
		if exists {
			retPath = checkpath
		}
	}

	if retPath == "" {
		//check the current directory
		checkpath = buildfilename
		exists, err = goutils.FileOrDirExists(checkpath)
		if exists {
			retPath = checkpath
		}
	}

	if retPath == "" {
		//check one directory up
		checkpath = path.Join("../", buildfilename)
		exists, err = goutils.FileOrDirExists(checkpath)
		if exists {
			retPath = checkpath
		}
	}

	if retPath == "" {
		//check one directory up
		checkpath = path.Join("../../", buildfilename)
		exists, err = goutils.FileOrDirExists(checkpath)
		if exists {
			retPath = checkpath
		}
	}

	return retPath
}

func writeError(err error) {
	if err != nil {
		fmt.Println("ERROR:", err)
	}
}

func writeErrorMsg(message string, err error) {
	if err != nil {
		fmt.Println("ERROR:", message, err)
	}
}

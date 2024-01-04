package main

import (
	"errors"
	"flag"
	"fmt"
	goutils_data "goutils/data"
	goutils_io "goutils/io"
	goutils_web "goutils/web"
	"log"
	_ "math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	var err error
	//var ok bool
	var bc = BuildConfig{}

	bc.currentDir, err = goutils_io.GetCurrentDirectory()
	fmt.Println("currentDir=", bc.currentDir)

	// ## set build output directories from node build
	bc.builddir = "build"
	bc.buildappdir = "build/app"

	// ## get the directory name of the git folder
	bc.git_folder, err = getGitFolder()
	if bc.git_folder == "" {
		panic("gitFolder is empty")
	}
	bc.git_branch, err = getGitBranch()
	if bc.git_branch == "" {
		panic("gitBranch is empty")
	}
	bc.app_git_sha, err = getGitCommitSHA(bc.git_branch)
	if bc.app_git_sha == "" {
		panic("gitCommitHash is empty")
	}

	fmt.Printf("gitFolder=%s gitBranch=%s gitCommitHash=%s", bc.git_folder, bc.git_branch, bc.app_git_sha)

	// ## determine if master or test
	if bc.git_branch == "master" || bc.git_branch == "main" || bc.git_branch == "origin/master" || bc.git_branch == "origin/main" {
		bc.env_type = "prod"
		bc.node_env = "production"
	} else {
		bc.env_type = "test"
		bc.node_env = "development"
	}
	fmt.Println("env_type=", bc.env_type)

	// ## get build config file path
	bc.build_config_path = getBuildConfigPath(bc.env_type)
	if bc.build_config_path == "" {
		panic("build_config_path not found")
	}
	fmt.Println("build_config_path=", bc.build_config_path)

	// ## get the build configs
	err = getBuildConfig(&bc)
	if err != nil {
		panic("app config not found")
	}

	if bc.database_server_name == "" {
		panic("database_server_name is empty")
	}

	if bc.database_name == "" {
		panic("database_name is empty")
	}

	if bc.database_username == "" {
		panic("database_username is empty")
	}

	if bc.container_registry == "" {
		panic("container_registry is empty")
	}

	if bc.container_port == "" {
		panic("container_port is empty")
	}

	if bc.external_port == "" {
		panic("external_port is empty")
	}

	if bc.app_name == "" {
		panic("app_name is empty")
	}

	if bc.app_url == "" {
		panic("app_url is empty")
	}

	if bc.docker_manager_url == "" {
		panic("docker_manager_url is empty")
	}

	// ## get build date
	now := time.Now()
	bc.app_builddate = now.Format("2006-01-02 15:04") //YYYY-MM-dd HH mm

	// ## get app version from _version.txt
	err = getAppVersion(&bc)
	if err != nil {
		panic("could not read _version.txt")
	}
	bc.app_version = bc.app_version + now.Format(".200601021504")
	fmt.Println("app_version=", bc.app_version)

	// ## copy Dockerfile from source to build folder
	err = copyDockerFile(&bc)
	if err != nil {
		fmt.Println("ERROR: error writing to Dockerfile")
		panic(err)
	}
	fmt.Println("copyDockerFile finished")

	// ## update the appconfig variables
	err = copyAppConfig(&bc)
	if err != nil {
		fmt.Println("ERROR: failed to write appconfig.json")
		panic(err)
	}
	fmt.Println("appconfig.json copied")

	bc.buildDirFullPath = path.Join(bc.currentDir, bc.builddir)
	fmt.Println("buildDirFullPath=", bc.buildDirFullPath)
	err = buildContainer(&bc)
	if err != nil {
		fmt.Println("ERROR: failed to buildContainer")
		panic(err)
	}
	fmt.Println("buildContainer finished")

	err = publishContainer(&bc)
	if err != nil {
		fmt.Println("ERROR: failed to publishContainer")
		panic(err)
	}
	fmt.Println("publishContainer finished")

	err = deployToSwarm(&bc)
	if err != nil {
		fmt.Println("ERROR: failed to deployToSwarm")
		panic(err)
	}
	fmt.Println("deployToSwarm finished")

	err = checkAppState(bc.app_url)
	if err != nil {
		fmt.Println("ERROR: checkAppState  failed.")
		panic(err)
	}

	err = checkAppVersion(&bc)
	if err != nil {
		fmt.Println("ERROR: checkAppVersion  failed.")
		panic(err)
	}

	fmt.Println("build complete")

	os.Exit(0)

}

func getGitFolder() (string, error) {

	var rv string
	rv = goutils_data.ToString(os.Getenv("CURR_FOLDER"))
	if rv > "" {
		rv = filepath.Base(rv)
	}

	fmt.Println("CURR_FOLDER=", rv)

	if rv == "" {

		fmt.Println("get from git rev-parse --show-toplevel")

		buffer, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			log.Fatal("failed getGitFolder ", err)
			return "", err
		}
		rv = string(buffer)
		if rv > "" {
			rv = filepath.Base(rv)
		}
		rv = strings.ReplaceAll(string(rv), "\r", "")
		rv = strings.ReplaceAll(string(rv), "\n", "")
		rv = strings.ReplaceAll(string(rv), " ", "")

	}

	if rv == "" {
		return "", errors.New("unable to get GIT Folder")
	}

	return rv, nil
}

func getGitBranch() (string, error) {

	////git branch -r
	//origin/masterorigin/test
	//try to get from env
	// BUILD_ENV: test
	// NODE_ENV: development
	// RUNNER_TAGS: $CI_RUNNER_TAGS
	// GIT_BRANCH: $CI_COMMIT_BRANCH
	// GIT_COMMIT_REF_SLUG: ${CI_COMMIT_REF_SLUG}

	var branch string
	branch = os.Getenv("GIT_BRANCH")

	if branch == "" {
		//get from params
		flag.StringVar(&branch, "-git_branch", "", "")
		flag.Parse()
	}

	if branch == "" {
		//try to get from git folder, git branch --show-current
		buffer, err := exec.Command("git", "branch", "--show-current").Output()
		if err != nil {
			log.Fatal("failed getGitBranch ", err)
			return "", err
		}

		var output = string(buffer)
		if len(output) > 0 {
			var rv = strings.ReplaceAll(string(buffer), "\r", "")
			rv = strings.ReplaceAll(string(rv), "\n", "")
			rv = strings.ReplaceAll(string(rv), " ", "")
			branch = rv
		}
	}

	log.Println("branch=" + branch)

	if branch == "" {
		return "", errors.New("unable to get GIT Branch")
	}

	return branch, nil
}

func getGitCommitSHA(branch string) (string, error) {

	// throws an error due to the way gitlab pulls from the repo: buffer, err := exec.Command("git", "rev-parse", branch).Output()
	var app_git_sha string
	app_git_sha = goutils_data.ToString(os.Getenv("GIT_SHA"))
	if app_git_sha == "" {
		buffer, err := exec.Command("git", "rev-parse", branch).Output()
		app_git_sha = string(buffer)
		if err != nil {
			fmt.Println("ERROR on git rev-parse", err)
			return "", err
		}
	}

	var rv = strings.ReplaceAll(app_git_sha, "\r", "")
	rv = strings.ReplaceAll(string(rv), "\n", "")
	rv = strings.ReplaceAll(string(rv), " ", "")

	if rv == "" {
		return "", errors.New("unable to get GIT SHA")
	}

	return rv, nil
}

func getBuildConfigPath(env_typ string) string {
	buildfilename := "buildconfig-" + env_typ + ".json"
	retPath := ""
	exists := false

	checkpath, err := os.UserHomeDir()
	log.Println("UserHomeDir=", checkpath)
	if err == nil {
		checkpath = path.Join(checkpath, buildfilename)
		log.Println("UserHomeDir + buildfilename=", checkpath)
		exists, err = goutils_io.FileOrDirExists(checkpath)
		if exists {
			retPath = checkpath
		}
	}

	if retPath == "" {
		//check the current directory
		checkpath = buildfilename
		log.Println("buildfilename=", checkpath)
		exists, err = goutils_io.FileOrDirExists(checkpath)
		if exists {
			retPath = checkpath
		}
	}

	if retPath == "" {
		//check one directory up
		checkpath = path.Join("../", buildfilename)
		log.Println("../buildfilename=", checkpath)
		exists, err = goutils_io.FileOrDirExists(checkpath)
		if exists {
			retPath = checkpath
		}
	}

	if retPath == "" {
		//check one directory up
		checkpath = path.Join("../../", buildfilename)
		log.Println("../../buildfilename=", checkpath)
		exists, err = goutils_io.FileOrDirExists(checkpath)
		if exists {
			retPath = checkpath
		}
	}

	return retPath
}

func getBuildConfig(bc *BuildConfig) error {

	fmt.Println("begin getBuildConfig " + bc.git_folder)

	data, err := goutils_data.ParseJSONFromFile(bc.build_config_path)
	if err != nil {
		fmt.Println("unable to ParseJSONObject")
		fmt.Println(err)
		return err
	}

	//fmt.Println("data", data)
	if err == nil {
		dataarr := data.([]interface{})
		//fmt.Println("dataarr", dataarr)
		if len(dataarr) > 0 {
			//loop through the array
			for i := 0; i < len(dataarr); i++ {
				item := dataarr[i].(map[string]interface{}) //interface
				if item["gitfolder"] == bc.git_folder {

					bc.app_name = goutils_data.ToString(item["app_name"])
					bc.database_server_name = goutils_data.ToString(item["database_server_name"])
					bc.database_name = goutils_data.ToString(item["database_name"])
					bc.database_username = goutils_data.ToString(item["database_username"])
					bc.database_password = goutils_data.ToString(item["database_password"])
					bc.external_port = goutils_data.ToString(item["external_port"])
					bc.container_port = goutils_data.ToString(item["container_port"])
					bc.container_registry = goutils_data.ToString(item["container_registry"])
					bc.app_url = goutils_data.ToString(item["app_url"])
					bc.app_url_appconfig = goutils_data.ToString(item["app_url_appconfig"])
					bc.docker_manager_url = goutils_data.ToString(item["docker_manager_url"])
					bc.container_replicas = goutils_data.ToString(item["container_replicas"])
					bc.container_servicename = goutils_data.ToString(item["container_servicename"])
					bc.appconfig_dst = goutils_data.ToString(item["appconfig_dst"])

					fmt.Println("app_name=", bc.app_name)
					fmt.Println("app_url=", bc.app_url)
					fmt.Println("app_url_appconfig=", bc.app_url_appconfig)
					fmt.Println("database_server_name=", bc.database_server_name)
					fmt.Println("database_name=", bc.database_name)
					fmt.Println("database_username=", bc.database_username)
					fmt.Println("database_password=", "*****")
					fmt.Println("external_port=", bc.external_port)
					fmt.Println("container_port=", bc.container_port)
					fmt.Println("container_registry=", bc.container_registry)

					break
				}
			}
		}
	}

	return nil
}

func getAppVersion(bc *BuildConfig) error {
	var err error
	bc.app_version, err = goutils_io.ReadFile("_version.txt")
	bc.app_version = strings.ReplaceAll(bc.app_version, "\n", "")
	bc.app_version = strings.ReplaceAll(bc.app_version, "\r", "")
	bc.app_version = strings.ReplaceAll(bc.app_version, " ", "")
	if err != nil {
		return err
	}
	return nil
}

func copyAppConfig(bc *BuildConfig) error {
	var content string
	var err error
	content, err = goutils_io.ReadFile("appconfig.json.template")
	if err != nil {
		fmt.Println("ERROR: reading appconfig.json.template")
		return err
	}
	content = strings.ReplaceAll(content, "{app_name}", bc.app_name)
	content = strings.ReplaceAll(content, "{app_git_sha}", bc.app_git_sha)
	content = strings.ReplaceAll(content, "{app_builddate}", bc.app_builddate)
	content = strings.ReplaceAll(content, "{app_version}", bc.app_version)

	err = goutils_io.WriteFile(bc.appconfig_dst+"/appconfig.json", content, true)
	if err != nil {
		return err
	}
	return nil
}

func copyDockerFile(bc *BuildConfig) error {
	// ## populate Dockerfile ENV variables
	var err error
	var dockerFileContents string
	dockerFileContents, err = goutils_io.ReadFile("Dockerfile")
	if err != nil {
		fmt.Println(err)
		return err
	}
	if dockerFileContents != "" {

		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{app_name}", bc.app_name)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{app_version}", bc.app_version)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{app_builddate}", bc.app_builddate)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{app_git_sha}", bc.app_git_sha)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{database_name}", bc.database_name)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{database_password}", bc.database_password)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{database_server_name}", bc.database_server_name)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{database_username}", bc.database_username)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{container_port}", bc.container_port)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{external_port}", bc.external_port)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{NODE_ENV}", bc.node_env)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{adminapi_host}", bc.adminapi_host)
		dockerFileContents = strings.ReplaceAll(dockerFileContents, "{adminapi_port}", bc.adminapi_port)

		//write new file
		err = goutils_io.WriteFile(bc.builddir+"/Dockerfile", dockerFileContents, true)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func buildContainer(bc *BuildConfig) error {

	var imagename string
	var output string
	var err error

	//- buildah bud -t node-webserver-prod:1.0
	// docker build -t node-webserver-prod:1.11 .
	imagename = bc.container_registry + "/" + bc.container_servicename + ":" + bc.app_version
	output, err = goutils_io.ExecCMD(bc.buildDirFullPath, "docker", "build", "-t", imagename, ".")
	if err != nil {
		return err
	}

	//check if the build is in the local images
	//curl --unix-socket /var/run/docker.sock http://localhost/images/json
	//output, err = goutils.ExecCMD("", "docker", "image", "ls")
	//the query the unix socket returns a json string
	output, err = goutils_io.ExecCMD("", "curl", "--unix-socket", "/var/run/docker.sock", "http://localhost/images/json")
	if err != nil {
		fmt.Println("ERROR: failed exec curl --unix-socket /var/run/docker.sock http://localhost/images/json")
		return err
	}
	fmt.Println(output)

	if strings.Index(output, imagename) < 0 {
		return errors.New("docker image not found from local docker " + imagename)
	}

	return nil
}

func publishContainer(bc *BuildConfig) error {
	// podman push node-webserver:app_version docker.corrections-tech.com/node-webserver:app_version
	// example to rename a tag
	// docker image tag node-webserver-prod:1.11 docker.corrections-tech.com/node-webserver-prod:1.11
	// docker push node-webserver:1.11 docker.corrections-tech.com/node-webserver:1.11
	_, err := goutils_io.ExecCMD("", "docker", "push", bc.container_registry+"/"+bc.container_servicename+":"+bc.app_version)
	if err != nil {
		return err
	}
	//check registry for the tags
	//	curl -X GET https://docker.corrections-tech.com/v2/node-webserver-prod/tags/list
	// {"name":"node-webserver-prod","tags":["1.2","1.11","1.3","1.8","1.0","1.4","1.5","1.6","1.9","1.1","latest","1.7"]}
	// curl https://prod-dockerregistry.corp.loc/v2/node-webserver-prod/tags/list
	schema := "https"
	var body string
	if strings.Index(bc.container_registry, ":5000") >= 0 {
		schema = "http"
	}

	url := schema + "://" + bc.container_registry + "/v2/" + bc.container_servicename + "/tags/list"
	fmt.Println(url)
	body, err = goutils_web.HTTPGetBody(url)
	if err != nil {
		fmt.Println("unable to get container_registry" + url)
		return err
	}

	fmt.Println(body)

	if !(strings.Index(body, bc.container_servicename) > -1) {
		fmt.Println("container_servicename not found " + bc.container_servicename)
		return errors.New("container_servicename not found " + bc.container_servicename)
	}

	if !(strings.Index(body, "\""+bc.app_version+"\"") > -1) {
		fmt.Println("app_version not found " + bc.app_version)
		return errors.New("app_version not found " + bc.app_version)
	}

	return nil
}

func deployToSwarm(bc *BuildConfig) error {
	fmt.Println("deployToSwarm ...")
	var serviceExists bool
	var output string
	var err error
	var jsond string
	var url string
	var service_index string

	serviceExists = false
	imagename := bc.container_registry + "/" + bc.container_servicename + ":" + bc.app_version

	//docker service create --name node-webserver-prod -p 8002:8080 docker.corrections-tech.com/node-webserver-prod:1.0
	// output, err := goutils.ExecCMD("", "docker", "service", "ls")
	// if err != nil {
	// 	fmt.Println("ERROR: failed on deployToSwarm - get service list")
	// 	return err
	// }

	output, err = goutils_web.HTTPGetBody(bc.docker_manager_url + "/services/" + bc.container_servicename)
	if err != nil || output == "" {
		fmt.Println("ERROR: failed to get service from " + bc.docker_manager_url + "/services")
		return err
	}

	fmt.Println("services=", output)

	//docker.corrections-tech.com/node-webserver-prod:1.6
	//"Name": "node-webserver-prod"
	if strings.Index(output, "\""+bc.container_servicename+"\"") > -1 {
		serviceExists = true
	}
	fmt.Println("serviceExists=", serviceExists)

	var mountstr string

	mountstr = `, "Mounts":[
		{
			"Target": "/nfsmount1",
			"Source": "nfsmount1",
			"Type": "volume",
			"ReadyOnly": false
		}
	]`

	if serviceExists {

		//update the service
		//docker service update --image docker.corrections-tech.com/node-webserver-prod:1.6 node-webserver-prod
		//output, err = goutils.ExecCMD("", "docker", "service", "update", "--image", imagename, app_name, "--force", "--detach")

		//get the service_index
		url = bc.docker_manager_url + "/services/" + bc.container_servicename
		output, err = goutils_web.HTTPGetBody(url)
		if err != nil {
			log.Println("unable to get services" + url)
			return err
		}

		//output := `{"ID":"umb8mcr05k0a9lmn0a5cxnszv","Version":{"Index":1821},"CreatedAt":"2023-10-23T23:00:41.291429436Z","UpdatedAt":"2023-10-24T00:29:57.150690114Z","Spec":{"Name":"nginx1","Labels":{},"TaskTemplate":{"ContainerSpec":{"Image":"nginx:stable-alpine3.17-slim","Isolation":"default"},"RestartPolicy":{"Condition":"any","MaxAttempts":3},"ForceUpdate":1,"Runtime":"container"},"Mode":{"Replicated":{"Replicas":3}}},"PreviousSpec":{"Name":"nginx1","Labels":{},"TaskTemplate":{"ContainerSpec":{"Image":"nginx:latest","Isolation":"default"},"RestartPolicy":{"Condition":"any","MaxAttempts":3},"ForceUpdate":1,"Runtime":"container"},"Mode":{"Replicated":{"Replicas":3}}},"Endpoint":{"Spec":{}},"UpdateStatus":{"State":"completed","StartedAt":"2023-10-24T00:29:32.05242763Z","CompletedAt":"2023-10-24T00:29:57.15066874Z","Message":"update completed"}}`
		jsonobj := make(map[string]interface{})

		err := goutils_data.ParseJSONObject(output, &jsonobj)
		if err != nil {
			log.Println("unable to parse json" + output)
			return err
		}

		version := jsonobj["Version"]
		versionmap := version.(map[string]interface{})
		service_index = goutils_data.ToString(versionmap["Index"])

		log.Println(versionmap)

		jsond = `{
		"Name": "` + bc.container_servicename + `",
		"TaskTemplate": {
			"ContainerSpec": {
				"Image": "` + imagename + `",
				"Init": false,
				"StopGracePeriod": 10000000000,
				"DNSConfig": {},
				"Isolation": "default"
				` + mountstr + `
			},
			"Placement": {
				"Platforms": [
					{
						"Architecture": "amd64",
						"OS": "linux"
					}
				]
			},
			"ForceUpdate": 1,
			"Runtime": "container"
		},
		"Mode": {
			"Replicated": {
				"Replicas": ` + bc.container_replicas + `
			}
		},
		"EndpointSpec": {
			"Mode": "vip",
			"Ports": [
				{
					"Protocol": "tcp",
					"TargetPort": ` + bc.container_port + `,
					"PublishedPort": ` + bc.external_port + `,
					"PublishMode": "ingress"
				}
			]
		}
	}`

		goutils_io.WriteFile("build/service.json", jsond, true)

		url = bc.docker_manager_url + "/services/" + bc.container_servicename + "/update?version=" + service_index

		fmt.Println(url)
		fmt.Println(jsond)

		//curl -X POST -H "Content-Type: application/json" -d @service-config.json http://10.20.1.155:2375/services/create
		output, err = goutils_web.HTTPostJson(url, jsond)
		if err != nil {
			fmt.Println("failed to deployToSwarm")
			return err
		}
		fmt.Println(output)

	} else {
		//create the service
		//docker service create --name node-webserver -p 8001:8080 docker.corrections-tech.com/node-webserver
		//_, err = goutils.ExecCMD("", "docker", "service", "create", "--name", app_name, "-p", external_port+":"+container_port, imagename)
		jsond = `{
			"Name": "` + bc.container_servicename + `",
			"TaskTemplate": {
				"ContainerSpec": {
					"Image": "` + imagename + `",
					"Init": false,
					"StopGracePeriod": 10000000000,
					"DNSConfig": {},
					"Isolation": "default"
					` + mountstr + `
				},
				"Placement": {
					"Platforms": [
						{
							"Architecture": "amd64",
							"OS": "linux"
						}
					]
				},
				"ForceUpdate": 1,
				"Runtime": "container"
			},
			"Mode": {
				"Replicated": {
					"Replicas": ` + bc.container_replicas + `
				}
			},
			"EndpointSpec": {
				"Mode": "vip",
				"Ports": [
					{
						"Protocol": "tcp",
						"TargetPort": ` + bc.container_port + `,
						"PublishedPort": ` + bc.external_port + `,
						"PublishMode": "ingress"
					}
				]
			}
		}`
		url = bc.docker_manager_url + "/services/create"

		fmt.Println(url)
		fmt.Println(jsond)

		//curl -X POST -H "Content-Type: application/json" -d @service-config.json http://10.20.1.155:2375/services/create
		output, err = goutils_web.HTTPostJson(url, jsond)
		if err != nil {
			fmt.Println("failed to deployToSwarm")
			return err
		}
		fmt.Println(output)
	}

	//wait and check service every 30 seconds
	var maxloops = 10
	var counter = 0
	for {

		counter = counter + 1

		//output, err = goutils.ExecCMD("", "docker", "service", "ls")
		output, err = goutils_web.HTTPGetBody(bc.docker_manager_url + "/services/" + bc.container_servicename)
		if err != nil || output == "" {
			fmt.Println("ERROR: failed to get service from " + bc.docker_manager_url + "/services")
			return err
		}

		fmt.Println("services=", output)

		//prod-dockerregistry.corp.loc/node-webserver:1.20
		if strings.Index(output, imagename) >= 0 {
			break
		} else {
			fmt.Println("image not found in output " + imagename)
		}

		timer := time.NewTimer(30 * time.Second)
		<-timer.C

		if counter >= maxloops {
			return errors.New("failed to verify docker service with image name " + imagename)
		}
	}

	return nil
}

func checkAppState(url string) error {
	fmt.Println("checking app state...")
	var code = 0
	var err error

	var maxloops = 10
	var counter = 0
	for {

		counter = counter + 1

		code, err = goutils_web.HTTPGetCode(url)
		if err != nil {
			fmt.Println("error on HTTPGetCode")
			fmt.Println(err)
		}

		if code == 200 {
			return nil
		}

		timer := time.NewTimer(30 * time.Second)
		<-timer.C

		if counter >= maxloops {
			return errors.New("ERROR: unable to HTTP GET " + url)
		}
	}
}

func checkAppVersion(bc *BuildConfig) error {

	url := bc.app_url + bc.app_url_appconfig

	fmt.Println("checkAppVersion " + url + ", version " + bc.app_version)

	var maxloops = 10
	var counter = 0
	for {

		counter = counter + 1

		body, err := goutils_web.HTTPGetBody(url)
		if err != nil {
			return err
		}

		fmt.Println(body)

		jsonData := make(map[string]interface{})
		err = goutils_data.ParseJSONObject(body, &jsonData)
		if err != nil {
			fmt.Println("error parsing json")
			return err
		}

		if goutils_data.ToString(jsonData["app_version"]) == bc.app_version {
			break
		}

		timer := time.NewTimer(30 * time.Second)
		<-timer.C

		if counter >= maxloops {
			return errors.New("app versions do not match " + bc.app_version)
		}
	}

	return nil
}

type BuildConfig struct {
	app_name              string
	database_server_name  string
	database_name         string
	database_username     string
	database_password     string
	external_port         string
	container_port        string
	container_registry    string
	app_url               string
	app_url_appconfig     string
	docker_manager_url    string
	container_replicas    string
	container_servicename string
	app_builddate         string
	app_version           string
	buildDirFullPath      string
	builddir              string
	buildappdir           string
	git_folder            string
	git_branch            string
	app_git_sha           string
	node_env              string
	currentDir            string
	build_config_path     string
	env_type              string
	adminapi_host         string
	adminapi_port         string
	appconfig_dst         string
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/bndr/gojenkins"
)

type s_jenkinsCred struct {
	jenkinsUrl string
	user       string
	password   string
}

func makeDir(dir string) {
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		log.Fatal(err)
	} else {
		logrus.Info("\t", dir, " folder created")
	}
}

// Create the SD struct folder
func createFolderStructure() {
	logrus.Info("Create folder:")

	makeDir("SDFile/switch")
	makeDir("SDFile/ReiNX/titles")
	makeDir("SDFile/tinfoil/nsp")
	makeDir("download")
	makeDir("release_list")

	fmt.Println("")
}

func setupLogrus() {
	Formatter := new(logrus.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	logrus.SetFormatter(Formatter)
}

func downloadLatestRelease(jenkins *gojenkins.Jenkins, project string) (err error) {

	logrus.Info(project, ":")
	logrus.Info("\tGet job ...")
	build, err := jenkins.GetJob(project)
	if err != nil {
		return err
	}

	logrus.Info("\tSearch last successful build ...")
	lastSuccessBuild, err := build.GetLastSuccessfulBuild()
	if err != nil {
		return err
	}

	lastBuild := project + "_" + strconv.Itoa(int(lastSuccessBuild.GetBuildNumber()))
	if _, err = os.Stat("release_list/" + lastBuild); !os.IsNotExist(err) {
		logrus.Warn("\t", project, " is up to date")
		return err
	}

	artifacts := lastSuccessBuild.GetArtifacts()

	logrus.Info("\tDownload release ...")
	for _, a := range artifacts {
		a.SaveToDir("download")
	}
	logrus.Info("\tDownload Success !")

	if _, err = os.Create("release_list/" + lastBuild); err != nil {
		logrus.Error(err)
	}

	return err
}

// Setup homebrew to download
func getHomebrewList() (homebrewList []string) {

	homebrewList = append(homebrewList, "Tinfoil")
	homebrewList = append(homebrewList, "appstore-nx")
	homebrewList = append(homebrewList, "ftpd")
	homebrewList = append(homebrewList, "Checkpoint")

	return
}

func getJenkinsCredentials() (result map[string]interface{}) {

	jsonFile, err := os.Open("credentials.json")
	if err != nil {
		logrus.Fatal(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &result)

	return
}

func connectToJenkins(jenkinsCredentials map[string]interface{}) (jenkins *gojenkins.Jenkins) {

	jenkins = gojenkins.CreateJenkins(nil, jenkinsCredentials["jenkins_url"].(string), jenkinsCredentials["user"].(string), jenkinsCredentials["password"].(string))

	_, err := jenkins.Init()
	if err != nil {
		logrus.Fatal(err)
	} else {
		logrus.Info("Successfuly Connected to jenkins !\n\n")
	}

	return
}

func main() {

	setupLogrus()

	createFolderStructure()

	jenkinsCredentials := getJenkinsCredentials()

	// Connect to jenkins
	jenkins := connectToJenkins(jenkinsCredentials)

	// Get all homebrew to install
	homebrewList := getHomebrewList()

	// download all homebrew
	for idx := range homebrewList {
		err := downloadLatestRelease(jenkins, homebrewList[idx])
		if err != nil {
			logrus.Error(err)
		}

		fmt.Println("")
	}
}

package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

func downloadLatestRelease(jenkins *gojenkins.Jenkins, project string) (fileName []gojenkins.Artifact, err error) {

	logrus.Info(project, ":")
	logrus.Info("\tGet job ...")
	build, err := jenkins.GetJob(project)
	if err != nil {
		return nil, err
	}

	logrus.Info("\tSearch last successful build ...")
	lastSuccessBuild, err := build.GetLastSuccessfulBuild()
	if err != nil {
		return nil, err
	}

	lastBuild := project + "_" + strconv.Itoa(int(lastSuccessBuild.GetBuildNumber()))
	if _, err = os.Stat("release_list/" + lastBuild); !os.IsNotExist(err) {
		logrus.Warn("\t", project, " is up to date")
		return nil, err
	}

	artifacts := lastSuccessBuild.GetArtifacts()

	logrus.Info("\tDownload release ...")
	for _, a := range artifacts {
		_, err = a.SaveToDir("download")
		if err != nil {
			return nil, err
		}
	}
	logrus.Info("\tDownload Success !")

	if _, err = os.Create("release_list/" + lastBuild); err != nil {
		logrus.Error(err)
	}

	return artifacts, err
}

// Setup homebrew to download
func getHomebrewList() (homebrewList []string) {

	homebrewList = append(homebrewList, "Tinfoil")
	homebrewList = append(homebrewList, "appstore-nx")
	homebrewList = append(homebrewList, "ftpd")
	homebrewList = append(homebrewList, "Checkpoint")
	homebrewList = append(homebrewList, "NX-Shell")
	homebrewList = append(homebrewList, "ReiNX")
	homebrewList = append(homebrewList, "Atmosphere-NX")

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

	// allow anonymous login
	if jenkinsCredentials["user"].(string) == "" && jenkinsCredentials["password"].(string) == "" {
		jenkins = gojenkins.CreateJenkins(nil, jenkinsCredentials["jenkins_url"].(string))
	} else {
		jenkins = gojenkins.CreateJenkins(nil, jenkinsCredentials["jenkins_url"].(string), jenkinsCredentials["user"].(string), jenkinsCredentials["password"].(string))
	}

	_, err := jenkins.Init()
	if err != nil {
		logrus.Fatal(err)
	} else {
		logrus.Info("Successfuly Connected to jenkins !\n\n")
	}

	return
}

func installInSD(file string) (err error) {

	if strings.HasSuffix(file, ".nro") {
		err = os.Rename("download/"+file, "SDFile/switch/"+file)
	} else if strings.HasSuffix(file, "zip") {
		if err = unZipFile("download/"+file, "SDFile/"); err != nil {
			logrus.Error(err)
		}
	}

	return
}

func downloadFile(url string, filePath string) error {

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	logrus.Info(filePath, " download success !")
	return nil
}

func unZipFile(zipFile string, out string) (err error) {
	var (
		dir, file bool = true, true
	)

	logrus.Info("Unzip: ", zipFile)

	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return
	}

	for _, zipFile := range zipReader.Reader.File {

		zippedFile, err := zipFile.Open()
		if err != nil {
			return err
		}
		defer zippedFile.Close()

		targetDir := out
		extractedFilePath := filepath.Join(
			targetDir,
			zipFile.Name,
		)

		if zipFile.FileInfo().IsDir() {
			if dir == true {
				logrus.Info("Directory Created:")
			}
			logrus.Info("\t", extractedFilePath, "/")

			os.MkdirAll(extractedFilePath, zipFile.Mode())
			file = true
			dir = false
		} else {
			if file == true {
				logrus.Info("File extracted:")
			}
			logrus.Info("\t", zipFile.Name)

			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				zipFile.Mode(),
			)
			if err != nil {
				return err
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				return err
			}
			file = false
			dir = true
		}
	}

	fmt.Println("")
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

	// for ReiNX only
	if _, err := os.Stat("SDFile/ReiNX/titles/010000000000100D"); err != nil {
		if err := downloadFile("https://reinx.guide/u/010000000000100D.zip", "download/010000000000100D.zip"); err != nil {
			logrus.Error("download 010000000000100D.zip: ", err)
		} else {
			if err = unZipFile("download/010000000000100D.zip", "SDFile/ReiNX/titles/"); err != nil {
				logrus.Error(err)
			}
		}
	}

	// download all homebrew
	for idx := range homebrewList {
		fileName, err := downloadLatestRelease(jenkins, homebrewList[idx])
		if err != nil {
			logrus.Error("\t", err)
		}

		// Move file to Sd card folder
		for _, file := range fileName {
			if err = installInSD(file.FileName); err != nil {
				return
			}
		}

		fmt.Println("")
	}
	os.RemoveAll("download")

}

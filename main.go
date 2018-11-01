package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	pb "gopkg.in/cheggaaa/pb.v1"
)

type release struct {
	Assets []asset `json:"assets"`
}

type asset struct {
	URL string `json:"browser_download_url"`
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
	logrus.Info("Create folder :")

	makeDir("SDFile/switch")
	makeDir("SDFile/tinfoil")
	makeDir("download")

	fmt.Println("")
}

func downloadFile(filePath string, url string) error {

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

	// Setup progress bar
	bar := pb.New(int(resp.ContentLength)).SetUnits(pb.U_BYTES)

	bar.Start()
	defer bar.Finish()

	reader := bar.NewProxyReader(resp.Body)

	// Write the body to file
	_, err = io.Copy(out, reader)
	if err != nil {
		return err
	}

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

func getLastRelease(filePath string, repo string) (string, error) {

	// format url
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	// perform request
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logrus.Error("HTTP Response Status: ", resp.StatusCode, http.StatusText(resp.StatusCode), " for ", repo)
		return "", errors.New(http.StatusText(resp.StatusCode))
	}

	// read
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// get download link
	r := &release{}
	err = json.Unmarshal(body, r)
	if err != nil {
		return "", err
	}

	// url as string
	if len(r.Assets) > 1 {
		for _, v := range r.Assets {
			if strings.HasSuffix(v.URL, ".nro") {
				url = v.URL
				break
			}
		}
	} else {
		url = r.Assets[0].URL
	}

	// get filename with extension
	split := strings.Split(url, "/")
	fileName := split[len(split)-1]

	if _, err := os.Stat(filePath + fileName); os.IsNotExist(err) {
		// Download the file
		logrus.Info("Download: ", fileName)

		err = downloadFile(filePath+fileName, url)
		if err != nil {
			return "", err
		}

		fmt.Println("")
		return filePath + fileName, err
	}

	return "", err
}

func InstallPackages(downloadFolder, repository, installationPath string) (err error) {

	release, err := getLastRelease(downloadFolder, repository)
	if err != nil {
		return
	} else if release == "" {
		logrus.Info(repository, " is up to date")
	} else if filepath.Ext(release) != ".zip" {
		split := strings.Split(release, "/")
		fileName := split[len(split)-1]
		// move file in folder then create link used for check if package is up to date
		os.Rename(release, installationPath+fileName)
		os.Symlink(installationPath+fileName, "download/"+fileName)
	} else if err = unZipFile(release, installationPath); err != nil {
		return
	}

	return
}

func main() {
	Formatter := new(logrus.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	logrus.SetFormatter(Formatter)

	createFolderStructure()

	downloadList := [][]string{
		//downloadFolder, repository, installationPAth
		{"download/", "Reisyukaku/ReiNX", "SDFile/"},
		{"download/", "switchbrew/nx-hbmenu", "SDFile/"},
		{"download/", "vgmoose/hb-appstore", "SDFile/"},
		{"download/", "FlagBrew/Checkpoint", "SDFile/switch/"},
		{"download/", "mtheall/ftpd", "SDFile/switch/"},
		{"download/", "Reisyukaku/ReiNXToolkit", "SDFile/switch/"},
		{"download/", "joel16/NX-Shell", "SDFile/switch/"},
	}

	for i := range downloadList {
		InstallPackages(downloadList[i][0], downloadList[i][1], downloadList[i][2])
	}
}

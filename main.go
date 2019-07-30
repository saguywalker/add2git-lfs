package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	rice "github.com/GeertJohan/go.rice"

	"github.com/labstack/echo"
)

const uploadsDir = "sample-files/"

func main() {
	os.MkdirAll(filepath.Join(".", uploadsDir), os.ModePerm)
	err := initLfs()
	if err != nil {
		panic(err)
	}

	e := echo.New()
	assetHandler := http.FileServer(rice.MustFindBox("public").HTTPBox())
	e.GET("/", echo.WrapHandler(assetHandler))
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", assetHandler)))
	e.POST("/upload", handleUpload)
	e.POST("/pushfiles", handlePushFiles)
	go open("http://localhost:12358/")
	e.Logger.Fatal(e.Start(":12358"))
}

func handleUpload(c echo.Context) error {
	/*fileInfo, err := c.FormFile("file")
	if err != nil {
		return c.String(http.StatusBadRequest, "Error when parsing files")
	}
	fullname := uploadsDir + fileInfo.Filename

	file, err := fileInfo.Open()
	if err != nil {
		message := fmt.Sprintf("Error when opening %v", fullname)
		return c.String(http.StatusBadRequest, message)
	}

	out, err := os.OpenFile(fullname, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		message := fmt.Sprintf("Error when uploading file %v", fullname)
		return c.String(http.StatusExpectationFailed, message)
	}

	io.Copy(out, file)

	return c.String(http.StatusOK, "Files are uploaded")*/
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["files"]

	for _, fileInfo := range files {
		fullname := uploadsDir + fileInfo.Filename

		file, err := fileInfo.Open()
		if err != nil {
			message := fmt.Sprintf("Error when opening %v", fullname)
			return c.String(http.StatusBadRequest, message)
		}
		defer file.Close()

		out, err := os.OpenFile(fullname, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			message := fmt.Sprintf("Error when uploading file %v", fullname)
			return c.String(http.StatusExpectationFailed, message)
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			message := fmt.Sprintf("Error when copying file %v", fullname)
			return c.String(http.StatusExpectationFailed, message)
		}
	}
	return c.String(http.StatusOK, "Files are uploaded")
}

func initLfs() error {
	var err error
	initLfsCmd := "git lfs install && git lfs track \"sample-files/*\" && git add .gitattributes"
	if runtime.GOOS == "windows" {
		_, err = exec.Command("cmd", "/C", initLfsCmd).Output()
	} else {
		_, err = exec.Command("bash", "-c", initLfsCmd).Output()
	}
	if err != nil {
		return err
	}

	return nil
}

func gitAddFile(filename string) error {
	var err error
	addCmd := fmt.Sprintf("git add %v", filename)
	if runtime.GOOS == "windows" {
		_, err = exec.Command("cmd", "/C", addCmd).Output()
	} else {
		_, err = exec.Command("bash", "-c", addCmd).Output()
	}
	if err != nil {
		return err
	}

	return nil
}

func gitCommitShell() error {
	var err error
	if runtime.GOOS == "windows" {
		commitCmd := "git commit -m upload-sample-files"
		_, err = exec.Command("cmd", "/C", commitCmd).Output()
	} else {
		commitCmd := "git commit -m upload-sample-files"
		_, err = exec.Command("bash", "-c", commitCmd).Output()
	}
	if err != nil {
		return err
	}
	return nil
}

func gitPushShell() error {
	var err error
	gitPushCmd := "git push origin master"
	if runtime.GOOS == "windows" {
		_, err = exec.Command("cmd", "/C", gitPushCmd).Output()
	} else {
		_, err = exec.Command("bash", "-c", gitPushCmd).Output()
	}
	if err != nil {
		return err
	}

	return nil
}

func handlePushFiles(c echo.Context) error {
	err := gitAddFile("sample-files")
	if err != nil {
		errMsg := fmt.Sprintf("Error when running \"git add sample-files\"\n%s", err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	err = gitCommitShell()
	if err != nil {
		errMsg := fmt.Sprintf("Error when running \"git commit\"\n%s", err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	err = gitPushShell()
	if err != nil {
		errMsg := fmt.Sprintf("Error when running \"git push\"\n%s", err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

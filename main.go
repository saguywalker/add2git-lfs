package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	rice "github.com/GeertJohan/go.rice"

	"github.com/labstack/echo"
)

const uploadsDir = "sample-files/"

func main() {
	/*c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err := gitAddFile("sample-files")
		if err != nil {
			panic(err)
		}

		err = gitCommitShell()
		if err != nil {
			panic(err)
		}

		err = gitPushShell()
		if err != nil {
			panic(err)
		}
		os.Exit(1)
	}()*/

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
	e.Logger.Fatal(e.Start(":12358"))
}

func handleUpload(c echo.Context) error {
	fileInfo, err := c.FormFile("file")
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

	return c.String(http.StatusOK, "Files are uploaded")
}

func initLfs() error {
	initLfsCmd := exec.Command("bash", "-c", "git lfs install && git lfs track \"sample-files/*\" && git add .gitattributes")
	out, err := initLfsCmd.Output()
	if err != nil {
		return err
	}
	fmt.Println(string(out))

	return nil
}

func gitAddFile(filename string) error {
	addCmd := fmt.Sprintf("git add %v", filename)
	gitAddCmd := exec.Command("bash", "-c", addCmd)
	_, err := gitAddCmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func gitCommitShell() error {
	gitCommitCmd := exec.Command("bash", "-c", "git commit -m \"upload sample files \"")
	out, err := gitCommitCmd.Output()
	if err != nil {
		return err
	}
	fmt.Println(string(out))

	return nil
}

func gitPushShell() error {
	gitPushCmd := exec.Command("bash", "-c", "git push origin master")
	out, err := gitPushCmd.Output()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(out))
	return nil
}

func handlePushFiles(c echo.Context) error {
	fmt.Println("In handlePushFiles")
	err := gitAddFile("sample-files")
	if err != nil {
		return c.String(http.StatusExpectationFailed, err.Error())
	}

	err = gitCommitShell()
	if err != nil {
		return c.String(http.StatusExpectationFailed, err.Error())
	}

	err = gitPushShell()
	if err != nil {
		return c.String(http.StatusExpectationFailed, err.Error())
	}

	return c.String(http.StatusOK, "/")
}

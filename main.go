package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/labstack/echo"
)

const uploadsDir = "sample-files/"

func main() {
	fmt.Println("add2git-web")
	e := echo.New()
	e.Static("/public", "public")
	e.File("/", "views/upload.html")
	e.POST("/upload", handleUpload)

	e.Logger.Fatal(e.Start(":12358"))
}

func handleUpload(c echo.Context) error {
	info, err := c.FormFile("file")
	if err != nil {
		return c.String(http.StatusBadRequest, "Error when uploading files")
	}
	fullname := uploadsDir + info.Filename
	out, err := os.OpenFile(fullname, os.O_WRONLY|os.O_CREATE, 0666)

	err = gitAddFile(fullname)
	if err != nil {
		message := fmt.Sprintf("Error when running git add %s", fullname)
		return c.String(http.StatusExpectationFailed, message)
	}

	err = gitCommitShell()
	if err != nil {
		message := fmt.Sprintf("Error when running git commit %s", fullname)
		return c.String(http.StatusExpectationFailed, message)
	}

	err = gitPushShell()
	if err != nil {
		message := fmt.Sprintf("Error when running git push (%s)", fullname)
		return c.String(http.StatusExpectationFailed, message)
	}

	return c.String(http.StatusOK, "Files uploaded")
}

func gitAddFile(filename string) error {
	addCmd := fmt.Sprintf("git add %v", filename)
	fmt.Println(addCmd)
	gitAddCmd := exec.Command("bash", "-c", "git add sample-files")
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
		return err
	}
	fmt.Println(string(out))
	return nil
}

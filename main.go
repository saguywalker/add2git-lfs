package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/labstack/echo"
)

const uploadsDir = "sample-files/"

func main() {
	c := make(chan os.Signal)
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
	}()

	fmt.Println("add2git-web")
	e := echo.New()
	//e.Use(middleware.Logger())
	e.Static("/public", "public")
	e.File("/", "views/upload.html")
	e.POST("/upload", handleUpload)

	e.Logger.Fatal(e.Start(":12358"))
	/*
		err := gitPushShell()
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
	*/
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
	/*
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
		}*/

	return c.String(http.StatusOK, "Files uploaded")
}

func gitAddFile(filename string) error {
	addCmd := fmt.Sprintf("git add %v", filename)
	fmt.Println(addCmd)
	gitAddCmd := exec.Command("bash", "-c", addCmd)
	_, err := gitAddCmd.Output()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func gitCommitShell() error {
	gitCommitCmd := exec.Command("bash", "-c", "git commit -m \"upload sample files \"")
	out, err := gitCommitCmd.Output()
	if err != nil {
		fmt.Println(err)
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

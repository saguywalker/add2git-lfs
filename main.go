package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/saguywalker/add2git-lfs/gitcommand"
	"github.com/saguywalker/add2git-lfs/helper"

	rice "github.com/GeertJohan/go.rice"

	"github.com/labstack/echo"
)

var remote string
var branch string
var uploadsDir string
var token string

func main() {
	flag.StringVar(&remote, "remote", "origin", "remote")
	flag.StringVar(&branch, "branch", "master", "branch")
	flag.StringVar(&uploadsDir, "folder", "sample-files", "folder to upload")
	flag.StringVar(&token, "token", "", "personal access token (https)")
	flag.Parse()

	fmt.Println("************Debugging*************")
	fmt.Printf("%s\n%s\n%s\n%s\n", remote, branch, uploadsDir, token)

	os.MkdirAll(filepath.Join(".", uploadsDir), os.ModePerm)
	err := helper.InitLfs(branch, uploadsDir)
	if err != nil {
		panic(err)
	}

	e := echo.New()
	assetHandler := http.FileServer(rice.MustFindBox("public").HTTPBox())
	e.GET("/", echo.WrapHandler(assetHandler))
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", assetHandler)))
	e.POST("/upload", handleUpload)
	e.POST("/pushfiles", handlePushFiles)
	go helper.Open("http://localhost:12358/")
	e.Logger.Fatal(e.Start(":12358"))
}

func handleUpload(c echo.Context) error {

	c.Request().ParseMultipartForm(32 << 20)
	form, err := c.MultipartForm()
	if err != nil {
		message := fmt.Sprintf("Error when parsing files %s", err.Error())
		return c.String(http.StatusBadRequest, message)
	}
	files := form.File["file"]

	var fullname string
	for _, file := range files {

		fullname = filepath.Join(".", uploadsDir, file.Filename)

		src, err := file.Open()
		if err != nil {
			message := fmt.Sprintf("Error when opening %v", file.Filename)
			return c.String(http.StatusBadRequest, message)
		}
		defer src.Close()

		dst, err := os.Create(fullname)
		if err != nil {
			message := fmt.Sprintf("Error when opening %v", file.Filename)
			return c.String(http.StatusBadRequest, message)
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			message := fmt.Sprintf("Error when opening %v", file.Filename)
			return c.String(http.StatusBadRequest, message)
		}
	}

	return c.String(http.StatusOK, "Files are uploaded")

}

func handlePushFiles(c echo.Context) error {
	err := gitcommand.GitAddFile(uploadsDir)
	if err != nil {
		errMsg := fmt.Sprintf("Error when running git add %s\n\n***************************************************\n%s", uploadsDir, err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	err = gitcommand.GitCommitShell(uploadsDir)
	if err != nil {
		errMsg := fmt.Sprintf("Error when running git commit\n\n***************************************************\n%s", err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	if token == "" {
		err = gitcommand.GitPushShell(remote, branch)
	} else {
		err = gitcommand.GitPushToken(remote, branch, token)
	}
	if err != nil {
		errMsg := fmt.Sprintf("Error when running git push\n\n***************************************************\n%s", err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}

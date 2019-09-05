package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	rice "github.com/GeertJohan/go.rice"
	"github.com/labstack/echo"
	"github.com/saguywalker/add2git-lfs/gitcommand"
)

func main() {
	branch := flag.String("branch", "master", "branch")
	email := flag.String("email", "", "user.email for commit")
	port := flag.Int("port", 12358, "port for webapp")
	remote := flag.String("remote", "origin", "remote")
	token := flag.String("token", "", "personal access token")
	uploadsDir := flag.String("folder", "sample-files", "folder to upload files")
	user := flag.String("user", "", "user.name for commit")

	flag.Parse()

	config := gitcommand.NewConfig(*branch, *email, runtime.GOOS, *remote, *token, *uploadsDir, *user)

	if config.User != "" {
		if err := config.ConfigUser("Name"); err != nil {
			panic(err)
		}
	}

	if *email != "" {
		if err := config.ConfigUser("Email"); err != nil {
			panic(err)
		}
	}

	os.MkdirAll(filepath.Join(".", config.UploadsDir), os.ModePerm)
	err := config.InitLfs()
	if err != nil {
		panic(err)
	}

	e := echo.New()

	assetHandler := http.FileServer(rice.MustFindBox("public").HTTPBox())
	e.GET("/", echo.WrapHandler(assetHandler))
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", assetHandler)))
	e.POST("/upload", config.HandleUpload)
	e.POST("/pushfiles", config.HandlePushFiles)

	go Open(fmt.Sprintf("http://127.0.0.1:%d", *port))
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", *port)))
}

//Open a browser according to URL
func Open(url string) error {
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

package gitcommand

import (
	"errors"
	"fmt"
	"testing"
)

var cases = []struct {
	in      string
	out     string
	isHTTPS bool
	err     error
}{
	{"git@github.com:saguywalker/add2git-lfs", "github.com/saguywalker/add2git-lfs.git", false, nil},
	{"git@github.com:username/repository.git", "github.com/username/repository.git", false, nil},
	{"http://github.com/user/repo", "github.com/user/repo.git", false, nil},
	{"https://gitlab.com/CinCan/tools", "gitlab.com/CinCan/tools.git", true, nil},
	{"git", "", false, errors.New("too short url")},
}

func TestSplitURL(t *testing.T) {
	for _, c := range cases {
		url, https, err := splitGitURL([]byte(c.in))

		if err != nil && (err.Error() != c.err.Error()) {
			t.Fatal(err)
		}

		if url != c.out || https != c.isHTTPS {
			fmt.Printf("%v\n", c)
			fmt.Printf("%s, %t, %v\n", url, https, err)
			t.Fatal(c.in)
		}
	}
}

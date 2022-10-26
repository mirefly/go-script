package gitsyncer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootDir = func() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".gitsyncer")
	}
	return ""
}

func loadConfig(c *viper.Viper) (*Config, error) {
	if c == nil {
		c = viper.New()
	}

	c.SetConfigName(".gitsyncer")

	if cwd, _ := os.Getwd(); cwd != "" {
		c.AddConfigPath(cwd)
	}

	c.AddConfigPath(rootDir())

	if err := c.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := c.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func removeProtocolPrefix(url string) string {
	prefixes := []string{"https://", "http:"}
	for _, p := range prefixes {
		url = strings.TrimPrefix(url, p)
	}
	return url
}

const BRANCH_REF_NAME_PREFIX = "refs/heads/"

func sync(t *Task) error {
	folder := filepath.Join(rootDir(), removeProtocolPrefix(t.Src.URL))
	srcAuth := &http.BasicAuth{Username: t.Src.Username, Password: t.Src.Password}
	dstAuth := &http.BasicAuth{Username: t.Dst.Username, Password: t.Dst.Password}

	r, err := git.PlainOpen(folder)
	if err != nil {
		r, err = git.PlainClone(folder, true, &git.CloneOptions{
			URL:      t.Src.URL,
			Auth:     srcAuth,
			Progress: os.Stdout,
		})
		if err != nil {
			return err
		}
	}

	srcRemote, err := r.Remote("origin")
	if err != nil {
		return fmt.Errorf("couldn't get origin remote: %s", err)
	}

	r.DeleteRemote("dst")
	dstRemote, err := r.CreateRemote(&config.RemoteConfig{
		Name: "dst",
		URLs: []string{t.Dst.URL},
	})
	if err != nil {
		return fmt.Errorf("couldn't create dst remote: %s", err)
	}

	if err := srcRemote.Fetch(&git.FetchOptions{
		Auth:     srcAuth,
		Progress: os.Stdout,
	}); err != nil {
		if err == git.NoErrAlreadyUpToDate {
			log.Printf("fetch from src: %s\n", err)
		} else {
			return fmt.Errorf("couldn't fetch origin: %s", err)
		}
	}

	if err := dstRemote.Push(&git.PushOptions{
		RemoteName: "dst",
		RefSpecs:   []config.RefSpec{"refs/remotes/origin/*:refs/heads/*"},
		Auth:       dstAuth,
		Progress:   os.Stdout,
		Force:      true,
	}); err != nil {
		if err == git.NoErrAlreadyUpToDate {
			log.Printf("push to dst: %s\n", err)
		} else {
			return fmt.Errorf("couldn't push all branch dst: %s", err)
		}
	}

	return nil
}

func Run() {
	c, err := loadConfig(nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, task := range c.Tasks {
		if err := sync(&task); err != nil {
			log.Fatal(err)
		}
	}
}

var Cmd = &cobra.Command{
	Use: "gitsyncer",
}

func init() {
	Cmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Run sync once",
		Run: func(cmd *cobra.Command, args []string) {
			Run()
		},
	})
}

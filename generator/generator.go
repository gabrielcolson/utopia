package generator

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/gabrielcolson/utopia/cfg"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
)

type Generator struct {
	opts Options
	auth transport.AuthMethod
}

type Options struct {
	URL            string
	ConfigFileName string
	DestDir        string
	Verbose        bool
}

func New(opts Options) (*Generator, error) {
	var u *url.URL
	g := &Generator{
		opts: opts,
	}

	if u = parseHttpUrl(g.opts.URL); u != nil {
		g.log("Detected HTTP url")

		g.auth = nil
	} else if u = parseSshUrl(g.opts.URL); u != nil {
		g.log("Detected SSH url")

		auth, err := sshAuth()
		g.auth = auth
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid URL: %s", g.opts.URL)
	}

	if g.opts.DestDir == "" {
		g.opts.DestDir = urlBaseName(u)
	}
	if g.opts.ConfigFileName == "" {
		g.opts.ConfigFileName = ".utopia.yml"
	}

	return g, nil
}

func (g *Generator) log(v ...interface{}) {
	if g.opts.Verbose {
		log.Println(v...)
	}
}

func sshAuth() (transport.AuthMethod, error) {
	sshPath := os.Getenv("HOME") + "/.ssh/id_rsa"
	sshKey, err := ioutil.ReadFile(sshPath)
	if err != nil {
		return nil, err
	}
	return ssh.NewPublicKeys("git", sshKey, "")
}

func (g *Generator) Generate() error {
	fs := memfs.New()

	r, err := g.clone(fs)
	if err != nil {
		return err
	}
	g.log("Clone successful!")

	config, err := g.loadConfig(fs)
	if err != nil {
		return err
	}

	selectedFeatures, err := g.promptFeaturesSelection(config)
	if err != nil {
		return err
	}

	if err = g.pullSelectedFeatures(r, config, selectedFeatures); err != nil {
		return err
	}

	if err = g.writeToDisk(fs, "/", g.opts.DestDir); err != nil {
		return err
	}

	fmt.Printf("Your project was successfully created in \"%s\"\n", g.opts.DestDir)

	return nil
}

func (g *Generator) clone(fs billy.Filesystem) (*git.Repository, error) {
	storage := memory.NewStorage()

	var progressOutput io.Writer
	if g.opts.Verbose {
		progressOutput = log.Writer()
	}

	return git.Clone(storage, fs, &git.CloneOptions{
		URL:      g.opts.URL,
		Auth:     g.auth,
		Progress: progressOutput,
	})
}

func (g *Generator) loadConfig(fs billy.Filesystem) (*cfg.Config, error) {
	configFile, err := fs.Open(g.opts.ConfigFileName)
	if err != nil {
		return nil, err
	}
	return cfg.FromYaml(configFile)
}

func (g *Generator) promptFeaturesSelection(config *cfg.Config) (features []string, err error) {
	options := make([]string, 0, len(config.Features))
	for featName := range config.Features {
		options = append(options, featName)
	}

	prompt := &survey.MultiSelect{
		Message: "Which feature do you need?",
		Options: options,
		VimMode: true,
	}

	err = survey.AskOne(prompt, &features)
	return
}

func (g *Generator) pullSelectedFeatures(r *git.Repository, c *cfg.Config, selectedFeature []string) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	for _, featName := range selectedFeature {
		branchName := c.Features[featName].Branch
		err := w.Pull(&git.PullOptions{
			RemoteName:    "origin",
			ReferenceName: plumbing.NewBranchReferenceName(branchName),
			Auth:          g.auth,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) writeToDisk(fs billy.Filesystem, srcDirPath, destDirPath string) error {
	if err := os.Mkdir(destDirPath, 0777); err != nil {
		return err
	}

	files, err := fs.ReadDir(srcDirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Name() == g.opts.ConfigFileName {
			continue
		}

		srcFilePath := path.Join(srcDirPath, file.Name())
		destFilePath := path.Join(destDirPath, file.Name())

		if file.IsDir() {
			if err = g.writeToDisk(fs, srcFilePath, destFilePath); err != nil {
				return err
			}
		} else {
			currentFile, err := fs.Open(srcFilePath)
			if err != nil {
				return err
			}

			f, err := os.Create(destFilePath)
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, currentFile); err != nil {
				return err
			}
		}
	}

	return nil
}

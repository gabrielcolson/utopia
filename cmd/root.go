package cmd

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"gopkg.in/yaml.v2"
	"net/url"

	"io"
	"io/ioutil"
	"os"
	"path"
)

const configFileName string = ".utopia.yml"

var rootCmd = &cobra.Command{
	Use:   "utopia",
	Short: "Utopia is a simple, git based and language agnostic template generator",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectDirName := args[0]
		url := args[1]

		u := Utopia{
			fs: memfs.New(),
		}

		err := u.cloneTemplate(url)
		if err != nil {
			return err
		}

		config, err := u.loadConfig()
		if err != nil {
			return err
		}

		selectedFeatures, err := askFeatures(config)
		if err != nil {
			return err
		}

		err = u.pullFeatures(selectedFeatures)
		if err != nil {
			return err
		}

		err = u.writeProjectToDisk("/", projectDirName)
		if err != nil {
			return err
		}

		fmt.Printf("Your project was successfully created in \"%s\"\n", projectDirName)

		return nil
	},
}

type Feature struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Branch      string `yaml:"branch"`
}

type Config struct {
	Features []Feature `yaml:"features"`
}

type Utopia struct {
	fs   billy.Filesystem
	repo *git.Repository
}

func (u *Utopia) cloneTemplate(url string) error {
	storage := memory.NewStorage()

	var publicKey transport.AuthMethod
	if isHttpUrl(url) {
		publicKey = nil
	} else { // is ssh url
		sshPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		sshKey, err := ioutil.ReadFile(sshPath)
		if err != nil {
			return err
		}
		publicKey, err = ssh.NewPublicKeys("git", sshKey, "")
		if err != nil {
			return err
		}
	}

	r, err := git.Clone(storage, u.fs, &git.CloneOptions{
		URL:  url,
		Auth: publicKey,
	})
	if err != nil {
		return err
	}

	u.repo = r
	return nil
}

func isHttpUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func (u *Utopia) loadConfig() (*Config, error) {
	configFile, err := u.fs.Open(configFileName)
	if err != nil {
		return nil, err
	}

	configDecoder := yaml.NewDecoder(configFile)
	config := new(Config)
	if err = configDecoder.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func askFeatures(config *Config) ([]Feature, error) {
	var featOptions []string
	for _, feat := range config.Features {
		featOptions = append(featOptions, fmt.Sprintf("%s: %s", feat.Name, feat.Description))
	}

	prompt := &survey.MultiSelect{
		Message: "Which feature do you need?",
		Options: featOptions,
		VimMode: true,
	}

	var selectedFeatIndices []int
	if err := survey.AskOne(prompt, &selectedFeatIndices); err != nil {
		return nil, err
	}

	selectedFeatures := make([]Feature, len(selectedFeatIndices))
	for i, index := range selectedFeatIndices {
		selectedFeatures[i] = config.Features[index]
	}

	return selectedFeatures, nil
}

func (u *Utopia) pullFeatures(features []Feature) error {
	w, err := u.repo.Worktree()
	if err != nil {
		return err
	}

	for _, feat := range features {
		err := w.Pull(&git.PullOptions{
			RemoteName:    "origin",
			ReferenceName: plumbing.NewBranchReferenceName(feat.Branch),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *Utopia) writeProjectToDisk(sourceDirPath string, targetDirPath string) error {
	fmt.Println("source:", sourceDirPath)
	fmt.Println("target:", targetDirPath)
	if err := os.Mkdir(targetDirPath, 0777); err != nil {
		return err
	}

	files, err := u.fs.ReadDir(sourceDirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Name() == configFileName {
			continue
		}

		sourceFilePath := path.Join(sourceDirPath, file.Name())
		targetFilePath := path.Join(targetDirPath, file.Name())

		if file.IsDir() {
			if err = u.writeProjectToDisk(sourceFilePath, targetFilePath); err != nil {
				return err
			}
		} else {
			currentFile, err := u.fs.Open(sourceFilePath)
			if err != nil {
				return err
			}

			f, err := os.Create(targetFilePath)
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

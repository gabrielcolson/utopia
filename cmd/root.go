package cmd

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path"
)

type Feature struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Branch      string `yaml:"branch"`
}

type Config struct {
	Features []Feature `yaml:"features"`
}

var rootCmd = &cobra.Command{
	Use:   "utopia",
	Short: "Utopia is a simple, git based and language agnostic template generator",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fs := memfs.New()
		storage := memory.NewStorage()

		url := args[0]
		r, err := git.Clone(storage, fs, &git.CloneOptions{
			URL: url,
		})
		if err != nil {
			return err
		}

		configFile, err := fs.Open("utopia.yml")
		if err != nil {
			return err
		}

		configDecoder := yaml.NewDecoder(configFile)
		var config Config
		if err = configDecoder.Decode(&config); err != nil {
			return err
		}

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
			return err
		}

		w, err := r.Worktree()
		if err != nil {
			return err
		}

		for _, featIndex := range selectedFeatIndices {
			branchName := config.Features[featIndex].Branch
			err := w.Pull(&git.PullOptions{
				RemoteName:    "origin",
				ReferenceName: plumbing.NewBranchReferenceName(branchName),
			})
			if err != nil {
				return err
			}
		}

		projectDirName := "project"

		if err := os.Mkdir(projectDirName, 0777); err != nil {
			return err
		}

		files, err := fs.ReadDir("/")
		if err != nil {
			return err
		}

		for _, file := range files {
			fmt.Println(file.Name())
			currentFile, err := fs.Open(file.Name())
			if err != nil {
				return err
			}

			p := path.Join(projectDirName, file.Name())
			f, err := os.Create(p)
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, currentFile); err != nil {
				return err
			}
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

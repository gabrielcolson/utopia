package cmd

import (
	"github.com/gabrielcolson/utopia/generator"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// persistent flags
var (
	Verbose bool
)

// local flags
var (
	DestDir string
)

var rootCmd = &cobra.Command{
	Use:   "utopia <git url>",
	Short: "Utopia is a simple, git based and language agnostic template generator",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		g := generator.New(generator.Options{
			URL:     args[0],
			DestDir: DestDir,
			Verbose: Verbose,
		})

		return g.Generate()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&DestDir, "dest", "", "", "Destination folder")

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Verbose output")
}

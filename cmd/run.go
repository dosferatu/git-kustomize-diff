/*
Copyright 2021 Daisuke Taniwaki.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dtaniwaki/git-kustomize-diff/pkg/gitkustomizediff"
	"github.com/spf13/cobra"
)

type runFlags struct {
	base                string
	target              string
	includeRegexpString string
	excludeRegexpString string
	kustomizePath       string
	gitPath             string
	debug               bool
	allowDirty          bool
}

var runCmd = &cobra.Command{
	Use:   "run target_dir",
	Short: "Run git-kustomize-diff",
	Long:  `Run git-kustomize-diff`,
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := gitkustomizediff.RunOpts{
			Base:          runOpts.base,
			Target:        runOpts.target,
			Debug:         runOpts.debug,
			AllowDirty:    runOpts.allowDirty,
			KustomizePath: runOpts.kustomizePath,
			GitPath:       runOpts.gitPath,
		}
		if runOpts.includeRegexpString != "" {
			includeRegexp, err := regexp.Compile(runOpts.includeRegexpString)
			if err != nil {
				return err
			}
			opts.IncludeRegexp = includeRegexp
		}
		if runOpts.excludeRegexpString != "" {
			excludeRegexp, err := regexp.Compile(runOpts.excludeRegexpString)
			if err != nil {
				return err
			}
			opts.ExcludeRegexp = excludeRegexp
		}

		dir := "."
		if len(args) == 1 {
			dir = args[0]
		}
		res, err := gitkustomizediff.Run(dir, opts)
		if err != nil {
			fmt.Printf("%+v\n", err)
			return nil
		}

		printRunResult(dir, opts, res)

		return nil
	},
}

var runOpts runFlags

func init() {
	runCmd.PersistentFlags().StringVar(&runOpts.base, "base", "origin/main", "base commitish")
	runCmd.PersistentFlags().StringVar(&runOpts.target, "target", "", "target commitish (default to the current branch)")
	runCmd.PersistentFlags().StringVar(&runOpts.includeRegexpString, "include", "", "include regexp (default to all)")
	runCmd.PersistentFlags().StringVar(&runOpts.excludeRegexpString, "exclude", "", "exclude regexp (default to none)")
	runCmd.PersistentFlags().StringVar(&runOpts.kustomizePath, "kustomize-path", "", "path of a kustomize binary (default to embeded)")
	runCmd.PersistentFlags().StringVar(&runOpts.gitPath, "git-path", "", "path of a git binary (default to git)")
	runCmd.PersistentFlags().BoolVar(&runOpts.debug, "debug", false, "debug mode")
	runCmd.PersistentFlags().BoolVar(&runOpts.allowDirty, "allow-dirty", false, "allow dirty tree")
}

func printRunResult(dirPath string, opts gitkustomizediff.RunOpts, res *gitkustomizediff.RunResult) {
	dirs := res.DiffMap.Dirs()
	fmt.Printf("# Git Kustomize Diff\n\n")

	fmt.Printf("%s...%s\n\n", res.BaseCommit, res.TargetCommit)

	fmt.Printf("<details><summary>Options</summary>\n\n")
	fmt.Println("| name | value |")
	fmt.Println("|-|-|")
	fmt.Printf("| dir | %s |\n", dirPath)
	fmt.Printf("| base | %s |\n", opts.Base)
	fmt.Printf("| target | %s |\n", opts.Target)
	fmt.Printf("\n</details>\n\n")

	fmt.Printf("<details><summary>Target Kustomizations</summary>\n\n")
	if len(dirs) > 0 {
		fmt.Printf("```\n%s\n```\n", strings.Join(dirs, "\n"))
	} else {
		fmt.Println("N/A")
	}
	fmt.Printf("\n</details>\n\n")

	found := false
	for _, dir := range dirs {
		text := res.DiffMap.Results[dir].AsMarkdown()
		if text != "" {
			fmt.Printf("## %s\n\n", dir)
			fmt.Printf("<details><summary>diff</summary>\n\n")
			fmt.Println(text)
			fmt.Printf("\n</details>\n\n")
			found = true
		}
	}
	if !found {
		fmt.Println("N/A")
	}
}

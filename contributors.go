// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"sort"

	"github.com/go-yaml/yaml"
	"github.com/google/go-github/github"
	"github.com/urfave/cli"
)

var cmdContributors = cli.Command{
	Name:        "contributors",
	Usage:       "generate contributors list of the milestone",
	Description: `generate contributors list of the milestone`,
	Action:      runContributors,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "milestone, m",
			Usage: "Generate which tag from",
		},
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Specify a config file",
		},
	},
}

func runContributors(cmd *cli.Context) {
	milestone := cmd.String("milestone")
	if milestone == "" {
		fmt.Println("Please specify a milestone")
		return
	}

	var err error
	var configContent []byte
	if cmd.String("config") == "" {
		configContent = defaultConfig
	} else {
		configContent, err = ioutil.ReadFile(cmd.String("config"))
		if err != nil {
			fmt.Printf("Load config from file %s failed: %v\n", cmd.String("config"), err)
			return
		}
	}

	var config Config
	err = yaml.Unmarshal(configContent, &config)
	if err != nil {
		fmt.Printf("Unmarshal config content failed: %v\n", err)
		return
	}

	client := github.NewClient(nil)
	ctx := context.Background()

	var contributorsMap = make(map[string]bool)
	var query = fmt.Sprintf(`repo:%s is:merged milestone:"%s"`, config.Repo, milestone)
	var p = 1
	var perPage = 100
	for {
		result, _, err := client.Search.Issues(ctx, query, &github.SearchOptions{
			ListOptions: github.ListOptions{
				Page:    p,
				PerPage: perPage,
			},
		})
		p++
		if err != nil {
			log.Fatal(err.Error())
		}

		for _, pr := range result.Issues {
			contributorsMap[*pr.User.Login] = true
		}

		if len(result.Issues) != perPage {
			break
		}
	}

	contributors := make([]string, 0, len(contributorsMap))
	for contributor, _ := range contributorsMap {
		contributors = append(contributors, contributor)
	}

	sort.Strings(contributors)

	for _, contributor := range contributors {
		fmt.Printf("* [@%s](https://github.com/%s)\n", contributor, contributor)
	}
}

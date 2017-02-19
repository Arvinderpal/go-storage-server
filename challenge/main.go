// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package main

import (
	"fmt"
	"os"

	"github.com/Arvinderpal/go-storage-server/challenge/common"
	daemon "github.com/Arvinderpal/go-storage-server/challenge/daemon/daemon"
	s "github.com/Arvinderpal/go-storage-server/challenge/daemon/server"

	"github.com/codegangsta/cli"
	l "github.com/op/go-logging"
)

var (
	config        = daemon.NewConfig()
	log           = l.MustGetLogger("challenge")
	socketAddress string
)

func main() {
	app := cli.NewApp()
	app.Name = "challenge"
	app.Usage = "Challenge"
	app.Version = common.Version
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "Enable debug messages",
		},
		cli.StringFlag{
			Destination: &config.DataDirBasePath,
			Name:        "dir",
			Value:       common.DataDirBasePath,
			Usage:       "data directory",
		},
		cli.StringFlag{
			Destination: &socketAddress,
			Name:        "s",
			Value:       common.ServerSockAddress,
			Usage:       "Sets the socket address to listen for connections",
		},
	}

	app.Before = initEnv
	app.Run(os.Args)
}

func initEnv(ctx *cli.Context) error {
	fmt.Printf("Initializing...")
	if ctx.Bool("debug") {
		common.SetupLOG(log, "DEBUG")
	} else {
		common.SetupLOG(log, "INFO")
	}
	fmt.Printf("done!\n")
	return nil
}

func run(cli *cli.Context) {

	fmt.Printf("Starting storage-server...")

	d, err := daemon.NewDaemon(config)
	if err != nil {
		log.Fatalf("Error while creating daemon: %s", err)
		return
	}

	server, err := s.NewServer(socketAddress, d)
	if err != nil {
		log.Fatalf("Error while creating server: %s", err)
	}
	defer server.Stop()
	server.Start()
}

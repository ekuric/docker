package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/docker/docker/api/types"
	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/timeutils"
)

var validDrivers = map[string]bool{
	"json-file": true,
	"journald":  true,
}

// CmdLogs fetches the logs of a given container.
//
// docker logs [OPTIONS] CONTAINER
func (cli *DockerCli) CmdLogs(args ...string) error {
	cmd := Cli.Subcmd("logs", []string{"CONTAINER"}, Cli.DockerCommands["logs"].Description, true)
	follow := cmd.Bool([]string{"f", "-follow"}, false, "Follow log output")
	since := cmd.String([]string{"-since"}, "", "Show logs since timestamp")
	times := cmd.Bool([]string{"t", "-timestamps"}, false, "Show timestamps")
	tail := cmd.String([]string{"-tail"}, "all", "Number of lines to show from the end of the logs")
	cmd.Require(flag.Exact, 1)

	cmd.ParseFlags(args, true)

	name := cmd.Arg(0)

	serverResp, err := cli.call("GET", "/containers/"+name+"/json", nil, nil)
	if err != nil {
		return err
	}

	var c types.ContainerJSON
	if err := json.NewDecoder(serverResp.body).Decode(&c); err != nil {
		return err
	}

	if !validDrivers[c.HostConfig.LogConfig.Type] {
		return fmt.Errorf("\"logs\" command is supported only for \"json-file\" and \"journald\" logging drivers (got: %s)", c.HostConfig.LogConfig.Type)
	}

	v := url.Values{}
	v.Set("stdout", "1")
	v.Set("stderr", "1")

	if *since != "" {
		ts, err := timeutils.GetTimestamp(*since, time.Now())
		if err != nil {
			return err
		}
		v.Set("since", ts)
	}

	if *times {
		v.Set("timestamps", "1")
	}

	if *follow {
		v.Set("follow", "1")
	}
	v.Set("tail", *tail)

	sopts := &streamOpts{
		rawTerminal: c.Config.Tty,
		out:         cli.out,
		err:         cli.err,
	}

	_, err = cli.stream("GET", "/containers/"+name+"/logs?"+v.Encode(), sopts)
	return err
}

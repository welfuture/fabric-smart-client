/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package view

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/cli"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/client/view"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/grpc"
)

var (
	// Function used to terminate the CLI
	terminate = os.Exit
	// Function used to redirect output to
	outWriter io.Writer = os.Stderr
)

// CommandRegistrar registers commands
type CommandRegistrar interface {
	// Command adds a new top-level command to the CLI
	Command(name, help string, onCommand cli.CLICommand) *kingpin.CmdClause
}

func RegisterViewCommand(cli CommandRegistrar) {
	ic := &invokeCmd{}
	cmd := cli.Command("view", "Invoke a view", ic.invoke)

	ic.endpoint = cmd.Flag("endpoint", "Sets the endpoint of the node to connect to (host:port)").String()
	ic.input = cmd.Flag("input", "Sets the input to the view function, encoded either as base64, or as-is").String()
	ic.function = cmd.Flag("function", "Sets the function name to be invoked").String()
	ic.stdin = cmd.Flag("stdin", "Sets standard input as the input stream").Bool()
}

type invokeCmd struct {
	endpoint *string
	input    *string
	function *string
	stdin    *bool
	rawInput []byte
}

func (ic *invokeCmd) validateInput() error {
	if ic.endpoint == nil {
		return fmt.Errorf("endpoint must be specified")
	}

	if ic.function == nil {
		return fmt.Errorf("function name must be specified")
	}

	// Check if input is to be read from stdin
	if ic.stdin != nil && *ic.stdin {
		stdinInput, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed reading input from stdin: %v", err)
		}
		ic.rawInput = stdinInput
		return nil
	}

	if ic.input == nil {
		// If input isn't specified, the input is nil
		return nil
	}

	// Check if it's a base64 encoded string
	rawBase64Encoded, err := base64.StdEncoding.DecodeString(*ic.input)
	if err == nil {
		ic.rawInput = rawBase64Encoded
		return nil
	}

	ic.rawInput = []byte(*ic.input)

	return nil
}

func (ic *invokeCmd) GetHash() hash.Hash {
	return sha256.New()
}

func (ic *invokeCmd) Hash(msg []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(msg)
	return h.Sum(nil), nil
}

func (ic *invokeCmd) invoke(config cli.Config) error {
	if err := ic.validateInput(); err != nil {
		return err
	}

	cc := &grpc.ConnectionConfig{
		Address:           *ic.endpoint,
		TLSEnabled:        true,
		TLSRootCertFile:   path.Join(config.TLSConfig.PeerCACertPath),
		ConnectionTimeout: 10 * time.Second,
	}

	signer, err := view.NewX509SigningIdentity(config.SignerConfig.IdentityPath, config.SignerConfig.KeyPath)
	if err != nil {
		return err
	}

	c, err := view.NewClient(
		&view.Config{
			ConnectionConfig: cc,
		},
		signer,
		ic,
	)
	if err != nil {
		return err
	}

	res, err := c.CallView(*ic.function, ic.rawInput)
	if err != nil {
		return err
	}

	switch v := res.(type) {
	case []byte:
		fmt.Println(string(v))
	default:
		fmt.Println(v)
	}

	return nil
}

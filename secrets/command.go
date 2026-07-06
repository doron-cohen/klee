package secrets

import (
	"context"
	"fmt"
	"os"

	"github.com/doron-cohen/klee/config"
	"github.com/urfave/cli/v3"
)

// Command returns a CLI command for managing secrets via the given store.
func Command(store config.SecretStore) *cli.Command {
	return &cli.Command{
		Name:  "secrets",
		Usage: "manage secrets",
		Commands: []*cli.Command{
			{
				Name:  "get",
				Usage: "retrieve a secret value",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					key := cmd.Args().First()
					if key == "" {
						return fmt.Errorf("usage: secrets get <key>")
					}
					val, err := store.Get(key)
					if err != nil {
						return err
					}
					_, _ = fmt.Fprintln(os.Stdout, val)
					return nil
				},
			},
			{
				Name:  "set",
				Usage: "store a secret value",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					ws, ok := store.(config.WritableSecretStore)
					if !ok {
						return fmt.Errorf("secrets set: configured store (%T) does not support write operations", store)
					}
					args := cmd.Args().Slice()
					if len(args) != 2 {
						return fmt.Errorf("usage: secrets set <key> <value>")
					}
					return ws.Set(args[0], args[1])
				},
			},
		},
	}
}

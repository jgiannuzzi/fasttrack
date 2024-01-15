package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/rotisserie/eris"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
)

var DecodeCmd = &cobra.Command{
	Use:    "decode",
	Short:  "Decodes a binary Aim stream",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		f := os.Stdin
		if len(args) > 0 {
			var err error
			f, err = os.Open(args[0])
			if err != nil {
				return err
			}
			//nolint:errcheck
			defer f.Close()
		}

		decoder := encoding.NewDecoder(f)
		for {
			data, err := decoder.Next()
			if len(data) > 0 {
				keys := maps.Keys(data)
				slices.Sort(keys)
				for _, key := range keys {
					fmt.Printf("%s: %#v\n", key, data[key])
				}
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return eris.Wrap(err, "error decoding binary AIM stream")
			}
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(DecodeCmd)
}

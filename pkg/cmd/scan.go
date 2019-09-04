package cmd

import (
	"context"

	"github.com/int128/migerr/pkg/usecases/scan"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

type Scan struct {
	UseCase scan.UseCase
}

func (s *Scan) New(ctx context.Context) *cobra.Command {
	c := &cobra.Command{
		Use:   "scan PACKAGE...",
		Short: "Scan usage of error handling",
		RunE: func(_ *cobra.Command, args []string) error {
			if err := s.UseCase.Do(ctx, args...); err != nil {
				return xerrors.Errorf("could not scan the packages: %w", err)
			}
			return nil
		},
	}
	return c
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/steipete/gogcli/internal/outfmt"
)

var (
	version      = "0.12.0-dev"
	commit       = ""
	date         = ""
	distribution = "buildr"
)

func VersionString() string {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "dev"
	}
	parts := []string{distribution}
	if c := strings.TrimSpace(commit); c != "" {
		parts = append(parts, c)
	}
	if d := strings.TrimSpace(date); d != "" {
		parts = append(parts, d)
	}
	return fmt.Sprintf("%s (%s)", v, strings.Join(parts, " "))
}

type VersionCmd struct{}

func (c *VersionCmd) Run(ctx context.Context) error {
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{
			"version":      strings.TrimSpace(version),
			"commit":       strings.TrimSpace(commit),
			"date":         strings.TrimSpace(date),
			"distribution": distribution,
		})
	}
	fmt.Fprintln(os.Stdout, VersionString())
	return nil
}

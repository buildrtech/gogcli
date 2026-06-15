package cmd

import (
	"context"
	_ "embed"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/steipete/gogcli/internal/outfmt"
)

//go:embed VERSION
var embeddedVersion string

const sentinelDev = "dev"

var (
	version       = sentinelDev
	commit        = ""
	date          = ""
	distribution  = "buildr"
	readBuildInfo = debug.ReadBuildInfo
)

func resolvedVersion() string {
	v := strings.TrimSpace(version)
	if v != "" && v != sentinelDev {
		return v
	}
	info, ok := readBuildInfo()
	if ok {
		moduleVersion := strings.TrimSpace(info.Main.Version)
		if moduleVersion != "" && moduleVersion != "(devel)" {
			return moduleVersion
		}
	}
	if baked := strings.TrimSpace(embeddedVersion); baked != "" {
		return baked
	}
	return sentinelDev
}

func VersionString() string {
	v := resolvedVersion()
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
	stdout := stdoutWriter(ctx)
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, stdout, map[string]any{
			"version":      resolvedVersion(),
			"commit":       strings.TrimSpace(commit),
			"date":         strings.TrimSpace(date),
			"distribution": distribution,
		})
	}
	fmt.Fprintln(stdout, VersionString())
	return nil
}

package checkpoint

import (
	"github.com/docker/cli/cli/command/formatter"
	"github.com/moby/moby/api/types/checkpoint"
)

const (
	defaultCheckpointFormat = "table {{.Name}}"
	checkpointNameHeader    = "CHECKPOINT NAME"
)

// NewFormat returns a format for use with a checkpoint Context
func NewFormat(source string) formatter.Format {
	if source == formatter.TableFormatKey {
		return defaultCheckpointFormat
	}
	return formatter.Format(source)
}

// FormatWrite writes formatted checkpoints using the Context
func FormatWrite(ctx formatter.Context, checkpoints []checkpoint.Summary) error {
	render := func(format func(subContext formatter.SubContext) error) error {
		for _, cp := range checkpoints {
			if err := format(&checkpointContext{c: cp}); err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(newCheckpointContext(), render)
}

type checkpointContext struct {
	formatter.HeaderContext
	c checkpoint.Summary
}

func newCheckpointContext() *checkpointContext {
	cpCtx := checkpointContext{}
	cpCtx.Header = formatter.SubHeaderContext{
		"Name": checkpointNameHeader,
	}
	return &cpCtx
}

func (c *checkpointContext) MarshalJSON() ([]byte, error) {
	return formatter.MarshalJSON(c)
}

func (c *checkpointContext) Name() string {
	return c.c.Name
}

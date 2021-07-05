package prettylogzap

import (
	"errors"
	"net/url"
	"os"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	parsedLine struct {
		Timestamp string
		Logger    string
		Caller    string
		Level     string
		Message   string
		Fields    [][]string
	}

	colorPadding struct {
		color   *color.Color
		padding int
	}

	settings struct {
		level     map[string]colorPadding
		timestamp colorPadding
		logger    colorPadding
		caller    colorPadding
		message   colorPadding
	}
)

var (
	defaultColor = color.New(color.FgWhite)

	prettySettings = settings{
		level:     map[string]colorPadding{},
		timestamp: colorPadding{color.New(color.FgYellow), 0},
		logger:    colorPadding{color.New(color.FgWhite), 10},
		caller:    colorPadding{color.New(color.FgWhite), 20},
		message:   colorPadding{color.New(color.FgWhite, color.Bold), 30},
	}

	colors = map[string]*color.Color{
		"debug": color.New(color.FgMagenta),
		"info":  color.New(color.FgBlue),
		"warn":  color.New(color.FgYellow),
		"error": color.New(color.FgRed),
		"fatal": color.New(color.FgRed, color.Bold),
	}

	ErrNonParseableLine = errors.New("line could not be parsed")
	ErrInvalidColor     = errors.New("invalid color")
	ErrInvalidName      = errors.New("invalid name")
)

func (s *settings) parseLevel(levelName string) colorPadding {
	if cp, found := s.level[levelName]; found {
		return cp
	}

	if c, exists := colors[levelName]; exists {
		s.level[levelName] = colorPadding{c, 5}
	} else {
		s.level[levelName] = colorPadding{defaultColor, 5}
	}

	return s.level[levelName]
}

func SetColorPadding(name string, c *color.Color, padding int) error {
	if c == nil {
		return ErrInvalidColor
	}

	cp := colorPadding{c, padding}
	switch name {
	case "timestamp":
		prettySettings.timestamp = cp
	case "logger":
		prettySettings.logger = cp
	case "caller":
		prettySettings.caller = cp
	case "message":
		prettySettings.message = cp
	case "debug":
		prettySettings.level["debug"] = cp
	case "info":
		prettySettings.level["info"] = cp
	case "warn":
		prettySettings.level["warn"] = cp
	case "error":
		prettySettings.level["error"] = cp
	default:
		return ErrInvalidName
	}

	return nil
}

func NewPrettySink(encoderConfig zapcore.EncoderConfig) func(u *url.URL) (zap.Sink, error) {
	factory := func(u *url.URL) (zap.Sink, error) {
		switch u.Host {
		case "stdout":
			return prettySink{Sink: os.Stdout, encoderConfig: encoderConfig}, nil
		case "stderr":
			return prettySink{Sink: os.Stderr, encoderConfig: encoderConfig}, nil
		}
		return os.OpenFile(u.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	}
	return factory
}

package prettylogzap

import (
	"bytes"
	"strings"

	"github.com/fatih/color"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	Separator      = " "
	FieldQuotes    = "\""
	FieldSeparator = "="
)

type prettySink struct {
	zap.Sink
	encoderConfig zapcore.EncoderConfig
}

func (w prettySink) Write(p []byte) (int, error) {
	line, err := w.parse(p)
	if err != nil {
		return w.Sink.Write(p)
	}
	data := w.Prettify(line)
	return w.Sink.Write(data)
}

func (w prettySink) Close() error {
	return nil
}

func (w prettySink) parse(line []byte) (*parsedLine, error) {
	if !gjson.ValidBytes(line) {
		return nil, ErrNonParseableLine
	}

	parsed := &parsedLine{}
	gjson.ParseBytes(line).ForEach(func(key, value gjson.Result) bool {
		switch key.String() {
		case w.encoderConfig.TimeKey:
			parsed.Timestamp = value.String()
		case w.encoderConfig.NameKey:
			parsed.Logger = value.String()
		case w.encoderConfig.CallerKey:
			parsed.Caller = value.String()
		case w.encoderConfig.LevelKey:
			parsed.Level = value.String()
		case w.encoderConfig.MessageKey:
			parsed.Message = value.String()
		default:
			parsed.Fields = append(parsed.Fields, []string{key.String(), value.String()})
		}
		return true
	})

	return parsed, nil
}

func (w prettySink) Prettify(line *parsedLine) []byte {
	buffer := &bytes.Buffer{}

	if line.Timestamp != "" {
		w.writeTo(buffer, line.Timestamp, prettySettings.timestamp.padding, prettySettings.timestamp.color)
	}

	if line.Logger != "" {
		w.writeTo(buffer, line.Logger, prettySettings.logger.padding, prettySettings.logger.color)
	}

	if line.Caller != "" {
		w.writeTo(buffer, line.Caller, prettySettings.caller.padding, prettySettings.caller.color)
	}

	cpLevel := prettySettings.parseLevel(strings.ToLower(line.Level))
	if line.Level != "" {
		w.writeTo(buffer, strings.ToUpper(line.Level), cpLevel.padding, cpLevel.color)
	}

	w.writeTo(buffer, line.Message, prettySettings.message.padding, prettySettings.message.color)
	w.writeFieldsTo(buffer, line.Fields, cpLevel.color)

	buffer.WriteString("\n")
	return buffer.Bytes()
}

func (w prettySink) writeTo(buffer *bytes.Buffer, value string, padding int, color *color.Color) {
	value = w.padRight(value, padding)
	value = color.Sprint(value)
	buffer.WriteString(value)
	buffer.WriteString(Separator)
}

func (w prettySink) writeFieldsTo(buffer *bytes.Buffer, fields [][]string, color *color.Color) {
	for _, field := range fields {
		buffer.WriteString(color.Sprint(field[0]))
		buffer.WriteString(FieldSeparator)
		if strings.Contains(field[1], " ") {
			buffer.WriteString(FieldQuotes)
			buffer.WriteString(field[1])
			buffer.WriteString(FieldQuotes)
		} else {
			buffer.WriteString(field[1])
		}
		buffer.WriteString(Separator)
	}
}

func (w prettySink) padRight(str string, size int) string {
	size -= len(str)
	if size < 0 {
		size = 0
	}
	return str + strings.Repeat(" ", size)
}

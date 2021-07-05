package prettylogzap

import (
	"strings"

	"github.com/fatih/color"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const (
	separator      = ' '
	fieldSeparator = '='
	fieldQuotes    = "\""
)

var pool = buffer.NewPool()

type prettySink struct {
	zap.Sink
	encoderConfig zapcore.EncoderConfig
}

func (w prettySink) Write(p []byte) (int, error) {
	line, err := w.parse(p)
	if err != nil {
		return w.Sink.Write(p)
	}
	return w.Sink.Write(w.prettify(line))
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

func (w prettySink) prettify(line *parsedLine) []byte {
	buf := pool.Get()

	if line.Timestamp != "" {
		w.writeTo(buf, line.Timestamp, prettySettings.timestamp.padding, prettySettings.timestamp.color)
	}

	if line.Logger != "" {
		w.writeTo(buf, line.Logger, prettySettings.logger.padding, prettySettings.logger.color)
	}

	if line.Caller != "" {
		w.writeTo(buf, line.Caller, prettySettings.caller.padding, prettySettings.caller.color)
	}

	cpLevel := prettySettings.parseLevel(strings.ToLower(line.Level))
	if line.Level != "" {
		w.writeTo(buf, strings.ToUpper(line.Level), cpLevel.padding, cpLevel.color)
	}

	w.writeTo(buf, line.Message, prettySettings.message.padding, prettySettings.message.color)
	w.writeFieldsTo(buf, line.Fields, cpLevel.color)

	buf.AppendString("\n")
	data := buf.Bytes()
	buf.Free()
	return data
}

func (w prettySink) writeTo(buf *buffer.Buffer, value string, padding int, color *color.Color) {
	value = w.padRight(value, padding)
	value = color.Sprint(value)
	buf.AppendString(value)
	buf.AppendByte(separator)
}

func (w prettySink) writeFieldsTo(buf *buffer.Buffer, fields [][]string, color *color.Color) {
	for _, field := range fields {
		buf.AppendString(color.Sprint(field[0]))
		buf.AppendByte(fieldSeparator)
		if strings.Contains(field[1], " ") {
			buf.AppendString(fieldQuotes)
			buf.AppendString(field[1])
			buf.AppendString(fieldQuotes)
		} else {
			buf.AppendString(field[1])
		}
		buf.AppendByte(separator)
	}
}

func (w prettySink) padRight(str string, size int) string {
	size -= len(str)
	if size <= 0 {
		return str
	}
	return str + strings.Repeat(" ", size)
}

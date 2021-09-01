# Pretty Log Zap

Formata a exibição de logs estruturados em JSON em um formato compatível com seres humanos.


![Prettylog](https://github.com/globocom/prettylogzap/raw/master/prettylogzap.png)


## Utilização
```go
import (
	"fmt"

	"github.com/globocom/prettylogzap"
	"go.uber.org/zap"
)


func NewLoggerDevelopment() (*zap.Logger, error) {

	// create zap config
	zapConf = zap.NewDevelopmentConfig()
	zapConf.DisableCaller = true

	// set encoding json
	zapConf.Encoding = "json"

	// set output path:
	// pretty://stderr
	// pretty://stdout
	zapConf.OutputPaths = []string{"pretty://stdout"}

	// register pretty sink
	prettySink := prettylogzap.NewPrettySink(zapConf.EncoderConfig)
	if err := zap.RegisterSink("pretty", prettySink); err != nil {
		return nil, fmt.Errorf("register prettysink error: %w", err)
	}

	logger, err := zapConf.Build()
	if err != nil {
		return nil, fmt.Errorf("fail on define zap as logger: %w", err)
	}

	return logger, nil
}

```

## Agradecimento
Esse projeto foi inspirado no projeto [prettylog](https://github.com/globocom/prettylog). Obrigado!

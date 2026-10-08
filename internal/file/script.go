package file

import (
	"strings"
	"unicode"

	"al.essio.dev/pkg/shellescape"
	"github.com/gofiber/fiber/v2"
)

// ScriptCommand accepts an argument list, never an additional shell program.
func ScriptCommand(path, parameters string, windows bool) (string, error) {
	arguments, err := scriptArguments(parameters, windows)
	if err != nil {
		return "", err
	}
	quote := shellescape.Quote
	prefix := ""
	if windows {
		prefix = "& "
		quote = func(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }
	}
	parts := []string{quote(path)}
	for _, argument := range arguments {
		parts = append(parts, quote(argument))
	}
	return prefix + strings.Join(parts, " "), nil
}

func scriptArguments(parameters string, windows bool) ([]string, error) {
	arguments := []string{}
	var value strings.Builder
	var quote rune
	started, escaped := false, false
	runes := []rune(parameters)
	for i := 0; i < len(runes); i++ {
		current := runes[i]
		if current == 0 {
			return nil, fiber.NewError(fiber.StatusBadRequest, "invalid script parameters")
		}
		if escaped {
			value.WriteRune(current)
			escaped = false
			continue
		}
		escape := rune('\\')
		if windows {
			escape = '`'
		}
		if current == escape && quote != '\'' {
			escaped = true
			started = true
			continue
		}
		if quote != 0 {
			if current == quote {
				if windows && quote == '\'' && i+1 < len(runes) && runes[i+1] == '\'' {
					value.WriteRune(current)
					i++
					continue
				}
				quote = 0
			} else {
				value.WriteRune(current)
			}
			continue
		}
		if current == '\'' || current == '"' {
			quote = current
			started = true
			continue
		}
		if unicode.IsSpace(current) {
			if started {
				arguments = append(arguments, value.String())
				value.Reset()
				started = false
			}
			continue
		}
		started = true
		value.WriteRune(current)
	}
	if quote != 0 || escaped {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid script parameters")
	}
	if started {
		arguments = append(arguments, value.String())
	}
	return arguments, nil
}

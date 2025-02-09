package runtime

import (
	"encoding/json"
	"errors"
	"strings"
)

func ScriptContentInsideStringLiteral[T any](v T, errs ...error) (string, error) {
	return scriptContent(v, true, errs...)
}

func ScriptContentOutsideStringLiteral[T any](v T, errs ...error) (string, error) {
	return scriptContent(v, false, errs...)
}

func scriptContent[T any](v T, insideStringLiteral bool, errs ...error) (string, error) {
	if errors.Join(errs...) != nil {
		return "", errors.Join(errs...)
	}
	jd, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if insideStringLiteral {
		return strings.Trim(string(jd), "\""), nil
	}
	return string(jd), nil
}

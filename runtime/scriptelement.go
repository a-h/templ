package runtime

import "errors"

// JoinErrs joins an optional list of errors.
func JoinErrs[T any](v T, errs ...error) (T, error) {
	return v, errors.Join(errs...)
}

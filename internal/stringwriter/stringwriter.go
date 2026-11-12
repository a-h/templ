package stringwriter

import "io"

// WriteStrings writes each string to w in turn, so that callers do not
// allocate a concatenated string.
func WriteStrings(w io.Writer, ss ...string) (err error) {
	for _, s := range ss {
		if _, err = io.WriteString(w, s); err != nil {
			return err
		}
	}
	return nil
}

// SPDX-FileCopyrightText: 2019 The Go Language Server Authors
// SPDX-License-Identifier: BSD-3-Clause

package protocol

import (
	"fmt"
	"strconv"
	"testing"

	"encoding/json"
	"github.com/google/go-cmp/cmp"
)

func TestShowMessageParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"message":"error message","type":1}`
		wantUnknown = `{"message":"unknown message","type":0}`
	)
	wantType := ShowMessageParams{
		Message: "error message",
		Type:    MessageTypeError,
	}
	wantTypeUnkonwn := ShowMessageParams{
		Message: "unknown message",
		Type:    MessageType(0),
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          ShowMessageParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Unknown",
				field:          wantTypeUnkonwn,
				want:           wantUnknown,
				wantMarshalErr: false,
				wantErr:        false,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name             string
			field            string
			want             ShowMessageParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Unknown",
				field:            wantUnknown,
				want:             wantTypeUnkonwn,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got ShowMessageParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestShowMessageRequestParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"actions":[{"title":"Retry"}],"message":"error message","type":1}`
		wantUnknown = `{"actions":[{"title":"Retry"}],"message":"unknown message","type":0}`
	)
	wantType := ShowMessageRequestParams{
		Actions: []MessageActionItem{
			{
				Title: "Retry",
			},
		},
		Message: "error message",
		Type:    MessageTypeError,
	}
	wantTypeUnkonwn := ShowMessageRequestParams{
		Actions: []MessageActionItem{
			{
				Title: "Retry",
			},
		},
		Message: "unknown message",
		Type:    MessageType(0),
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          ShowMessageRequestParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Unknown",
				field:          wantTypeUnkonwn,
				want:           wantUnknown,
				wantMarshalErr: false,
				wantErr:        false,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name             string
			field            string
			want             ShowMessageRequestParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Unknown",
				field:            wantUnknown,
				want:             wantTypeUnkonwn,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got ShowMessageRequestParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestMessageActionItem(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"title":"Retry"}`
		wantOpenLog = `{"title":"Open Log"}`
	)
	wantType := MessageActionItem{
		Title: "Retry",
	}
	wantTypeOpenLog := MessageActionItem{
		Title: "Open Log",
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          MessageActionItem
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Unknown",
				field:          wantTypeOpenLog,
				want:           wantOpenLog,
				wantMarshalErr: false,
				wantErr:        false,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name             string
			field            string
			want             MessageActionItem
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Unknown",
				field:            wantOpenLog,
				want:             wantTypeOpenLog,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got MessageActionItem
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestLogMessageParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"message":"error message","type":1}`
		wantUnknown = `{"message":"unknown message","type":0}`
	)
	wantType := LogMessageParams{
		Message: "error message",
		Type:    MessageTypeError,
	}
	wantTypeUnknown := LogMessageParams{
		Message: "unknown message",
		Type:    MessageType(0),
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          LogMessageParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Unknown",
				field:          wantTypeUnknown,
				want:           wantUnknown,
				wantMarshalErr: false,
				wantErr:        false,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name             string
			field            string
			want             LogMessageParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Unknown",
				field:            wantUnknown,
				want:             wantTypeUnknown,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got LogMessageParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestWorkDoneProgressCreateParams(t *testing.T) {
	t.Parallel()

	const (
		wantToken    = int32(1569)
		invalidToken = int32(1348)
	)
	var (
		wantString        = `{"token":"` + strconv.FormatInt(int64(wantToken), 10) + `"}`
		wantInvalidString = `{"token":"` + strconv.FormatInt(int64(invalidToken), 10) + `"}`
		wantNumber        = `{"token":` + strconv.FormatInt(int64(wantToken), 10) + `}`
		wantInvalidNumber = `{"token":` + strconv.FormatInt(int64(invalidToken), 10) + `}`
	)
	token := NewProgressToken(strconv.FormatInt(int64(wantToken), 10))
	wantTypeString := WorkDoneProgressCreateParams{
		Token: *token,
	}
	numberToken := NewNumberProgressToken(wantToken)
	wantTypeNumber := WorkDoneProgressCreateParams{
		Token: *numberToken,
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          WorkDoneProgressCreateParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid/String",
				field:          wantTypeString,
				want:           wantString,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Valid/Number",
				field:          wantTypeNumber,
				want:           wantNumber,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid/String",
				field:          wantTypeString,
				want:           wantInvalidString,
				wantMarshalErr: false,
				wantErr:        true,
			},
			{
				name:           "Invalid/Number",
				field:          wantTypeNumber,
				want:           wantInvalidNumber,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             WorkDoneProgressCreateParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid/String",
				field:            wantString,
				want:             wantTypeString,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Valid/Number",
				field:            wantNumber,
				want:             wantTypeNumber,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid/String",
				field:            wantInvalidString,
				want:             wantTypeString,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
			{
				name:             "Invalid/Number",
				field:            wantInvalidNumber,
				want:             wantTypeNumber,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got WorkDoneProgressCreateParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(fmt.Sprint(got.Token), strconv.FormatInt(int64(wantToken), 10)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestWorkDoneProgressCancelParams(t *testing.T) {
	t.Parallel()

	const (
		wantToken    = int32(1569)
		invalidToken = int32(1348)
	)
	var (
		want        = `{"token":` + strconv.FormatInt(int64(wantToken), 10) + `}`
		wantInvalid = `{"token":` + strconv.FormatInt(int64(invalidToken), 10) + `}`
	)
	token := NewNumberProgressToken(wantToken)
	wantType := WorkDoneProgressCancelParams{
		Token: *token,
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          WorkDoneProgressCancelParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             WorkDoneProgressCancelParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got WorkDoneProgressCancelParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(fmt.Sprint(got.Token), strconv.FormatInt(int64(wantToken), 10)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestMessageType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		m    MessageType
		want string
	}{
		{
			name: "Error",
			m:    MessageTypeError,
			want: "error",
		},
		{
			name: "Warning",
			m:    MessageTypeWarning,
			want: "warning",
		},
		{
			name: "Info",
			m:    MessageTypeInfo,
			want: "info",
		},
		{
			name: "Log",
			m:    MessageTypeLog,
			want: "log",
		},
		{
			name: "Unknown",
			m:    MessageType(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.m.String(); got != tt.want {
				t.Errorf("MessageType.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestMessageType_Enabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		m     MessageType
		level MessageType
		want  bool
	}{
		{
			name:  "ErrorError",
			m:     MessageTypeError,
			level: MessageTypeError,
			want:  true,
		},
		{
			name:  "ErrorInfo",
			m:     MessageTypeError,
			level: MessageTypeInfo,
			want:  false,
		},
		{
			name:  "ErrorUnknown",
			m:     MessageTypeError,
			level: MessageType(0),
			want:  false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.m.Enabled(tt.level); got != tt.want {
				t.Errorf("MessageType.Enabled(%v) = %v, want %v", tt.level, tt.want, got)
			}
		})
	}
}

func TestToMessageType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level string
		want  MessageType
	}{
		{
			name:  "Error",
			level: "error",
			want:  MessageTypeError,
		},
		{
			name:  "Warning",
			level: "warning",
			want:  MessageTypeWarning,
		},
		{
			name:  "Info",
			level: "info",
			want:  MessageTypeInfo,
		},
		{
			name:  "Log",
			level: "log",
			want:  MessageTypeLog,
		},
		{
			name:  "Unknown",
			level: "0",
			want:  MessageType(0),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ToMessageType(tt.level); got != tt.want {
				t.Errorf("ToMessageType(%v) = %v, want %v", tt.level, tt.want, got)
			}
		})
	}
}

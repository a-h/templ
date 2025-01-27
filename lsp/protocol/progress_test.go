// SPDX-FileCopyrightText: 2021 The Go Language Server Authors
// SPDX-License-Identifier: BSD-3-Clause

package protocol

import (
	"fmt"
	"testing"

	"encoding/json"
	"github.com/google/go-cmp/cmp"
)

func TestWorkDoneProgressBegin(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"kind":"begin","title":"testTitle","cancellable":true,"message":"testMessage","percentage":30}`
		wantNil     = `{"kind":"begin","title":"testTitle"}`
		wantInvalid = `{"kind":"invalid","title":"invalidTitle","cancellable":false,"message":"invalidMessage","percentage":0}`
	)
	wantType := WorkDoneProgressBegin{
		Kind:        WorkDoneProgressKindBegin,
		Title:       "testTitle",
		Cancellable: true,
		Message:     "testMessage",
		Percentage:  uint32(30),
	}
	wantTypeNil := WorkDoneProgressBegin{
		Kind:  WorkDoneProgressKindBegin,
		Title: "testTitle",
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          WorkDoneProgressBegin
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
				name:           "Nil",
				field:          wantTypeNil,
				want:           wantNil,
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
			want             WorkDoneProgressBegin
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
				name:             "Nil",
				field:            wantNil,
				want:             wantTypeNil,
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

				var got WorkDoneProgressBegin
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

func TestWorkDoneProgressReport(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"kind":"report","cancellable":true,"message":"testMessage","percentage":30}`
		wantNil     = `{"kind":"report"}`
		wantInvalid = `{"kind":"invalid","cancellable":false,"message":"invalidMessage","percentage":0}`
	)
	wantType := WorkDoneProgressReport{
		Kind:        WorkDoneProgressKindReport,
		Cancellable: true,
		Message:     "testMessage",
		Percentage:  uint32(30),
	}
	wantTypeNil := WorkDoneProgressReport{
		Kind: WorkDoneProgressKindReport,
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          WorkDoneProgressReport
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
				name:           "Nil",
				field:          wantTypeNil,
				want:           wantNil,
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
			want             WorkDoneProgressReport
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
				name:             "Nil",
				field:            wantNil,
				want:             wantTypeNil,
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

				var got WorkDoneProgressReport
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

func TestWorkDoneProgressEnd(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"kind":"end","message":"testMessage"}`
		wantNil     = `{"kind":"end"}`
		wantInvalid = `{"kind":"invalid","message":"invalidMessage"}`
	)
	wantType := WorkDoneProgressEnd{
		Kind:    WorkDoneProgressKindEnd,
		Message: "testMessage",
	}
	wantTypeNil := WorkDoneProgressEnd{
		Kind: WorkDoneProgressKindEnd,
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          WorkDoneProgressEnd
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
				name:           "Nil",
				field:          wantTypeNil,
				want:           wantNil,
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
			want             WorkDoneProgressEnd
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
				name:             "Nil",
				field:            wantNil,
				want:             wantTypeNil,
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

				var got WorkDoneProgressEnd
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

func TestWorkDoneProgressParams(t *testing.T) {
	t.Parallel()

	const wantWorkDoneToken = "156edea9-9d8d-422f-b7ee-81a84594afbb"
	const want = `{"workDoneToken":"` + wantWorkDoneToken + `"}`

	wantType := WorkDoneProgressParams{
		WorkDoneToken: NewProgressToken(wantWorkDoneToken),
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          WorkDoneProgressParams
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
			want             WorkDoneProgressParams
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
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got WorkDoneProgressParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if workDoneToken := got.WorkDoneToken; workDoneToken != nil {
					if diff := cmp.Diff(fmt.Sprint(workDoneToken), wantWorkDoneToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestPartialResultParams(t *testing.T) {
	t.Parallel()

	const wantPartialResultParams = "156edea9-9d8d-422f-b7ee-81a84594afbb"
	const want = `{"partialResultToken":"` + wantPartialResultParams + `"}`

	wantType := PartialResultParams{
		PartialResultToken: NewProgressToken(wantPartialResultParams),
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          PartialResultParams
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
			want             PartialResultParams
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
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got PartialResultParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if partialResultToken := got.PartialResultToken; partialResultToken != nil {
					if diff := cmp.Diff(fmt.Sprint(partialResultToken), wantPartialResultParams); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

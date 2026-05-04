package utils

import (
	"testing"
	"time"
)

func TestStringToDateRFC1123(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantZero  bool
		wantYear  int
		wantMonth time.Month
		wantDay   int
	}{
		{
			name:      "valid RFC1123 date",
			input:     "Fri, 18 Oct 2019 07:40:06 GMT",
			wantYear:  2019,
			wantMonth: time.October,
			wantDay:   18,
		},
		{
			name:     "empty string returns zero time",
			input:    "",
			wantZero: true,
		},
		{
			name:     "malformed string returns zero time",
			input:    "not-a-date",
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringToDateRFC1123(tt.input)
			if tt.wantZero {
				if !got.IsZero() {
					t.Errorf("expected zero time for input %q, got %v", tt.input, got)
				}
				return
			}
			if got.Year() != tt.wantYear {
				t.Errorf("year: expected %d, got %d", tt.wantYear, got.Year())
			}
			if got.Month() != tt.wantMonth {
				t.Errorf("month: expected %v, got %v", tt.wantMonth, got.Month())
			}
			if got.Day() != tt.wantDay {
				t.Errorf("day: expected %d, got %d", tt.wantDay, got.Day())
			}
		})
	}
}

func TestDateToSeconds_roundtrip(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if DateToSeconds(t0) != t0.Unix() {
		t.Errorf("expected %d, got %d", t0.Unix(), DateToSeconds(t0))
	}
}

func TestSecondsToTime_roundtrip(t *testing.T) {
	t0 := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	got := SecondsToTime(DateToSeconds(t0))
	if !got.Equal(t0) {
		t.Errorf("expected %v, got %v", t0, got)
	}
}

func TestTimeToMicroSecond_roundtrip(t *testing.T) {
	t0 := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	micros := TimeToMicroSecond(t0)
	got := time.Unix(0, micros*int64(time.Microsecond))
	if !got.Equal(t0) {
		t.Errorf("expected %v, got %v", t0, got)
	}
}

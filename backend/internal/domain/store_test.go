package domain

import (
	"testing"
	"time"
)

func TestStoreSettingsIsOpenAt(t *testing.T) {
	location := time.FixedZone("UTC-3", -3*60*60)
	sundayEvening := time.Date(2026, 11, 22, 22, 30, 0, 0, time.UTC) // 19:30 local
	sundayMorning := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)  // 09:00 local
	mondayEvening := time.Date(2026, 11, 23, 22, 30, 0, 0, time.UTC) // 19:30 local

	tests := []struct {
		name     string
		store    StoreSettings
		now      time.Time
		wantOpen bool
	}{
		{
			name:     "kill switch desligado",
			store:    StoreSettings{IsOpen: false},
			now:      sundayEvening,
			wantOpen: false,
		},
		{
			name:     "sem horários cadastrados usa apenas o kill switch",
			store:    StoreSettings{IsOpen: true},
			now:      sundayEvening,
			wantOpen: true,
		},
		{
			name: "dentro do intervalo",
			store: StoreSettings{
				IsOpen: true,
				OpeningHours: []OpeningHours{{
					Day:       "sunday",
					Intervals: []OpeningInterval{{Open: "18:00", Close: "23:30"}},
				}},
			},
			now:      sundayEvening,
			wantOpen: true,
		},
		{
			name: "fora do intervalo",
			store: StoreSettings{
				IsOpen: true,
				OpeningHours: []OpeningHours{{
					Day:       "sunday",
					Intervals: []OpeningInterval{{Open: "18:00", Close: "23:30"}},
				}},
			},
			now:      sundayMorning,
			wantOpen: false,
		},
		{
			name: "dia marcado como fechado",
			store: StoreSettings{
				IsOpen: true,
				OpeningHours: []OpeningHours{{
					Day:       "sunday",
					Closed:    true,
					Intervals: []OpeningInterval{{Open: "18:00", Close: "23:30"}},
				}},
			},
			now:      sundayEvening,
			wantOpen: false,
		},
		{
			name: "dia sem configuração",
			store: StoreSettings{
				IsOpen: true,
				OpeningHours: []OpeningHours{{
					Day:       "sunday",
					Intervals: []OpeningInterval{{Open: "18:00", Close: "23:30"}},
				}},
			},
			now:      mondayEvening,
			wantOpen: false,
		},
		{
			name: "virada de meia-noite",
			store: StoreSettings{
				IsOpen: true,
				OpeningHours: []OpeningHours{{
					Day:       "sunday",
					Intervals: []OpeningInterval{{Open: "23:00", Close: "02:00"}},
				}},
			},
			now:      time.Date(2026, 11, 23, 2, 30, 0, 0, time.UTC), // 23:30 local de domingo
			wantOpen: true,
		},
		{
			name: "intervalo inválido não abre a loja",
			store: StoreSettings{
				IsOpen: true,
				OpeningHours: []OpeningHours{{
					Day:       "sunday",
					Intervals: []OpeningInterval{{Open: "18:00", Close: "18:00"}},
				}},
			},
			now:      sundayEvening,
			wantOpen: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.store.IsOpenAt(test.now, location); got != test.wantOpen {
				t.Errorf("IsOpenAt() = %t, want %t", got, test.wantOpen)
			}
		})
	}
}

func TestStoreSettingsNextOpeningAt(t *testing.T) {
	location := time.FixedZone("UTC-3", -3*60*60)
	// 19:30 de domingo no fuso local.
	now := time.Date(2026, 11, 22, 22, 30, 0, 0, time.UTC)
	hours := []OpeningHours{
		{Day: "sunday", Intervals: []OpeningInterval{{Open: "18:00", Close: "23:30"}}},
		{Day: "monday", Intervals: []OpeningInterval{{Open: "18:00", Close: "23:30"}}},
		{Day: "tuesday", Closed: true},
	}

	tests := []struct {
		name  string
		store StoreSettings
		now   time.Time
		want  string
	}{
		{
			name:  "kill switch desligado não tem previsão",
			store: StoreSettings{IsOpen: false, OpeningHours: hours},
			now:   now,
		},
		{
			name:  "sem horários não tem previsão",
			store: StoreSettings{IsOpen: true},
			now:   now,
		},
		{
			name:  "próxima abertura é amanhã quando o horário de hoje passou",
			store: StoreSettings{IsOpen: true, OpeningHours: hours[:2]},
			now:   now,
			want:  "2026-11-23T21:00:00Z",
		},
		{
			name:  "próxima abertura é hoje quando o horário ainda não chegou",
			store: StoreSettings{IsOpen: true, OpeningHours: hours[:2]},
			now:   time.Date(2026, 11, 22, 18, 0, 0, 0, time.UTC), // 15:00 local
			want:  "2026-11-22T21:00:00Z",
		},
		{
			name:  "pula dias fechados",
			store: StoreSettings{IsOpen: true, OpeningHours: hours[1:]},
			now:   now,
			want:  "2026-11-23T21:00:00Z",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.store.NextOpeningAt(test.now, location)
			if test.want == "" {
				if got != nil {
					t.Fatalf("NextOpeningAt() = %v, want nil", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("NextOpeningAt() = nil, want %s", test.want)
			}
			if formatted := got.Format(time.RFC3339); formatted != test.want {
				t.Errorf("NextOpeningAt() = %s, want %s", formatted, test.want)
			}
		})
	}
}

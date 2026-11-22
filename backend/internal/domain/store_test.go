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

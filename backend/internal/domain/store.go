package domain

import (
	"strings"
	"time"
)

// StoreSettingsID é o identificador fixo do documento único de configuração.
const StoreSettingsID = "main"

// StoreSettings é a configuração operacional da loja. O campo IsOpen é o kill
// switch manual; OpeningHours complementa a decisão quando há horários cadastrados.
type StoreSettings struct {
	Name         string
	IsOpen       bool
	OpeningHours []OpeningHours
	DeliveryTime string
	WhatsApp     string
	Instagram    string
	Address      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// OpeningHours descreve um dia da semana. Dia sem configuração é considerado
// fechado; close menor que open indica virada de meia-noite.
type OpeningHours struct {
	Day       string
	Closed    bool
	Intervals []OpeningInterval
}

// OpeningInterval é uma faixa de horário no fuso da operação.
type OpeningInterval struct {
	Open  string
	Close string
}

// IsOpenAt aplica o kill switch e, quando existem horários cadastrados, a faixa
// vigente no fuso informado. Sem horários cadastrados, o kill switch decide:
// o seed de desenvolvimento ainda não contém openingHours e os horários reais
// dependem da validação da US-025.
func (s StoreSettings) IsOpenAt(now time.Time, location *time.Location) bool {
	if !s.IsOpen {
		return false
	}
	if len(s.OpeningHours) == 0 {
		return true
	}
	if location == nil {
		location = time.UTC
	}

	local := now.In(location)
	weekday := strings.ToLower(local.Weekday().String())
	minutes := local.Hour()*60 + local.Minute()

	for _, day := range s.OpeningHours {
		if !strings.EqualFold(strings.TrimSpace(day.Day), weekday) {
			continue
		}
		if day.Closed {
			continue
		}
		for _, interval := range day.Intervals {
			if intervalContains(interval, minutes) {
				return true
			}
		}
	}

	return false
}

func intervalContains(interval OpeningInterval, minutes int) bool {
	open, err := parseClock(interval.Open)
	if err != nil {
		return false
	}
	close, err := parseClock(interval.Close)
	if err != nil || close == open {
		return false
	}

	if close < open {
		return minutes >= open || minutes < close
	}
	return minutes >= open && minutes < close
}

func parseClock(value string) (int, error) {
	parsed, err := time.Parse("15:04", strings.TrimSpace(value))
	if err != nil {
		return 0, err
	}
	return parsed.Hour()*60 + parsed.Minute(), nil
}

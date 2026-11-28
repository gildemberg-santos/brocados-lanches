package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Limites de entrada da configuração da loja. Não são dados de negócio.
const (
	MaxDeliveryTimeLength = 60
	MaxInstagramLength    = 200
	MaxAddressLength      = 300
	MaxIntervalsPerDay    = 10
)

var whatsappPattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

var weekdays = map[string]struct{}{
	"sunday": {}, "monday": {}, "tuesday": {}, "wednesday": {},
	"thursday": {}, "friday": {}, "saturday": {},
}

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
	Webhook      WebhookSettings
	Pix          PixSettings
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

// Validate confere as invariantes da configuração da loja antes de persistir.
// Horários inválidos ou sobrepostos poderiam deixar a loja aceitar pedidos fora
// do expediente; o kill switch sozinho não é suficiente.
func (s StoreSettings) Validate() error {
	if err := validateRequiredString("name", s.Name, MaxNameLength); err != nil {
		return err
	}
	if err := validateRequiredString("deliveryTime", s.DeliveryTime, MaxDeliveryTimeLength); err != nil {
		return err
	}
	if err := validateWhatsApp("whatsapp", s.WhatsApp); err != nil {
		return err
	}
	if strings.TrimSpace(s.Instagram) != "" {
		if err := validateMaxLength("instagram", s.Instagram, MaxInstagramLength); err != nil {
			return err
		}
	}
	if err := validateRequiredString("address", s.Address, MaxAddressLength); err != nil {
		return err
	}
	if err := s.Webhook.Validate(); err != nil {
		return err
	}
	if err := s.Pix.Validate(); err != nil {
		return err
	}
	return validateOpeningHours(s.OpeningHours)
}

func validateWhatsApp(field, value string) error {
	if err := validateRequiredString(field, value, 16); err != nil {
		return err
	}
	if !whatsappPattern.MatchString(value) {
		return invalid(field, CodeInvalid, "must be in E.164 format, e.g. +5585999999999")
	}
	return nil
}

func validateOpeningHours(days []OpeningHours) error {
	if len(days) > 7 {
		return invalid("openingHours", CodeOutOfRange, "must have at most 7 days")
	}

	seenDays := make(map[string]struct{}, len(days))
	for index, day := range days {
		field := fmt.Sprintf("openingHours[%d]", index)
		normalized := strings.ToLower(strings.TrimSpace(day.Day))
		if _, valid := weekdays[normalized]; !valid {
			return invalid(field+".day", CodeInvalid, "must be a lowercase weekday")
		}
		if _, duplicate := seenDays[normalized]; duplicate {
			return invalid(field+".day", CodeDuplicate, "duplicates another weekday")
		}
		seenDays[normalized] = struct{}{}

		if day.Closed {
			continue
		}
		if len(day.Intervals) == 0 {
			return invalid(field+".intervals", CodeRequired, "at least one interval is required when the day is open")
		}
		if len(day.Intervals) > MaxIntervalsPerDay {
			return invalid(field+".intervals", CodeOutOfRange, fmt.Sprintf("must have at most %d intervals", MaxIntervalsPerDay))
		}

		occupied := make([]bool, 24*60)
		for intervalIndex, interval := range day.Intervals {
			intervalField := fmt.Sprintf("%s.intervals[%d]", field, intervalIndex)
			open, err := parseClock(interval.Open)
			if err != nil {
				return invalid(intervalField+".open", CodeInvalid, "must use HH:MM")
			}
			close, err := parseClock(interval.Close)
			if err != nil {
				return invalid(intervalField+".close", CodeInvalid, "must use HH:MM")
			}
			if open == close {
				return invalid(intervalField+".close", CodeInvalid, "must be different from open")
			}
			for minute := open; minute != close; minute = (minute + 1) % (24 * 60) {
				if occupied[minute] {
					return invalid(intervalField, CodeInvalid, "overlaps another interval")
				}
				occupied[minute] = true
			}
		}
	}
	return nil
}

// IsOpenAt aplica o kill switch e, quando existem horários cadastrados, a faixa
// vigente no fuso informado. Sem horários cadastrados, o kill switch decide:
// o seed de desenvolvimento traz horários copiados do status publicado em
// /contato como referência, sujeitos à validação da loja.
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

// NextOpeningAt devolve o próximo instante em UTC em que a loja abrirá,
// considerando o kill switch e os horários no fuso informado. Sem horários
// cadastrados ou com o kill switch desligado não há previsão: o operador é quem
// controla a abertura.
func (s StoreSettings) NextOpeningAt(now time.Time, location *time.Location) *time.Time {
	if !s.IsOpen || len(s.OpeningHours) == 0 {
		return nil
	}
	if location == nil {
		location = time.UTC
	}

	local := now.In(location)
	currentDay := strings.ToLower(local.Weekday().String())
	currentMinutes := local.Hour()*60 + local.Minute()

	for offset := 0; offset <= 7; offset++ {
		dayDate := local.AddDate(0, 0, offset)
		dayName := strings.ToLower(dayDate.Weekday().String())
		for _, day := range s.OpeningHours {
			if day.Closed || !strings.EqualFold(strings.TrimSpace(day.Day), dayName) {
				continue
			}
			for _, interval := range day.Intervals {
				open, err := parseClock(interval.Open)
				if err != nil {
					continue
				}
				if offset == 0 && dayName == currentDay && open <= currentMinutes {
					continue
				}
				candidate := time.Date(
					dayDate.Year(), dayDate.Month(), dayDate.Day(),
					open/60, open%60, 0, 0, location,
				).UTC()
				return &candidate
			}
		}
	}
	return nil
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

package solis

type BinaryData struct {
	_                     [31]byte
	Temperature           uint16
	VDC1                  uint16
	VDC2                  uint16
	_                     [2]byte
	ADC1                  uint16
	ADC2                  uint16
	_                     [2]byte
	AAC                   uint16
	_                     [4]byte
	VAC                   uint16
	_                     [4]byte
	Frequency             uint16
	GenerationNow         uint16
	_                     [6]byte
	GeneratedYesterday    uint16
	GeneratedToday        uint16
	_                     [2]byte
	GeneratedTotal        uint16
	Message               [8]byte
	_                     [4]byte
	GeneratedCurrentMonth uint16
	_                     [2]byte
	GeneratedLastMonth    uint16
}

type InverterData struct {
	Message               string  `json:"message"`
	Temperature           float64 `json:"temp"`
	VDC1                  float64 `json:"vdc1"`
	VDC2                  float64 `json:"vdc2"`
	ADC1                  float64 `json:"adc1"`
	ADC2                  float64 `json:"adc2"`
	AAC                   float64 `json:"aac"`
	VAC                   float64 `json:"vac"`
	Frequency             float64 `json:"freq"`
	GenerationNow         uint64  `json:"wattsnow"`
	GeneratedYesterday    float64 `json:"gen_yesterday"`
	GeneratedToday        float64 `json:"gen_today"`
	GeneratedTotal        float64 `json:"gen_total"`
	GeneratedCurrentMonth uint64  `json:"gen_current_month"`
	GeneratedLastMonth    uint64  `json:"gen_last_month"`
}

func (id *InverterData) ParseBinaryData(data *BinaryData) {
	var message []byte
	copy(message, data.Message[:])
	id.Message = string(message)
	id.Temperature = float64(data.Temperature) / 10
	id.VDC1 = float64(data.VDC1) / 10
	id.VDC2 = float64(data.VDC2) / 10
	id.ADC1 = float64(data.ADC1) / 10
	id.ADC2 = float64(data.ADC2) / 10
	id.AAC = float64(data.AAC) / 10
	id.VAC = float64(data.VAC) / 10
	id.Frequency = float64(data.Frequency) / 100
	id.GenerationNow = uint64(data.GenerationNow)
	id.GeneratedYesterday = float64(data.GeneratedYesterday) / 100
	id.GeneratedToday = float64(data.GeneratedToday) / 100
	id.GeneratedTotal = float64(data.GeneratedTotal) / 10
	id.GeneratedCurrentMonth = uint64(data.GeneratedCurrentMonth)
	id.GeneratedLastMonth = uint64(data.GeneratedLastMonth)
}

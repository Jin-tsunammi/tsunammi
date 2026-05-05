package helius

type priorityFeeRequest struct {
	Jsonrpc string   `json:"jsonrpc"`
	ID      string   `json:"id"`
	Method  string   `json:"method"`
	Params  []Params `json:"params"`
}

type Options struct {
	// PriorityLevel string `json:"priorityLevel"`
	IncludeAllPriorityFeeLevels bool `json:"includeAllPriorityFeeLevels"`
}

type Params struct {
	AccountKeys []string `json:"accountKeys"`
	Options     Options  `json:"options"`
}

type Result[T any] struct {
	Result T `json:"result"`
	Error  *struct {
		Code    int32  `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type PriorityFeeEstimate struct {
	Levels PriorityFeeLevels `json:"priorityFeeLevels"`
}

type PriorityFeeLevels struct {
	Min       float64 `json:"min"`
	Low       float64 `json:"low"`
	Medium    float64 `json:"medium"`
	High      float64 `json:"high"`
	VeryHigh  float64 `json:"veryHigh"`
	UnsafeMax float64 `json:"unsafeMax"`
}

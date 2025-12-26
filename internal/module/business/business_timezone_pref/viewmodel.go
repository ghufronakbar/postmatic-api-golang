package business_timezone_pref

type BusinessTimezonePrefResponse struct {
	RootBusinessId string `json:"rootBusinessId"`
	Timezone       string `json:"timezone"`
	Offset         string `json:"offset"`
	Label          string `json:"label"`
}

package business_timezone_pref

type UpsertBusinessTimezonePrefInput struct {
	Timezone string `json:"timezone" validate:"required"`
}

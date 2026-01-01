package business_timezone_pref

type UpsertBusinessTimezonePrefInput struct {
	BusinessRootID int64  `json:"businessRootId" validate:"required"`
	Timezone       string `json:"timezone" validate:"required"`
}

package postpro

var (
	TIME_FORMAT                   string  = "2006-01-02 15:04:05,999"
	NUMBER_FIRST_MEASURES_REMOVED int     = 5
	TEMPERATURE_THRESHOLD_FACTOR  float64 = 0.35
	TEMPERATURE_THRESHOLD_MINIMUM float64 = 780
	GRADIENT_LIMIT_FACTOR         float64 = 3
	WIDTH_MINIMUM                 int64   = 2
)

var DATABASE_PATH string = "C:/Users/VERBRUTH/Files/Post_Pro_Tr2/go_train2/database/processed.db"

var DATABASE CalculationsDatabase = CalculationsDatabase{}

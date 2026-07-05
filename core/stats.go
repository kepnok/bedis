package core

var (
	KeyStats map[string]int

	KEY_METRIC = "Keys"
)

func UpdateDBStat(metric string, value int) {
	KeyStats[metric] = value
}

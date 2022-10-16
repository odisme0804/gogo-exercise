package daomock

//go:generate mockgen -destination=mock.go -package=$GOPACKAGE gogo-exercise/pkg/dao TaskDAO

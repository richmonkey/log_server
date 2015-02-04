all:log

log:log.go IndexLogger.go DailyLogger.go reload.go
	go build log.go IndexLogger.go DailyLogger.go reload.go

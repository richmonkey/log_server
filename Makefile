all:log

log:log.go IndexLogger.go DailyLogger.go
	go build log.go IndexLogger.go DailyLogger.go

build:
	go build -o ./bin/liquidation-bot cmd/main.go

run:
	go run cmd/main.go

clean:
	rm -f ./bin/liquidation-bot

build-pm2:
	make build && pm2 stop bot
	pm2 flush && pm2 delete bot

init-pm2:
	PRIVATE_KEY=$$PRIVATE_KEY pm2 start ./bin/liquidation-bot --name "bot" --log-date-format "YYYY-MM-DD HH:mm:ss" --merge-logs

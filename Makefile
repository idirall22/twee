gen:
	protoc -I proto/ proto/*.proto --go_out=plugins=grpc:pb/

up:
	docker-compose up -d

down:
	docker-compose down

test-auth:
	go clean -testcache
	cd auth/ && go test -v ./...	

test-timeline:
	go clean -testcache
	cd timeline/ && go test -v ./...	

test-follow:
	go clean -testcache
	cd follow/ && go test -v ./...	

test-user:
	go clean -testcache
	cd user/ && go test -v ./...	

test-tweet:
	go clean -testcache
	cd tweet/ && go test -v ./...	

test-notification:
	go clean -testcache
	cd notification/ && go test -v ./...	

client:
	go run cmd/main.go
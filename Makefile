gen:
	protoc -I proto/ proto/*.proto --go_out=plugins=grpc:pb/

up:
	docker-compose up -d

down:
	docker-compose down -d

test-auth:
	cd auth/ && go test -v ./...	

test-tweet:
	cd tweet/ && go test -v ./...	
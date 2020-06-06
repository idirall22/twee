gen:
	protoc -I proto/ proto/*.proto --go_out=plugins=grpc:pb/

up:
	docker-compose up -d

down:
	docker-compose down

test-auth:
	cd auth/ && go test -v ./...	

test-timeline:
	cd timeline/ && go test -v ./...	

test-follow:
	cd follow/ && go test -v ./...	

test-user:
	cd user/ && go test -v ./...	

test-tweet:
	cd tweet/ && go test -v ./...	
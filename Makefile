all: build build-frontend

build:
	go build -o wuerfler

build-frontend:
	rm -rf frontend-build
	rm -rf frontend/public/build
	(cd frontend && yarn build)
	cp -a frontend/public frontend-build

	
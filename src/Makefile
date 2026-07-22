.PHONY: setup build-posts run clean

setup:
	./setup_posts_services.sh

build-posts:
	./build_posts_services.sh

run:
	docker-compose up

clean:
	rm -rf bin/*
	docker-compose down -v

all: setup build-posts run 
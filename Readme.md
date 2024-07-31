docker build:

docker build -t my-go-api .

run docker:

docker run --network=ms_postgres_network --name=my_go_api \
  -p 8080:8080 \
  -e DB_HOST=url \
  -e DB_PORT=port \
  -e DB_NAME=name \
  -e DB_USERNAME=username \
  -e DB_PASSWORD=pasword \
  -e PORT=:8080 \
  -e ENV=prod \
  my-go-api
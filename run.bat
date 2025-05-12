docker build -t go-backend .
docker stop bfc-backend
docker rm bfc-backend
docker run -d -p 8080:8080 --name bfc-backend go-backend
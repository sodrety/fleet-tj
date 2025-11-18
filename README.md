# Fleet Management System

## How to Run
```bash
docker compose up --build

# Mock data
go run scripts/mock_publisher.go -vehicle B1234XYZ

# API
curl localhost:8080/vehicles/B1234XYZ/location

* Visualizer
http://localhost:8081

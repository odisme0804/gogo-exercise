# Coding Exercise

Implement a Restful task list API as well as run this application in container.

- Spec
  - Fields of task:
      - name
          - Type: String
      - status
          - Type: Bool
          - Value
              - 0=Incomplete
              - 1=Complete
  - Reponse headers
      - Content-Type=application/json
  - Unit Test
  - Manage codebase on Github

- Runtime Environment Requirement
    - Go 1.17.8+
    - Gin 1.7.7+
    - Docker

### 1.  GET /api/tasks (list tasks)
```
{
    "result": [
        {"id": 1, "name": "name", "status": 0}
    ]
}
```

### 2.  POST /api/tasks  (create task)
```
request
{
  "name": "買晚餐"
}

response status code 201
{
    "result": {"name": "買晚餐", "status": 0, "id": 1}
}
```

### 3. PUT /api/tasks/{id} (update task)
```
request
{
  "name": "買早餐",
  "status": 1
  "id": 1
}

response status code 200
{
  "result":{
    "name": "買早餐",
    "status": 1,
    "id": 1
  }
}
```

### 4. DELETE /api/tasks/{id} (delete task)
```
response status code 204, no response body
```

# Project Structure

The project structure is defined as the following:
```
├─ cmd           - main applications for this project
│  └─ app        - the task server app    
├─ internal      - private application and library code
│  └─ mock       - mock files of interfaces
└─pkg            - public application and library code
   ├─ dao        - data-related logic
   └─ server     - business logic controllers
```


# Development
The default port is 8080

## local run
```
make dev
```
## test
```
make test
```
## build binary
```
make build
```
## build and run the binary
```
make run
```
## build docker image
```
make docker-build
```
## run with docker-compose
```
docker-compose up -d
```

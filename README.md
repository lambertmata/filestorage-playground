### File storage API for testing purposes
#### Usage
```
// Default port set to 8080
go run app/cmd/main.go

// With custom port
ENV=9090 go run app/cmd/main.go
```

#### Create object

```
URL: PUT /objects
Content-Type: Form-data
Body:
{
    name: The file name
    file: The file payloyad
}
```
#### List objects 
```
URL: GET /objects
```
#### Get objects
```
URL: GET /objects/{id}
```
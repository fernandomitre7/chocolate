# To-Dos:

- Add email template for email confirmation!
- Add Middleware on routes to add special checks, like on /users one is used to check the email
- Use new Go's error with fmt.Printf(%w) to be able to use errors.Is, errors.As, etc

# Ideas:

- Refactor the service thinking in a `Service` object

PROS ?
CONS ?

```go
type Service struct {
    server  *http.Server
    db      *db.Instance
    conf    *config.ServiceConfig
    logger  *log.Logger
}

service := NewService(db, conf, log)
service.Init()
service.Start()
service.Shutdown()
service.
```
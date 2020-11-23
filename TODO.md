# To-Dos:


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
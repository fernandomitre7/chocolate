# chocolate
This is my base RESTful API with user authentication, to be used to start any personal cloud project

It includes user registration (creation), auth (login) and management (CRUD).
You could also use it as a starting point to create an `Users` microservice or as the start point on building any monolithic application (or modular monolithic whatever is in fad right now).


# Run Locally:

## 1. Install Go Dependencies:

Remember to set the GOPATH env var to the local location of this repo directory, i.e.: `$HOME/Code/chocolate`

```
go get github.com/gorilla/mux
go get github.com/lib/pq
go get github.com/dgrijalva/jwt-go
go get github.com/satori/go.uuid
go get golang.org/x/crypto/bcrypt
```

## 2. Run DB:
```        
docker run --name chocolate-db -p 5432:5432 \
    -v chocolate-db-data:/var/lib/postgresql/data \
    -e POSTGRES_USER=chocolate -e POSTGRES_PASSWORD=chocolate \
    -e POSTGRES_DB=chocolate-db -d postgres
```

To stop the DB
```
docker container stop chocolate-db
docker container rm chocolate-db
```

To connect to DB and run queries:
```
docker run -it --rm --link chocolate-db:postgres postgres psql -h chocolate-db -U chocolate -d chocolate-db
```

# 3. Configuration

All the service expected configuration should live under the `~/../chocolate/config/` directory.


* JWT:

Directory `config/` needs to have the `.priv` and `.pub` RSA keys for JWT. Here is how you can create them:

* Global runtime configurations:
```json
{
    "env": "dev", // type of environment: dev, prod, test, etc
    "debug": true, // This will enable DEBG log level, TODO: use different log levels not only debug
    "test": true, // flag to check if we are testing system, currently only used to not send the confirmation email
}
```


```
openssl genrsa -f4 -out jwt_key.priv 4096
openssl rsa -in jwt_key.priv -outform PEM -pubout -out jwt_key.pub
```






## 3. Start local service:
    
 `$ ./bin/start.sh`
    


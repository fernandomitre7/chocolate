# chocolate
This is my base RESTful API with user authentication, to be used to start any personal cloud project

It includes user registration (creation), auth (login) and management (CRUD).
You could also use it as a starting point to create an `Users` microservice or as the start point on building any monolithic application (or modular monolithic whatever is in fad right now).


Remember to set the GOPATH env var to the local location of this repo directory, i.e.: `$HOME/Code/chocolate`

go get github.com/gorilla/mux
go get github.com/lib/pq
go get github.com/dgrijalva/jwt-go
go get github.com/satori/go.uuid
go get golang.org/x/crypto/bcrypt
# Captain's Log
- generate basic server-side webapp in Go with production-grade infrastucture so it's ready to deploy on production and local development. use postgresql as database. this website is a basic todo app with responsive design to run on mobile.
- there are errors in the code. fix them.
- i can't start the server with `make dev`. this is the error from docker.
Step 9/10 : RUN go install github.com/pressly/goose/v3/cmd/goose@latest
 ---> Running in e617f21d8688
go: downloading github.com/pressly/goose v2.7.0+incompatible
go: downloading github.com/pressly/goose/v3 v3.27.1
go: github.com/pressly/goose/v3/cmd/goose@latest: github.com/pressly/goose/v3@v3.27.1 requires go >= 1.25.7 (running go 1.21.13; GOTOOLCHAIN=local)

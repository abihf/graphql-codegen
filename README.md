# graphql-codegen
Code Generator for Initializing [graphql-go](https://github.com/graph-gophers/graphql-go) server

## Install
```sh
go get -u github.com/abihf/graphql-codegen
```
## Usage
```sh
graphql-codegen -dir out/dir -package resolve "source/schema.gql"
```
It will generate resolvers inside `out/dir`.

### Tips
Use [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) or [goreturns](https://github.com/sqs/goreturns) to automatically remove unnecessary import and format the generated codes.

## License
[MIT](LICENSE)

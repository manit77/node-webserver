module build

go 1.19

replace goutils => ../goutils

require goutils v0.0.0

require (
	github.com/golang-jwt/jwt/v4 v4.4.3 // indirect
	github.com/jinzhu/copier v0.1.0 // indirect
	golang.org/x/crypto v0.13.0 // indirect
)

module build

go 1.19

replace goutils => ../go-utils

require (
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	goutils v0.0.0-00010101000000-000000000000
)

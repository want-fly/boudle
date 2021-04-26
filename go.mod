module boudle

go 1.13

require (
	github.com/bwmarrin/snowflake v0.3.0
	github.com/gin-gonic/gin v1.7.1
	github.com/imdario/mergo v0.3.12 // indirect
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20210305010621-2afb4311ab10 // indirect
)

replace k8s.io/client-go v11.0.0+incompatible => k8s.io/client-go v0.21.0

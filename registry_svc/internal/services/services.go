package services

// GrpcServices Wrap all services needed for the grpc handlers
type Services struct {
	*JWTokenService
}

func NewServices() Services {
	svc := Services{
		JWTokenService: NewJWTokenService(),
	}
	return svc
}

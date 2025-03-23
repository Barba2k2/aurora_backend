package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Barba2k2/aurora_backend/src/controllers"
	"github.com/Barba2k2/aurora_backend/src/middlewares"
	"github.com/Barba2k2/aurora_backend/src/repositories"
	"github.com/Barba2k2/aurora_backend/src/services"
	"github.com/Barba2k2/aurora_backend/src/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
)

// getEnv obtem uma variavel de ambiente ou retorna um valor padrão
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt obtem uma variavel de ambiente como int ou retorna um valor padrão
func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// setupDatabase configura a conexão com o banco de dados
func setupDatabase() (*gorm.DB, error) {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "aurora")

	dbURI := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := gorm.Open("postgres", dbURI)
	if err != nil {
		return nil, err
	}

	// Habilitamos logs SQL em desenvolvimento
	if getEnv("APP_ENV", "development") == "development" {
		db.LogMode(true)
	}

	// Configuramos a conexão
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.DB().SetConnMaxLifetime(time.Hour)

	return db, nil
}

// setupRouter configura o router gin
func setupRouter() *gin.Engine {
	// Definimos o modo do Gin
	if getEnv("APP_ENV", "development") == "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Configuramos o CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{getEnv("CORS_ALLOW_ORIGINS", "*")},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Use(gin.Recovery())

	return router
}

func main() {
	// Carregamos as variaveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	// Inicializamos o router
	router := setupRouter()

	// Incializamos os componentes
	userRepo := repositories.NewUserRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)

	// Utilitarios
	passwordUtil := utils.NewPasswordUtil(12)
	jwtUtil := utils.NewJWTUtil(utils.JWTConfig{
		AccessSecret:  getEnv("JWT_ACCESS_SECRET", "access_secret_key"),
		RefreshSecret: getEnv("JWT_REFRESH_SECRET", "refresh_secret_key"),
		Issuer:        getEnv("JWT_ISSUER", "aurora_backend"),
	})

	// Servicos de notificacao
	emailService := services.NewEmailService(services.EmailConfig{
		Host:         getEnv("SMTP_HOST", "smtp.example.com"),
		Port:         getEnvAsInt("SMTP_PORT", 587),
		Username:     getEnv("SMTP_USERNAME", "username"),
		Password:     getEnv("SMTP_PASSWORD", "password"),
		FromEmail:    getEnv("SMTP_FROM_EMAIL", "from@example.com"),
		FromName:     getEnv("SMTP_FROM_NAME", "Aurora"),
		TemplatesDir: getEnv("SMTP_TEMPLATES_DIR", "./templates/email"),
		IsSMTP:       true,
		ServiceType:  getEnv("EMAIL_SERVICE", "smtp"),
	})

	smsService := services.NewSMSService(services.SMSConfig{
		Provider:   getEnv("SMS_PROVIDER", "twilio"),
		AccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
		AuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
		FromNumber: getEnv("TWILIO_FROM_NUMBER", ""),
	})

	whatsAppService := services.NewWhatsAppService(services.WhatsAppConfig{
		Provider:      getEnv("WHATSAPP_PROVIDER", "twilio"),
		PhoneNumberID: getEnv("META_PHONE_NUMBER_ID", ""),
		AccessToken:   getEnv("META_ACCESS_TOKEN", ""),
		AccountSID:    getEnv("TWILIO_ACCOUNT_SID", ""),
		AuthToken:     getEnv("TWILIO_AUTH_TOKEN", ""),
		FromNumber:    getEnv("TWILIO_WHATSAPP_FROM", ""),
	})

	authService := services.NewAuthService(
		userRepo,
		tokenRepo,
		passwordUtil,
		jwtUtil,
		emailService,
		smsService,
		whatsAppService,
		services.DefaultAuthConfig(),
	)

	// Middlewares
	authMiddleware := middlewares.NewAuthMiddleware(authService)

	// Controladores
	clientAuthController := controllers.NewClientAuthController(authService)
	professionalAuthController := controllers.NewProfessionalAuthController(authService)

	// Configuracao das rotas
	api := router.Group("/api/v1")

	// Rotas de cliente
	clientRoutes := api.Group("/client")
	clientAuthController.RegisterRoutes(clientRoutes)

	// Rotas protegidas do cliente
	clientProtected := clientRoutes.Group("")
	clientProtected.Use(authMiddleware.RequireAuth())
	clientProtected.Use(authMiddleware.RequireClient())
	{
		// Protect client routes
	}

	// Rotas do profissional
	professionalRoutes := api.Group("/professional")
	professionalAuthController.RegisterRoutes(professionalRoutes)

	// Rotas protegidas do profissional
	professionalProtected := professionalRoutes.Group("")
	professionalProtected.Use(authMiddleware.RequireAuth())
	professionalProtected.Use(authMiddleware.RequireProfessional())
	{
		// Protect professional routes
	}

	// Inicia o servidor
	port := getEnv("PORT", "8080")
	log.Printf("Servidor iniciado na porta %s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}

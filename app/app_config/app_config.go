package appconfig

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
	"strings"
)

func ReadConfig() *viper.Viper {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	v := viper.New()
	v.SetConfigType("json")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind environment variables
	v.BindEnv("app.id", "APP_ID")
	v.BindEnv("app.env", "APP_ENV")
	v.BindEnv("app.host", "APP_HOST")
	v.BindEnv("app.port", "APP_PORT")
	v.BindEnv("app.httplogs", "APP_HTTPLOGS")
	v.BindEnv("db.postgres.dsn", "DB_POSTGRES_DSN")
	v.BindEnv("payment.stripe_secret_key", "STRIPE_SECRET_KEY")
	v.BindEnv("payment.stripe_endpoint_secret", "STRIPE_ENDPOINT_SECRET")
	v.BindEnv("payment.authorize_login_id", "AUTHORIZE_LOGIN_ID")
	v.BindEnv("payment.authorize_transaction_key", "AUTHORIZE_TRANSACTION_KEY")
	v.BindEnv("payment.authorize_net_webhook_signature_key", "AUTHORIZE_NET_WEBHOOK_SIGNATURE_KEY")
	v.Set("db.postgres.driver", "postgres")
	v.Set("db.postgres.name", "postgres")

	return v
}

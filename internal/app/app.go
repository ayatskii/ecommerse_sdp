package app

import (
	"fmt"
	"os"

	"github.com/ecommerce/payment-system/config"
	"github.com/ecommerce/payment-system/internal/facade"
	"github.com/ecommerce/payment-system/internal/observer"
	"github.com/ecommerce/payment-system/internal/repository"
	"github.com/ecommerce/payment-system/internal/service"
	"github.com/ecommerce/payment-system/pkg/logger"
)

type Application struct {
	Config          *config.Config
	Repository      repository.Repository
	CartService     *service.CartService
	CustomerService *service.CustomerService
	CheckoutFacade  *facade.CheckoutFacade
	EventSubject    *observer.Subject
}

func Initialize(configPath string) (*Application, error) {

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := logger.Init(
		cfg.Logging.Level,
		cfg.Logging.Format,
		cfg.Logging.Output,
		cfg.Logging.FilePath,
	); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Info(fmt.Sprintf("Starting %s v%s", cfg.App.Name, cfg.App.Version))

	var repo repository.Repository

	useDatabase := cfg.App.Environment == "production" || os.Getenv("USE_DATABASE") == "true"

	if useDatabase {
		repo, err = repository.NewSQLiteRepository(cfg.Database.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
		fmt.Println("✓ Using SQLite database")
	} else {
		repo, err = repository.NewFileRepository("data/store.json")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize file repository: %w", err)
		}
	}

	cartService := service.NewCartService(repo)
	customerService := service.NewCustomerService(repo)

	eventSubject := observer.NewSubject()

	if cfg.Notifications.Email.Enabled {
		emailNotifier := observer.NewEmailNotifier(
			cfg.Notifications.Email.FromAddress,
			cfg.Notifications.Email.SMTPHost,
			cfg.Notifications.Email.SMTPPort,
			cfg.Notifications.Email.WorkerPoolSize,
		)
		eventSubject.Attach(emailNotifier)
	}

	if cfg.Notifications.SMS.Enabled {
		smsNotifier := observer.NewSMSNotifier(
			cfg.Notifications.SMS.Provider,
			cfg.Notifications.SMS.RateLimit,
		)
		eventSubject.Attach(smsNotifier)
	}

	if cfg.Notifications.Audit.Enabled {
		auditLogger, err := observer.NewAuditLogger(cfg.Notifications.Audit.LogPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create audit logger: %w", err)
		}
		eventSubject.Attach(auditLogger)
	}

	if cfg.Metrics.Enabled {
		metricsCollector := observer.NewMetricsCollector(cfg.Metrics.ExportInterval)
		eventSubject.Attach(metricsCollector)
	}

	checkoutFacade := facade.NewCheckoutFacade(cfg, repo, eventSubject)

	app := &Application{
		Config:          cfg,
		Repository:      repo,
		CartService:     cartService,
		CustomerService: customerService,
		CheckoutFacade:  checkoutFacade,
		EventSubject:    eventSubject,
	}

	logger.Info("Application initialized successfully")

	return app, nil
}

func (a *Application) Shutdown() error {
	logger.Info("Shutting down application")

	if err := a.Repository.Close(); err != nil {
		logger.Error(fmt.Sprintf("Failed to close repository: %v", err))
	}

	if err := logger.Sync(); err != nil {
		return err
	}

	return nil
}

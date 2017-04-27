package app

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/julienbayle/go-json-rest-middleware-jwt"
	"github.com/julienbayle/jeparticipe/email"
	"github.com/julienbayle/jeparticipe/services"

	"fmt"
)

const (
	TestMode = "test"
	ProdMode = "prod"
)

type App struct {
	RepositoryService  *services.RepositoryService
	ActivityService    *services.ActivityService
	EventService       *services.EventService
	Secret             string
	SuperAdminPassword string
}

// Inits a new "Jeparticipe" application
func NewApp(dbFilePath string) *App {
	repositoryService := services.NewRepositoryService(dbFilePath)
	repositoryService.CreateCollectionIfNotExists(services.EventsBucketName)
	repositoryService.CreateCollectionIfNotExists(services.PropertiesBucketName)

	// App secret is used to generate tokens (event confirmation code, JWT toket, ...)
	secret := services.GetProperty(repositoryService, "secret", services.NewPassword(64))

	// Superadmin password allows to be admin in all events
	superAdminPassword := services.GetProperty(repositoryService, "superadminpass", services.NewPassword(12))

	return &App{
		Secret:             secret,
		SuperAdminPassword: superAdminPassword,
		RepositoryService:  repositoryService,
		ActivityService: &services.ActivityService{
			RepositoryService: repositoryService,
		},
		EventService: &services.EventService{
			RepositoryService: repositoryService,
			EmailRelay: &email.EmailRelay{
				Send: email.SendWithMailjet,
			},
			Secret: secret,
		},
	}
}

// Closes socket or open files on shutdown
func (app *App) ShutDown() {
	app.RepositoryService.ShutDown()
}

// Build an "jeparticipe" API endpoint
func (app *App) BuildApi(mode string, baseUrl string) *rest.Api {

	// Initialize the API endpoint
	api := rest.NewApi()

	if mode == ProdMode {
		api.Use(rest.DefaultProdStack...)
	} else {
		api.Use(rest.DefaultCommonStack...)
	}

	// Init CORS middleware
	api.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			return true
		},
		AllowedMethods:                []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:                []string{"Accept", "Content-Type", "Origin", "Authorization"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})

	// Init JWT middleware
	jwt_middleware := &jwt.JWTMiddleware{
		Key:   []byte(app.Secret),
		Realm: "Jeparticipe auth",
		Authenticator: func(userId string, password string) bool {
			return services.Authenticate(app.EventService, app.SuperAdminPassword, userId, password)
		},
		LogFunc: func(logMessage string) {
			fmt.Printf("JWT Middleware : %s", logMessage)
		},
	}

	// Use the JWT middleware only if there is a JWT token
	api.Use(&rest.IfMiddleware{
		Condition: func(request *rest.Request) bool {
			return request.Header.Get("Authorization") != ""
		},
		IfTrue: jwt_middleware,
	})

	// Adds routes
	uLogin := baseUrl + "/login"
	uBackup := baseUrl + "/backup"
	uEvent := baseUrl + "/event"
	uBucket := uEvent + "/:event/activity/:acode"

	router, err := rest.MakeRouter(
		rest.Post(uLogin, jwt_middleware.LoginHandler),

		rest.Get(uBackup, app.RepositoryService.Backup),

		rest.Post(uEvent, app.EventService.CreatePendingEvent),
		rest.Get(uEvent+"/:event/lostaccount", app.EventService.SendEventInformationByMail),
		rest.Get(uEvent+"/:event/confirm/:confirm_code", app.EventService.ConfirmEvent),
		rest.Get(uEvent+"/:event/status", app.EventService.GetEventStatus),
		rest.Get(uEvent+"/:event/config", app.EventService.GetEventConfig),
		rest.Put(uEvent+"/:event/config", app.EventService.SetEventConfig),

		rest.Get(uBucket, app.ActivityService.GetActivity),
		rest.Put(uBucket+"/state/:state", app.ActivityService.UpdateActivityState),
		rest.Put(uBucket+"/participant", app.ActivityService.AddAParticipantToAnActivity),
		rest.Get(uBucket+"/participant/:pcode/delete", app.ActivityService.RemoveAParticipantFromAnActivity),
	)

	if err != nil {
		panic(err)
	}

	// Start REST API
	api.SetApp(router)

	return api
}

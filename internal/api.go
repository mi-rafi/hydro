package internal

import (
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/rs/zerolog/log"
)

// Context structure for handle context from main.
type Context struct {
	echo.Context
	Ctx context.Context
}

// API structure containing the necessary server settings and responsible for starting and stopping it.
type API struct {
	e    *echo.Echo
	addr string
	cli  HydroponicClient
	repo HydroponicRepo
	t    TimeLoader
}

// AppConfig structure containing the server settings necessary for its operation.
type AppConfig struct {
	NetInterface string
	Timeout      time.Duration
}

func (ac *AppConfig) checkConfig() {
	log.Debug().Msg("checking api application config")

	if ac.NetInterface == "" {
		ac.NetInterface = "localhost:9000"
	}
	if ac.Timeout <= 0 {
		ac.Timeout = 10 * time.Millisecond
	}
}

// SearchRequest is strust for storage and validate query param.
type SearchRequest struct {
	Start time.Time `validate:"required" query:"s"`
	End   time.Time `validate:"required" query:"e"`
}

type ChangePhRequest struct {
	IsUp bool `validate:"required" query:"up"`
}

type TimeLoadResponse struct {
	LastTime time.Time `json:"lastTime"`
}

type TimeLoadRequest struct {
	LastTime time.Time `json:"lastTime"  validate:"required"`
}

// Validator - to add custom validator in echo.
type Validator struct {
	validator *validator.Validate
}

// Validate add go-playground/validator in echo.
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

// NewApp returns a new ready-to-launch API object with adjusted settings.
func NewApp(ctx context.Context, appCfg AppConfig, hc HydroponicClient, hr HydroponicRepo, t TimeLoader) (*API, error) {
	appCfg.checkConfig()

	log.Debug().Interface("api app config", appCfg).Msg("starting initialize api application")

	e := echo.New()

	a := &API{
		e:    e,
		addr: appCfg.NetInterface,
		cli:  hc,
		repo: hr,
		t:    t,
	}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &Context{
				Context: c,
				Ctx:     ctx,
			}
			return next(cc)
		}
	})
	e.Validator = &Validator{validator: validator.New()}
	e.Use(logMiddleware)

	e.GET("/healthcheck", a.handleHealthcheck)

	g := e.Group("/api")
	g.GET("/light", a.handleLightState)
	g.GET("/data", a.handleSearch)
	g.GET("/time", a.handleLoadTime)
	g.POST("/time", a.handleStoreTime)
	g.POST("/light", a.handleChangeLight)
	g.POST("/ph", a.handleChangePh)
	g.POST("/soil", a.handleAddSoil)
	g.POST("/water", a.handleAddWater)

	log.Debug().Msg("endpoints registered")

	return a, nil
}

func (a *API) handleHealthcheck(c echo.Context) error {
	return ok(c)
}

func (a *API) handleLoadTime(c echo.Context) error {
	st, err := a.t.GetStartupTime()
	if err != nil {
		log.Error().Err(err).Msg("can not load time from file")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, &TimeLoadResponse{LastTime: st})
}

func (a *API) handleStoreTime(c echo.Context) error {
	var err error

	request := &TimeLoadRequest{}
	if err := c.Bind(request); err != nil {
		log.Debug().Err(err).Msg("handleStoreTime Bind err")
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	log.Debug().
		Time("storeTime", request.LastTime).
		Msg("handleSearch run")

	if err = c.Validate(request); err != nil {
		log.Debug().Err(err).Msg("handleStoreTime Validate err")
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	err = a.t.StoreStartupTime(request.LastTime)
	if err != nil {
		log.Error().Err(err).Msg("can not store time to file")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return ok(c)
}

func (a *API) handleSearch(c echo.Context) error {
	var err error

	request := &SearchRequest{}
	if err := c.Bind(request); err != nil {
		log.Debug().Err(err).Msg("handleSearch Bind err")
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	log.Debug().
		Time("start", request.Start).
		Time("end", request.End).
		Msg("handleSearch run")

	if err = c.Validate(request); err != nil {
		log.Debug().Err(err).Msg("handleSearch Validate err")
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	var ctx context.Context
	cc, b := c.(Context)
	if !b {
		log.Warn().Msg("incorrect context, use common")
		ctx = context.Background()
	} else {
		ctx = cc.Ctx
	}
	r, err := a.repo.GetLastData(ctx, request.Start, request.End)

	if err != nil {
		log.Err(err).Msg("saving error")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, r)
}

func (a *API) handleLightState(c echo.Context) error {
	log.Debug().Msg("handleLightState run")
	r := a.cli.GetLightState()

	return c.JSON(http.StatusOK, r)
}

func (a *API) handleAddSoil(c echo.Context) error {
	a.cli.SendAddSoil()
	return ok(c)
}

func (a *API) handleAddWater(c echo.Context) error {
	a.cli.SendAddWater()
	return ok(c)
}

func (a *API) handleChangePh(c echo.Context) error {

	request := &PhState{}
	if err := c.Bind(request); err != nil {
		log.Debug().Err(err).Msg("handleSearch Bind err")
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	if err := c.Validate(request); err != nil {
		log.Debug().Err(err).Msg("handleSearch Validate err")
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	if request.IsUp {
		a.cli.SendUpPh()
	} else {
		a.cli.SendDownPh()
	}
	return ok(c)
}

func (a *API) handleChangeLight(c echo.Context) error {
	a.cli.SendChangeLight()
	return ok(c)
}

// Run start the server.
func (a *API) Run() error {
	return a.e.Start(a.addr)
}

// Close stop the server.
func (a *API) Close() error {
	log.Debug().Msg("shutting down server")
	return a.e.Close()
}

func ok(c echo.Context) error {
	return c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
}

func logMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		res := c.Response()
		start := time.Now()

		err := next(c)

		stop := time.Now()

		log.Debug().
			Str("remote", req.RemoteAddr).
			Str("user_agent", req.UserAgent()).
			Str("method", req.Method).
			Str("path", c.Path()).
			Int("status", res.Status).
			Dur("duration", stop.Sub(start)).
			Str("duration_human", stop.Sub(start).String()).
			Msgf("called url %s", req.URL)

		return err
	}
}

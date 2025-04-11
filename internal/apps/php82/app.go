package php82

import (
	"github.com/go-chi/chi/v5"
	"github.com/leonelquinteros/gotext"

	"github.com/tnb-labs/panel/internal/apps/php"
	"github.com/tnb-labs/panel/internal/biz"
)

type App struct {
	php *php.App
}

func NewApp(t *gotext.Locale, task biz.TaskRepo) *App {
	return &App{
		php: php.NewApp(t, task),
	}
}

func (s *App) Route(r chi.Router) {
	s.php.Route(82)(r)
}

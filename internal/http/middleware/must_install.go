package middleware

import (
	"net/http"
	"strings"

	"github.com/go-rat/chix"
	"github.com/leonelquinteros/gotext"

	"github.com/tnb-labs/panel/internal/biz"
)

// MustInstall 确保已安装应用
func MustInstall(t *gotext.Locale, app biz.AppRepo) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var slugs []string
			if strings.HasPrefix(r.URL.Path, "/api/website") {
				slugs = append(slugs, "nginx")
			} else if strings.HasPrefix(r.URL.Path, "/api/container") {
				slugs = append(slugs, "podman", "docker")
			} else if strings.HasPrefix(r.URL.Path, "/api/apps/") {
				pathArr := strings.Split(r.URL.Path, "/")
				if len(pathArr) < 4 {
					render := chix.NewRender(w)
					defer render.Release()
					render.Status(http.StatusForbidden)
					render.JSON(chix.M{
						"msg": t.Get("app not found"),
					})
					return
				}
				slugs = append(slugs, pathArr[3])
			}

			flag := false
			for _, s := range slugs {
				if installed, _ := app.IsInstalled("slug = ?", s); installed {
					flag = true
					break
				}
			}
			if !flag && len(slugs) > 0 {
				render := chix.NewRender(w)
				defer render.Release()
				render.Status(http.StatusForbidden)
				render.JSON(chix.M{
					"msg": t.Get("app %s not installed", slugs),
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

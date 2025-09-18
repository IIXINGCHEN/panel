package apps

import (
	"github.com/google/wire"

	"github.com/acepanel/panel/internal/apps/codeserver"
	"github.com/acepanel/panel/internal/apps/docker"
	"github.com/acepanel/panel/internal/apps/fail2ban"
	"github.com/acepanel/panel/internal/apps/frp"
	"github.com/acepanel/panel/internal/apps/gitea"
	"github.com/acepanel/panel/internal/apps/memcached"
	"github.com/acepanel/panel/internal/apps/minio"
	"github.com/acepanel/panel/internal/apps/mysql"
	"github.com/acepanel/panel/internal/apps/nginx"
	"github.com/acepanel/panel/internal/apps/php74"
	"github.com/acepanel/panel/internal/apps/php80"
	"github.com/acepanel/panel/internal/apps/php81"
	"github.com/acepanel/panel/internal/apps/php82"
	"github.com/acepanel/panel/internal/apps/php83"
	"github.com/acepanel/panel/internal/apps/php84"
	"github.com/acepanel/panel/internal/apps/phpmyadmin"
	"github.com/acepanel/panel/internal/apps/podman"
	"github.com/acepanel/panel/internal/apps/postgresql"
	"github.com/acepanel/panel/internal/apps/pureftpd"
	"github.com/acepanel/panel/internal/apps/redis"
	"github.com/acepanel/panel/internal/apps/rsync"
	"github.com/acepanel/panel/internal/apps/s3fs"
	"github.com/acepanel/panel/internal/apps/supervisor"
)

var ProviderSet = wire.NewSet(
	codeserver.NewApp,
	docker.NewApp,
	fail2ban.NewApp,
	frp.NewApp,
	gitea.NewApp,
	memcached.NewApp,
	minio.NewApp,
	mysql.NewApp,
	nginx.NewApp,
	php74.NewApp,
	php80.NewApp,
	php81.NewApp,
	php82.NewApp,
	php83.NewApp,
	php84.NewApp,
	phpmyadmin.NewApp,
	podman.NewApp,
	postgresql.NewApp,
	pureftpd.NewApp,
	redis.NewApp,
	rsync.NewApp,
	s3fs.NewApp,
	supervisor.NewApp,
)

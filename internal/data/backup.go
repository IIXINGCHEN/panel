package data

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/shirou/gopsutil/disk"
	"gorm.io/gorm"

	"github.com/tnb-labs/panel/internal/app"
	"github.com/tnb-labs/panel/internal/biz"
	"github.com/tnb-labs/panel/pkg/db"
	"github.com/tnb-labs/panel/pkg/io"
	"github.com/tnb-labs/panel/pkg/shell"
	"github.com/tnb-labs/panel/pkg/tools"
	"github.com/tnb-labs/panel/pkg/types"
)

type backupRepo struct {
	db      *gorm.DB
	setting biz.SettingRepo
	website biz.WebsiteRepo
}

func NewBackupRepo(db *gorm.DB, setting biz.SettingRepo, website biz.WebsiteRepo) biz.BackupRepo {
	return &backupRepo{
		db:      db,
		setting: setting,
		website: website,
	}
}

// List 备份列表
func (r *backupRepo) List(typ biz.BackupType) ([]*types.BackupFile, error) {
	path, err := r.GetPath(typ)
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	list := make([]*types.BackupFile, 0)
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		list = append(list, &types.BackupFile{
			Name: file.Name(),
			Path: filepath.Join(path, file.Name()),
			Size: tools.FormatBytes(float64(info.Size())),
			Time: info.ModTime(),
		})
	}

	return list, nil
}

// Create 创建备份
// typ 备份类型
// target 目标名称
// path 可选备份保存路径
func (r *backupRepo) Create(typ biz.BackupType, target string, path ...string) error {
	defPath, err := r.GetPath(typ)
	if err != nil {
		return err
	}
	if len(path) > 0 && path[0] != "" {
		defPath = path[0]
	}

	switch typ {
	case biz.BackupTypeWebsite:
		return r.createWebsite(defPath, target)
	case biz.BackupTypeMySQL:
		return r.createMySQL(defPath, target)
	case biz.BackupTypePostgres:
		return r.createPostgres(defPath, target)
	case biz.BackupTypePanel:
		return r.createPanel(defPath)

	}

	return errors.New("未知备份类型")
}

// Delete 删除备份
func (r *backupRepo) Delete(typ biz.BackupType, name string) error {
	path, err := r.GetPath(typ)
	if err != nil {
		return err
	}

	file := filepath.Join(path, name)
	return io.Remove(file)
}

// Restore 恢复备份
// typ 备份类型
// backup 备份压缩包，可以是绝对路径或者相对路径
// target 目标名称
func (r *backupRepo) Restore(typ biz.BackupType, backup, target string) error {
	if !io.Exists(backup) {
		path, err := r.GetPath(typ)
		if err != nil {
			return err
		}
		backup = filepath.Join(path, backup)
	}

	switch typ {
	case biz.BackupTypeWebsite:
		return r.restoreWebsite(backup, target)
	case biz.BackupTypeMySQL:
		return r.restoreMySQL(backup, target)
	case biz.BackupTypePostgres:
		return r.restorePostgres(backup, target)
	}

	return errors.New("未知备份类型")
}

// CutoffLog 切割日志
// path 保存目录绝对路径
// target 待切割日志文件绝对路径
func (r *backupRepo) CutoffLog(path, target string) error {
	if !io.Exists(target) {
		return errors.New("日志文件不存在")
	}

	to := filepath.Join(path, fmt.Sprintf("%s_%s.zip", time.Now().Format("20060102150405"), filepath.Base(target)))
	if err := io.Compress(filepath.Dir(target), []string{filepath.Base(target)}, to); err != nil {
		return err
	}

	return io.Remove(target)
}

// ClearExpired 清理过期备份
// path 备份目录绝对路径
// prefix 目标文件前缀
// save 保存份数
func (r *backupRepo) ClearExpired(path, prefix string, save int) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var filtered []os.FileInfo
	for _, file := range files {
		if strings.HasPrefix(file.Name(), prefix) && strings.HasSuffix(file.Name(), ".zip") {
			info, err := os.Stat(filepath.Join(path, file.Name()))
			if err != nil {
				continue
			}
			filtered = append(filtered, info)
		}
	}

	// 排序所有备份文件，从新到旧
	slices.SortFunc(filtered, func(a, b os.FileInfo) int {
		if a.ModTime().After(b.ModTime()) {
			return -1
		}
		if a.ModTime().Before(b.ModTime()) {
			return 1
		}
		return 0
	})
	if len(filtered) <= save {
		return nil
	}

	// 切片保留 save 份，删除剩余
	toDelete := filtered[save:]
	for _, file := range toDelete {
		filePath := filepath.Join(path, file.Name())
		if app.IsCli {
			fmt.Printf("|-清理过期文件：%s\n", filePath)
		}
		if err = os.Remove(filePath); err != nil {
			if app.IsCli {
				fmt.Printf("|-清理失败：%v\n", err)
			} else {
				return fmt.Errorf("清理失败：%v", err)
			}
		}
	}

	return nil
}

// GetPath 获取备份路径
func (r *backupRepo) GetPath(typ biz.BackupType) (string, error) {
	backupPath, err := r.setting.Get(biz.SettingKeyBackupPath)
	if err != nil {
		return "", err
	}
	if !slices.Contains([]biz.BackupType{biz.BackupTypePath, biz.BackupTypeWebsite, biz.BackupTypeMySQL, biz.BackupTypePostgres, biz.BackupTypeRedis, biz.BackupTypePanel}, typ) {
		return "", errors.New("未知备份类型")
	}

	backupPath = filepath.Join(backupPath, string(typ))
	if !io.Exists(backupPath) {
		if err = os.MkdirAll(backupPath, 0644); err != nil {
			return "", err
		}
	}

	return backupPath, nil
}

// createWebsite 创建网站备份
func (r *backupRepo) createWebsite(to string, name string) error {
	website, err := r.website.GetByName(name)
	if err != nil {
		return err
	}

	if err = r.preCheckPath(to, website.Path); err != nil {
		return err
	}

	start := time.Now()
	backup := filepath.Join(to, fmt.Sprintf("%s_%s.zip", website.Name, time.Now().Format("20060102150405")))
	if err = io.Compress(website.Path, nil, backup); err != nil {
		return err
	}

	if app.IsCli {
		fmt.Printf("|-备份耗时：%s\n", time.Since(start).String())
		fmt.Printf("|-已备份至文件：%s\n", filepath.Base(backup))
	}
	return nil
}

// createMySQL 创建 MySQL 备份
func (r *backupRepo) createMySQL(to string, name string) error {
	rootPassword, err := r.setting.Get(biz.SettingKeyMySQLRootPassword)
	if err != nil {
		return err
	}
	mysql, err := db.NewMySQL("root", rootPassword, "/tmp/mysql.sock", "unix")
	if err != nil {
		return err
	}
	defer mysql.Close()
	if exist, _ := mysql.DatabaseExists(name); !exist {
		return fmt.Errorf("数据库不存在：%s", name)
	}
	size, err := mysql.DatabaseSize(name)
	if err != nil {
		return err
	}
	if err = r.preCheckDB(to, size); err != nil {
		return err
	}

	if err = os.Setenv("MYSQL_PWD", rootPassword); err != nil {
		return err
	}
	start := time.Now()
	backup := filepath.Join(to, fmt.Sprintf("%s_%s.sql", name, time.Now().Format("20060102150405")))
	if _, err = shell.Execf(`mysqldump -u root '%s' > '%s'`, name, backup); err != nil {
		return err
	}
	if err = os.Unsetenv("MYSQL_PWD"); err != nil {
		return err
	}

	if err = io.Compress(filepath.Dir(backup), []string{filepath.Base(backup)}, backup+".zip"); err != nil {
		return err
	}
	if err = io.Remove(backup); err != nil {
		return err
	}

	if app.IsCli {
		fmt.Printf("|-备份耗时：%s\n", time.Since(start).String())
		fmt.Printf("|-已备份至文件：%s\n", filepath.Base(backup+".zip"))
	}
	return nil
}

// createPostgres 创建 PostgreSQL 备份
func (r *backupRepo) createPostgres(to string, name string) error {
	postgres, err := db.NewPostgres("postgres", "", "127.0.0.1", 5432)
	if err != nil {
		return err
	}
	defer postgres.Close()
	if exist, _ := postgres.DatabaseExist(name); !exist {
		return fmt.Errorf("数据库不存在：%s", name)
	}
	size, err := postgres.DatabaseSize(name)
	if err != nil {
		return err
	}
	if err = r.preCheckDB(to, size); err != nil {
		return err
	}

	start := time.Now()
	backup := filepath.Join(to, fmt.Sprintf("%s_%s.sql", name, time.Now().Format("20060102150405")))
	if _, err = shell.Execf(`su - postgres -c "pg_dump '%s'" > '%s'`, name, backup); err != nil {
		return err
	}

	if err = io.Compress(filepath.Dir(backup), []string{filepath.Base(backup)}, backup+".zip"); err != nil {
		return err
	}
	if err = io.Remove(backup); err != nil {
		return err
	}

	if app.IsCli {
		fmt.Printf("|-备份耗时：%s\n", time.Since(start).String())
		fmt.Printf("|-已备份至文件：%s\n", filepath.Base(backup+".zip"))
	}
	return nil
}

// createPanel 创建面板备份
func (r *backupRepo) createPanel(to string) error {
	backup := filepath.Join(to, fmt.Sprintf("panel_%s.zip", time.Now().Format("20060102150405")))

	if err := r.preCheckPath(to, filepath.Join(app.Root, "panel")); err != nil {
		return err
	}

	start := time.Now()

	temp, err := os.MkdirTemp("", "panel-backup")
	if err != nil {
		return err
	}

	if err = io.Cp(filepath.Join(app.Root, "panel"), temp); err != nil {
		return err
	}
	if err = io.Cp("/usr/local/sbin/panel-cli", temp); err != nil {
		return err
	}
	if err = io.Cp("/usr/local/etc/panel/config.yml", temp); err != nil {
		return err
	}

	_ = io.Chmod(temp, 0600)
	if err = io.Compress(temp, nil, backup); err != nil {
		return err
	}
	if err = io.Chmod(backup, 0600); err != nil {
		return err
	}

	if app.IsCli {
		fmt.Printf("|-备份耗时：%s\n", time.Since(start).String())
		fmt.Printf("|-已备份至文件：%s\n", filepath.Base(backup))
	}

	return io.Remove(temp)
}

// restoreWebsite 恢复网站备份
func (r *backupRepo) restoreWebsite(backup, target string) error {
	if !io.Exists(backup) {
		return errors.New("备份文件不存在")
	}

	website, err := r.website.GetByName(target)
	if err != nil {
		return err
	}

	if err = io.Remove(website.Path); err != nil {
		return err
	}
	if err = io.UnCompress(backup, website.Path); err != nil {
		return err
	}
	if err = io.Chmod(website.Path, 0755); err != nil {
		return err
	}
	if err = io.Chown(website.Path, "www", "www"); err != nil {
		return err
	}

	return nil
}

// restoreMySQL 恢复 MySQL 备份
func (r *backupRepo) restoreMySQL(backup, target string) error {
	if !io.Exists(backup) {
		return errors.New("备份文件不存在")
	}

	rootPassword, err := r.setting.Get(biz.SettingKeyMySQLRootPassword)
	if err != nil {
		return err
	}
	mysql, err := db.NewMySQL("root", rootPassword, "/tmp/mysql.sock", "unix")
	if err != nil {
		return err
	}
	defer mysql.Close()
	if exist, _ := mysql.DatabaseExists(target); !exist {
		return fmt.Errorf("数据库不存在：%s", target)
	}
	if err = os.Setenv("MYSQL_PWD", rootPassword); err != nil {
		return err
	}

	clean := false
	if !strings.HasSuffix(backup, ".sql") {
		backup, err = r.autoUnCompressSQL(backup)
		if err != nil {
			return err
		}
		clean = true
	}

	if _, err = shell.Execf(`mysql -u root '%s' < '%s'`, target, backup); err != nil {
		return err
	}
	if err = os.Unsetenv("MYSQL_PWD"); err != nil {
		return err
	}
	if clean {
		_ = io.Remove(filepath.Dir(backup))
	}

	return nil
}

// restorePostgres 恢复 PostgreSQL 备份
func (r *backupRepo) restorePostgres(backup, target string) error {
	if !io.Exists(backup) {
		return errors.New("备份文件不存在")
	}

	postgres, err := db.NewPostgres("postgres", "", "127.0.0.1", 5432)
	if err != nil {
		return err
	}
	defer postgres.Close()
	if exist, _ := postgres.DatabaseExist(target); !exist {
		return fmt.Errorf("数据库不存在：%s", target)
	}

	clean := false
	if !strings.HasSuffix(backup, ".sql") {
		backup, err = r.autoUnCompressSQL(backup)
		if err != nil {
			return err
		}
		clean = true
	}

	if _, err = shell.Execf(`su - postgres -c "psql '%s'" < '%s'`, target, backup); err != nil {
		return err
	}
	if clean {
		_ = io.Remove(filepath.Dir(backup))
	}

	return nil
}

// preCheckPath 预检空间和 inode 是否足够
// to 备份保存目录
// path 待备份目录
func (r *backupRepo) preCheckPath(to, path string) error {
	size, err := io.SizeX(path)
	if err != nil {
		return err
	}
	files, err := io.CountX(path)
	if err != nil {
		return err
	}

	usage, err := disk.Usage(to)
	if err != nil {
		return err
	}

	if app.IsCli {
		fmt.Printf("|-目标大小：%s\n", tools.FormatBytes(float64(size)))
		fmt.Printf("|-目标文件数：%d\n", files)
		fmt.Printf("|-备份目录可用空间：%s\n", tools.FormatBytes(float64(usage.Free)))
		fmt.Printf("|-备份目录可用Inode：%d\n", usage.InodesFree)
	}

	if uint64(size) > usage.Free {
		return errors.New("备份目录空间不足")
	}
	// 对于 fuse 等文件系统，可能没有 inode 的概念
	/*if uint64(files) > usage.InodesFree {
		return errors.New("备份目录Inode不足")
	}*/

	return nil
}

// preCheckDB 预检空间和 inode 是否足够
// to 备份保存目录
// size 数据库大小
func (r *backupRepo) preCheckDB(to string, size int64) error {
	usage, err := disk.Usage(to)
	if err != nil {
		return err
	}

	if app.IsCli {
		fmt.Printf("|-目标大小：%s\n", tools.FormatBytes(float64(size)))
		fmt.Printf("|-备份目录可用空间：%s\n", tools.FormatBytes(float64(usage.Free)))
		fmt.Printf("|-备份目录可用Inode：%d\n", usage.InodesFree)
	}

	if uint64(size) > usage.Free {
		return errors.New("备份目录空间不足")
	}

	return nil
}

// autoUnCompressSQL 自动处理压缩文件
func (r *backupRepo) autoUnCompressSQL(backup string) (string, error) {
	temp, err := os.MkdirTemp("", "sql-uncompress")
	if err != nil {
		return "", err
	}

	if err = io.UnCompress(backup, temp); err != nil {
		return "", err
	}

	backup = "" // 置空，防止干扰后续判断
	if files, err := os.ReadDir(temp); err == nil {
		if len(files) != 1 {
			return "", fmt.Errorf("压缩文件中包含的文件数量不为1，实际为%d", len(files))
		}
		if strings.HasSuffix(files[0].Name(), ".sql") {
			backup = filepath.Join(temp, files[0].Name())
		}
	}

	if backup == "" {
		return "", errors.New("无法找到.sql备份文件")
	}

	return backup, nil
}

func (r *backupRepo) FixPanel() error {
	if app.IsCli {
		fmt.Println("|-开始修复面板...")
	}

	// 检查关键文件是否正常
	flag := false
	if !io.Exists("/usr/local/etc/panel/config.yml") {
		flag = true
	}
	if !io.Exists(filepath.Join(app.Root, "panel", "web")) {
		flag = true
	}
	if !io.Exists(filepath.Join(app.Root, "panel", "storage", "app.db")) {
		flag = true
	}
	if io.Exists("/tmp/panel-storage.zip") {
		flag = true
	}
	// 检查数据库连接
	if err := r.db.Exec("VACUUM").Error; err != nil {
		flag = true
	}
	if err := r.db.Exec("PRAGMA wal_checkpoint(TRUNCATE);").Error; err != nil {
		flag = true
	}
	if !flag {
		return fmt.Errorf("文件正常无需修复，请运行 panel-cli update 更新面板")
	}

	// 再次确认是否需要修复
	if io.Exists("/tmp/panel-storage.zip") {
		// 文件齐全情况下只移除临时文件
		if io.Exists(filepath.Join(app.Root, "panel", "web")) &&
			io.Exists(filepath.Join(app.Root, "panel", "storage", "app.db")) &&
			io.Exists("/usr/local/etc/panel/config.yml") {
			if err := io.Remove("/tmp/panel-storage.zip"); err != nil {
				return fmt.Errorf("清理临时文件失败：%w", err)
			}
			if app.IsCli {
				fmt.Println("|-已清理临时文件，请运行 panel-cli update 更新面板")
			}
			return nil
		}
	}

	// 从备份目录中找最新的备份文件
	list, err := r.List(biz.BackupTypePanel)
	if err != nil {
		return err
	}
	slices.SortFunc(list, func(a *types.BackupFile, b *types.BackupFile) int {
		return int(b.Time.Unix() - a.Time.Unix())
	})
	if len(list) == 0 {
		return fmt.Errorf("未找到备份文件，无法自动修复")
	}
	latest := list[0]
	if app.IsCli {
		fmt.Printf("|-使用备份文件：%s\n", latest.Name)
	}

	// 解压备份文件
	if app.IsCli {
		fmt.Println("|-解压备份文件...")
	}
	if err = io.Remove("/tmp/panel-fix"); err != nil {
		return fmt.Errorf("清理临时目录失败：%w", err)
	}
	if err = io.UnCompress(latest.Path, "/tmp/panel-fix"); err != nil {
		return fmt.Errorf("解压备份文件失败：%w", err)
	}

	// 移动文件到对应位置
	if app.IsCli {
		fmt.Println("|-移动备份文件...")
	}
	if io.Exists(filepath.Join("/tmp/panel-fix", "panel")) && io.IsDir(filepath.Join("/tmp/panel-fix", "panel")) {
		if err = io.Remove(filepath.Join(app.Root, "panel")); err != nil {
			return fmt.Errorf("删除目录失败：%w", err)
		}
		if err = io.Mv(filepath.Join("/tmp/panel-fix", "panel"), filepath.Join(app.Root)); err != nil {
			return fmt.Errorf("移动目录失败：%w", err)
		}
	}
	if io.Exists(filepath.Join("/tmp/panel-fix", "config.yml")) {
		if err = io.Mv(filepath.Join("/tmp/panel-fix", "config.yml"), "/usr/local/etc/panel/config.yml"); err != nil {
			return fmt.Errorf("移动文件失败：%w", err)
		}
	}
	if io.Exists(filepath.Join("/tmp/panel-fix", "panel-cli")) {
		if err = io.Mv(filepath.Join("/tmp/panel-fix", "panel-cli"), "/usr/local/sbin/panel-cli"); err != nil {
			return fmt.Errorf("移动文件失败：%w", err)
		}
	}

	// tmp 目录下如果有 storage 备份，则解压回去
	if app.IsCli {
		fmt.Println("|-恢复面板数据...")
	}
	if io.Exists("/tmp/panel-storage.zip") {
		if err = io.UnCompress("/tmp/panel-storage.zip", filepath.Join(app.Root, "panel")); err != nil {
			return fmt.Errorf("恢复面板数据失败：%w", err)
		}
		if err = io.Remove("/tmp/panel-storage.zip"); err != nil {
			return fmt.Errorf("清理临时文件失败：%w", err)
		}
	}

	// 下载服务文件
	if !io.Exists("/etc/systemd/system/panel.service") {
		if _, err = shell.Execf(`wget -O /etc/systemd/system/panel.service https://dl.cdn.haozi.net/panel/panel.service && sed -i "s|/www|%s|g" /etc/systemd/system/panel.service`, app.Root); err != nil {
			return err
		}
	}

	// 处理权限
	if app.IsCli {
		fmt.Println("|-设置关键文件权限...")
	}
	if err = io.Chmod("/usr/local/etc/panel/config.yml", 0600); err != nil {
		return err
	}
	if err = io.Chmod("/etc/systemd/system/panel.service", 0644); err != nil {
		return err
	}
	if err = io.Chmod("/usr/local/sbin/panel-cli", 0700); err != nil {
		return err
	}
	if err = io.Chmod(filepath.Join(app.Root, "panel"), 0700); err != nil {
		return err
	}

	if err = io.Remove("/tmp/panel-fix"); err != nil {
		return err
	}

	if app.IsCli {
		fmt.Println("|-修复完成")
	}

	tools.RestartPanel()
	return nil
}

func (r *backupRepo) UpdatePanel(version, url, checksum string) error {
	// 预先优化数据库
	if err := r.db.Exec("VACUUM").Error; err != nil {
		return err
	}
	if err := r.db.Exec("PRAGMA wal_checkpoint(TRUNCATE);").Error; err != nil {
		return err
	}

	name := filepath.Base(url)
	if app.IsCli {
		fmt.Printf("|-目标版本：%s\n", version)
		fmt.Printf("|-下载链接：%s\n", url)
		fmt.Printf("|-文件名：%s\n", name)
	}

	if app.IsCli {
		fmt.Println("|-正在下载...")
	}
	if _, err := shell.Execf("wget -T 120 -t 3 -O /tmp/%s %s", name, url); err != nil {
		return fmt.Errorf("下载失败：%w", err)
	}
	if _, err := shell.Execf("wget -T 20 -t 3 -O /tmp/%s %s", name+".sha256", checksum); err != nil {
		return fmt.Errorf("下载失败：%w", err)
	}
	if !io.Exists(filepath.Join("/tmp", name)) || !io.Exists(filepath.Join("/tmp", name+".sha256")) {
		return errors.New("下载文件检查失败")
	}

	if app.IsCli {
		fmt.Println("|-校验下载文件...")
	}
	if check, err := shell.Execf("cd /tmp && sha256sum -c %s --ignore-missing", name+".sha256"); check != name+": OK" || err != nil {
		return errors.New("下载文件校验失败")
	}
	if err := io.Remove(filepath.Join("/tmp", name+".sha256")); err != nil {
		if app.IsCli {
			fmt.Println("|-清理校验文件失败：", err)
		}
		return fmt.Errorf("清理校验文件失败：%w", err)
	}

	if app.IsCli {
		fmt.Println("|-前置检查...")
	}
	if io.Exists("/tmp/panel-storage.zip") {
		return errors.New("检测到 /tmp 存在临时文件，可能是上次更新失败所致，请运行 panel-cli fix 修复后重试")
	}

	if app.IsCli {
		fmt.Println("|-备份面板数据...")
	}
	// 备份面板
	if err := r.Create(biz.BackupTypePanel, ""); err != nil {
		if app.IsCli {
			fmt.Println("|-备份面板失败：", err)
		}
		return fmt.Errorf("备份面板失败：%w", err)
	}
	if err := io.Compress(filepath.Join(app.Root, "panel/storage"), nil, "/tmp/panel-storage.zip"); err != nil {
		if app.IsCli {
			fmt.Println("|-备份面板数据失败：", err)
		}
		return fmt.Errorf("备份面板数据失败：%w", err)
	}
	if !io.Exists("/tmp/panel-storage.zip") {
		return errors.New("已备份面板数据检查失败")
	}

	if app.IsCli {
		fmt.Println("|-清理旧版本...")
	}
	if _, err := shell.Execf("rm -rf %s/panel/*", app.Root); err != nil {
		return fmt.Errorf("清理旧版本失败：%w", err)
	}

	if app.IsCli {
		fmt.Println("|-解压新版本...")
	}
	if err := io.UnCompress(filepath.Join("/tmp", name), filepath.Join(app.Root, "panel")); err != nil {
		return fmt.Errorf("解压失败：%w", err)
	}
	if !io.Exists(filepath.Join(app.Root, "panel", "web")) {
		return errors.New("解压失败，缺失文件")
	}
	if err := io.Remove(filepath.Join("/tmp", name)); err != nil {
		return fmt.Errorf("清理临时文件失败：%w", err)
	}

	if app.IsCli {
		fmt.Println("|-恢复面板数据...")
	}
	if err := io.UnCompress("/tmp/panel-storage.zip", filepath.Join(app.Root, "panel", "storage")); err != nil {
		return fmt.Errorf("恢复面板数据失败：%w", err)
	}
	if !io.Exists(filepath.Join(app.Root, "panel/storage/app.db")) {
		return errors.New("恢复面板数据失败")
	}

	if app.IsCli {
		fmt.Println("|-运行更新后脚本...")
	}
	if _, err := shell.Execf("curl -fsLm 10 https://dl.cdn.haozi.net/panel/auto_update.sh | bash"); err != nil {
		return fmt.Errorf("运行面板更新后脚本失败：%w", err)
	}
	if _, err := shell.Execf(`wget -O /etc/systemd/system/panel.service https://dl.cdn.haozi.net/panel/panel.service && sed -i "s|/www|%s|g" /etc/systemd/system/panel.service`, app.Root); err != nil {
		return fmt.Errorf("下载面板服务文件失败：%w", err)
	}
	if _, err := shell.Execf("panel-cli setting write version %s", version); err != nil {
		return fmt.Errorf("写入面板版本号失败：%w", err)
	}
	if err := io.Mv(filepath.Join(app.Root, "panel/cli"), "/usr/local/sbin/panel-cli"); err != nil {
		return fmt.Errorf("移动面板命令行工具失败：%w", err)
	}

	if app.IsCli {
		fmt.Println("|-设置关键文件权限...")
	}
	_ = io.Chmod("/usr/local/sbin/panel-cli", 0700)
	_ = io.Chmod("/etc/systemd/system/panel.service", 0644)
	_ = io.Chmod(filepath.Join(app.Root, "panel"), 0700)

	if app.IsCli {
		fmt.Println("|-更新完成")
	}

	_, _ = shell.Execf("systemctl daemon-reload")
	_ = io.Remove("/tmp/panel-storage.zip")
	_ = io.Remove(filepath.Join(app.Root, "panel/config.example.yml"))
	if sqlDB, err := r.db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	tools.RestartPanel()

	return nil
}

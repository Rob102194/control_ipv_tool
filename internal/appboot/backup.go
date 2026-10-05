package appboot

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/Rob102194/control_ipv_tool/internal/app/usecases"
	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/platform"
)

// backupsARetener es cuántos backups se conservan como máximo; al crear uno
// nuevo se borran automáticamente los más viejos por encima de este número.
const backupsARetener = 7

// backupManager implementa httpapi.BackupManager: resuelve el directorio
// efectivo de backups (configuración del negocio, o <dirBD>/backups por
// defecto) y delega la mecánica de archivo a internal/platform.
type backupManager struct {
	db     *sql.DB
	dbPath string
	cfgSvc *usecases.ConfiguracionService
	logger *slog.Logger
}

func (b *backupManager) directorioEfectivo(ctx context.Context) (string, error) {
	cfg, err := b.cfgSvc.Obtener(ctx)
	if err != nil {
		return "", err
	}
	dir, err := platform.ResolverDirectorioBackups(filepath.Dir(b.dbPath), cfg.BackupDir)
	if err != nil {
		// El usuario pudo haber configurado una ruta inválida/no escribible
		// (p. ej. una carpeta de nube que ya no existe): se reporta como
		// error de validación de ESA configuración, no como un 500 genérico.
		return "", domain.Invalid("backup_dir", err.Error())
	}
	return dir, nil
}

func (b *backupManager) Crear(ctx context.Context) (platform.BackupInfo, error) {
	dir, err := b.directorioEfectivo(ctx)
	if err != nil {
		return platform.BackupInfo{}, err
	}
	info, err := platform.CrearBackup(ctx, b.db, b.dbPath, dir)
	if err != nil {
		return platform.BackupInfo{}, err
	}
	platform.AplicarRetencion(dir, backupsARetener)
	return info, nil
}

func (b *backupManager) Listar(ctx context.Context) ([]platform.BackupInfo, error) {
	dir, err := b.directorioEfectivo(ctx)
	if err != nil {
		return nil, err
	}
	return platform.ListarBackups(dir)
}

func (b *backupManager) RutaArchivo(ctx context.Context, nombre string) (string, error) {
	dir, err := b.directorioEfectivo(ctx)
	if err != nil {
		return "", err
	}
	ruta, err := platform.RutaBackup(dir, nombre)
	if err != nil {
		return "", domain.NotFoundf("No se encontró el backup '%s'.", nombre)
	}
	return ruta, nil
}

// Restaurar reemplaza la BD de esta instancia por el backup indicado. Antes
// de tocar nada, saca un backup del estado ACTUAL (por si el restore fue un
// error) y luego cierra la conexión compartida: tras esto, la app queda
// inservible hasta que el usuario la reinicie — deliberado, ver
// RestaurarArchivo en internal/platform/backup.go. No se intenta reabrir la
// conexión ni re-cablear los casos de uso dentro del mismo proceso: es un
// cambio de los datos bajo los pies de toda la app, más seguro pedir un
// reinicio limpio que intentar un hot-swap.
func (b *backupManager) Restaurar(ctx context.Context, nombre string) error {
	dir, err := b.directorioEfectivo(ctx)
	if err != nil {
		return err
	}
	backupPath, err := platform.RutaBackup(dir, nombre)
	if err != nil {
		return domain.NotFoundf("No se encontró el backup '%s'.", nombre)
	}

	if _, err := platform.CrearBackup(ctx, b.db, b.dbPath, dir); err != nil {
		return fmt.Errorf("creando backup de seguridad antes de restaurar: %w", err)
	}
	if err := b.db.Close(); err != nil {
		return fmt.Errorf("cerrando la base de datos: %w", err)
	}
	if err := platform.RestaurarArchivo(b.dbPath, backupPath); err != nil {
		return err
	}
	b.logger.Warn("base de datos restaurada desde backup; es necesario reiniciar la aplicación",
		"backup", nombre)
	return nil
}

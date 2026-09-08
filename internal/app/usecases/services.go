package usecases

import (
	"context"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

// Deps agrupa las dependencias comunes a todos los servicios.
type Deps struct {
	Repos ports.Repos      // acceso a repositorios sin transacción (lecturas)
	UOW   ports.UnitOfWork // agrupa escrituras en una transacción
	Clock ports.Clock
	IDs   ports.IDGen
}

// Services reúne todos los servicios de casos de uso. La capa HTTP recibe esto.
type Services struct {
	Productos *ProductoService
	Areas     *AreaService
	Recetas   *RecetaService
	Ventas    *VentaService
	IPV       *IPVService
	Historial *HistorialService
}

// registrarCambio añade una entrada de historial usando la transacción en curso.
// Los servicios lo heredan al incrustar Deps.
func (d Deps) registrarCambio(ctx context.Context, r ports.Repos, tipo domain.TipoEntidad, entidadID, campo, anterior, nuevo string) error {
	h := domain.NuevoCambio(tipo, entidadID, campo, anterior, nuevo)
	h.ID = d.IDs.New()
	h.FechaCambio = d.Clock.Now()
	return r.Historial().Registrar(ctx, h)
}

// New construye todos los servicios a partir de las dependencias.
func New(d Deps) *Services {
	return &Services{
		Productos: &ProductoService{d},
		Areas:     &AreaService{d},
		Recetas:   &RecetaService{d},
		Ventas:    &VentaService{d},
		IPV:       &IPVService{d},
		Historial: &HistorialService{d},
	}
}

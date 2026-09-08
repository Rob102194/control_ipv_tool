package platform

import (
	"time"

	"github.com/google/uuid"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
)

// SystemClock implementa ports.Clock con el reloj real del sistema.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }

func (SystemClock) Today() domain.Date { return domain.DateFromTime(time.Now()) }

// UUIDGen implementa ports.IDGen con UUID v4 en cadena, igual que la versión
// Python (str(uuid.uuid4())).
type UUIDGen struct{}

func (UUIDGen) New() string { return uuid.NewString() }
